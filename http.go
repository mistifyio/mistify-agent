package agent

// the HTTP interface

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	runtime_pprof "runtime/pprof"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bakins/logrus-middleware"
	"github.com/bakins/net-http-recover"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/mistifyio/kvite"
	"github.com/mistifyio/mistify-agent/client"
)

type (
	// HTTPRequest is a container for an http request and its context
	HTTPRequest struct {
		ResponseWriter http.ResponseWriter
		Request        *http.Request
		Context        *Context
		vars           map[string]string
		Guest          *client.Guest
		GuestRunner    *GuestRunner
	}

	// HTTPResponse is a wrapper for http.ResponseWriter which provides access
	// to several convenience methods
	HTTPResponse struct {
		http.ResponseWriter
	}

	// HTTPErrorMessage is an enhanced error struct for http error responses
	HTTPErrorMessage struct {
		Message string   `json:"message"`
		Code    int      `json:"code"`
		Stack   []string `json:"stack"`
	}

	// Chain is a middleware chain
	Chain struct {
		alice.Chain
		ctx *Context
	}
)

const ctxKey string = "agentContext"

var (
	// ErrNotFound is the error for a resouce not found
	ErrNotFound = errors.New("not found")
)

// AttachProfiler enables debug profiling exposed on http api endpoints
func AttachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	for _, profile := range runtime_pprof.Profiles() {
		router.Handle(fmt.Sprintf("/debug/pprof/%s", profile.Name()), pprof.Handler(profile.Name()))
	}
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

// Run prepares and runs the http server
func Run(ctx *Context, address string) error {
	r := mux.NewRouter()
	r.StrictSlash(true)

	AttachProfiler(r)

	logrusMiddleware := logrusmiddleware.Middleware{
		Name: "agent",
	}
	commonMiddleware := alice.New(
		func(h http.Handler) http.Handler {
			return logrusMiddleware.Handler(h, "")
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		},
		func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				context.Set(r, ctxKey, ctx)
				h.ServeHTTP(w, r)
			})
		},
	)

	chain := Chain{
		ctx: ctx,
		Chain: alice.New(
			func(h http.Handler) http.Handler {
				return handlers.CombinedLoggingHandler(os.Stdout, h)
			},
			handlers.CompressHandler,
			func(h http.Handler) http.Handler {
				return recovery.Handler(os.Stderr, h, true)
			}),
	}

	// General
	r.HandleFunc("/metadata", chain.RequestWrapper(getMetadata)).Methods("GET")
	r.HandleFunc("/metadata", chain.RequestWrapper(setMetadata)).Methods("PATCH")

	r.HandleFunc("/images", chain.RequestWrapper(listImages)).Queries("type", "{type:[a-zA-Z]+}").Methods("GET")
	r.HandleFunc("/images", chain.RequestWrapper(listImages)).Methods("GET")
	r.HandleFunc("/images", chain.RequestWrapper(fetchImage)).Methods("POST")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(getImage)).Queries("type", "{type:[a-zA-Z]+}").Methods("GET")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(getImage)).Methods("GET")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(deleteImage)).Queries("type", "{type:[a-zA-Z]+}").Methods("DELETE")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(deleteImage)).Methods("DELETE")

	r.HandleFunc("/jobs", chain.RequestWrapper(getLatestJobs)).Methods("GET").Queries("limit", "{limit:[0-9]+}").Methods("GET")
	r.HandleFunc("/jobs", chain.RequestWrapper(getLatestJobs)).Methods("GET")
	r.HandleFunc("/jobs/{jobID}", chain.RequestWrapper(getJobStatus)).Methods("GET")

	r.HandleFunc("/guests", chain.RequestWrapper(listGuests)).Methods("GET")

	// Specific Guest
	r.HandleFunc("/guests", chain.RequestWrapper(createGuest)).Methods("POST") // Special setup
	r.HandleFunc("/guests/{id}", chain.GuestRunnerWrapper(getGuest)).Methods("GET")
	r.HandleFunc("/guests/{id}/jobs", chain.GuestRunnerWrapper(getLatestGuestJobs)).Queries("limit", "{limit:[0-9]+}").Methods("GET")
	r.HandleFunc("/guests/{id}/jobs", chain.GuestRunnerWrapper(getLatestGuestJobs)).Methods("GET")
	r.HandleFunc("/guests/{id}/jobs/{jobID}", chain.GuestRunnerWrapper(getJobStatus)).Methods("GET")
	//r.HandleFunc("/guests/{id}", chain.GuestRunnerWrapper(deleteGuest)).Methods("DELETE")
	r.HandleFunc("/guests/{id}/metadata", chain.GuestRunnerWrapper(getGuestMetadata)).Methods("GET")
	r.HandleFunc("/guests/{id}/metadata", chain.GuestRunnerWrapper(setGuestMetadata)).Methods("PATCH")

	r.HandleFunc("/guests/{id}/metrics/cpu", chain.GuestRunnerWrapper(getCPUMetrics)).Methods("GET")
	r.HandleFunc("/guests/{id}/metrics/disk", chain.GuestRunnerWrapper(getDiskMetrics)).Methods("GET")
	r.HandleFunc("/guests/{id}/metrics/nic", chain.GuestRunnerWrapper(getNicMetrics)).Methods("GET")

	for _, action := range []string{"shutdown", "reboot", "restart", "poweroff", "start", "suspend", "delete"} {
		r.HandleFunc(fmt.Sprintf("/guests/{id}/%s", action), chain.GuestActionWrapper(action)).Methods("POST")
	}

	for _, prefix := range []string{"/guests/{id}", "/guests/{id}/disks/{disk}"} {
		r.HandleFunc(fmt.Sprintf("%s/snapshots", prefix), chain.GuestRunnerWrapper(listSnapshots)).Methods("GET")
		r.HandleFunc(fmt.Sprintf("%s/snapshots", prefix), chain.GuestRunnerWrapper(createSnapshot)).Methods("POST")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}", prefix), chain.GuestRunnerWrapper(getSnapshot)).Methods("GET")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}", prefix), chain.GuestRunnerWrapper(deleteSnapshot)).Methods("DELETE")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}/rollback", prefix), chain.GuestRunnerWrapper(rollbackSnapshot)).Methods("POST")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}/download", prefix), chain.GuestRunnerWrapper(downloadSnapshot)).Methods("GET")
	}

	/*
		guest := guests.PathPrefix("/{id}").Subrouter()
		guest.HandleFunc("/vnc", RequestWrapper(ctx, vncGuest))
		guest.HandleFunc("/history", RequestWrapper(ctx, getGuestHistory)).Methods("GET")

	*/

	s := &http.Server{
		Addr:           address,
		Handler:        r,
		MaxHeaderBytes: 1 << 20,
	}
	return s.ListenAndServe()
}

// RequestWrapper turns a basic http request into an HTTPRequest
func (c *Chain) RequestWrapper(fn func(*HTTPRequest) *HTTPErrorMessage) http.HandlerFunc {
	return c.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := HTTPRequest{
			Context:        c.ctx,
			ResponseWriter: w,
			Request:        r,
		}
		if err := fn(&req); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"func":  "agent.RequestWrapper",
				"stack": strings.Join(err.Stack, "\t\n\t"),
				"path":  r.URL.Path,
			}).Error(err)
			req.JSON(err.Code, err)
		}
	})).ServeHTTP
}

// Parameter retrieves request parameter
func (r *HTTPRequest) Parameter(key string) string {
	vars := r.vars

	if vars == nil {
		vars = mux.Vars(r.Request)
		r.vars = vars
	}

	if vars == nil {
		return ""
	}

	return vars[key]
}

// SetHeader sets an http response header
func (r *HTTPRequest) SetHeader(key, val string) {
	r.ResponseWriter.Header().Set(key, val)
}

// JSON sends an http response with a json body
func (r *HTTPRequest) JSON(code int, obj interface{}) *HTTPErrorMessage {
	r.SetHeader("Content-Type", "application/json")
	r.ResponseWriter.WriteHeader(code)
	encoder := json.NewEncoder(r.ResponseWriter)
	if err := encoder.Encode(obj); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}
	return nil
}

// NewError creates an HTTPErrorMessage
func (r *HTTPRequest) NewError(err error, code int) *HTTPErrorMessage {
	if code <= 0 {
		code = http.StatusInternalServerError
	}
	msg := HTTPErrorMessage{
		Message: err.Error(),
		Code:    code,
		Stack:   make([]string, 0, 4),
	}
	for i := 1; ; i++ { //
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		msg.Stack = append(msg.Stack, fmt.Sprintf("%s:%d (0x%x)", file, line, pc))
	}
	return &msg
}

func getMetadata(r *HTTPRequest) *HTTPErrorMessage {
	metadata := make(map[string]string)

	err := r.Context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("hypervisor-metadata")
		if err != nil {
			return err
		}
		return b.ForEach(func(k string, v []byte) error {
			metadata[string(k)] = string(v)
			return nil
		})
	})

	if err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	return r.JSON(http.StatusOK, metadata)
}

func setMetadata(r *HTTPRequest) *HTTPErrorMessage {
	var metadata map[string]string

	err := json.NewDecoder(r.Request.Body).Decode(&metadata)
	if err != nil {
		return r.NewError(err, http.StatusBadRequest)
	}

	err = r.Context.db.Transaction(func(tx *kvite.Tx) error {
		for key, value := range metadata {
			b, err := tx.Bucket("hypervisor-metadata")
			if err != nil {
				return err
			}
			if value == "" {
				if err := b.Delete(key); err != nil {
					return err
				}
			} else {
				if err := b.Put(key, []byte(value)); err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	return getMetadata(r)
}

// JSON writes appropriate headers and JSON body to the http response
func (hr *HTTPResponse) JSON(code int, obj interface{}) {
	hr.Header().Set("Content-Type", "application/json")
	hr.WriteHeader(code)
	encoder := json.NewEncoder(hr)
	if err := encoder.Encode(obj); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
	}
}

// JSONError prepares an HTTPError with a stack trace and writes it with
// HTTPResponse.JSON
func (hr *HTTPResponse) JSONError(code int, err error) {
	httpError := &HTTPErrorMessage{
		Message: err.Error(),
		Code:    code,
		Stack:   make([]string, 0, 4),
	}
	for i := 1; ; i++ { //
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		httpError.Stack = append(httpError.Stack, fmt.Sprintf("%s:%d (0x%x)", file, line, pc))
	}
	hr.JSON(code, httpError)
}

// JSONMsg is a convenience method to write a JSON response with just a message
// string
func (hr *HTTPResponse) JSONMsg(code int, msg string) {
	msgObj := map[string]string{
		"message": msg,
	}
	hr.JSON(code, msgObj)
}

// GetContext retrieves a Context value for a request
func GetContext(r *http.Request) *Context {
	if value := context.Get(r, ctxKey); value != nil {
		return value.(*Context)
	}
	return nil
}

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

	"github.com/bakins/logrus-middleware"
	"github.com/bakins/net-http-recover"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/mistifyio/kvite"
)

type (
	// HTTPResponse is a wrapper for http.ResponseWriter which provides access
	// to several convenience methods
	HTTPResponse struct {
		http.ResponseWriter
	}

	// HTTPError is an enhanced error struct for http error responses
	HTTPError struct {
		Message string   `json:"message"`
		Code    int      `json:"code"`
		Stack   []string `json:"stack"`
	}
)

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error code: %d, message: %s", e.Code, e.Message)
}

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

	guestMiddleware := alice.New(
		GetGuestMiddleware,
		guestRunnerMiddleware,
	)

	// General
	r.HandleFunc("/metadata", getMetadata).Methods("GET")
	r.HandleFunc("/metadata", setMetadata).Methods("PATCH")

	r.HandleFunc("/images", listImages).Queries("type", "{type:[a-zA-Z]+}").Methods("GET")
	r.HandleFunc("/images", listImages).Methods("GET")
	r.HandleFunc("/images", fetchImage).Methods("POST")
	r.HandleFunc("/images/{id}", getImage).Queries("type", "{type:[a-zA-Z]+}").Methods("GET")
	r.HandleFunc("/images/{id}", getImage).Methods("GET")
	r.HandleFunc("/images/{id}", deleteImage).Queries("type", "{type:[a-zA-Z]+}").Methods("DELETE")
	r.HandleFunc("/images/{id}", deleteImage).Methods("DELETE")

	r.HandleFunc("/jobs", getLatestJobs).Methods("GET").Queries("limit", "{limit:[0-9]+}").Methods("GET")
	r.HandleFunc("/jobs", getLatestJobs).Methods("GET")
	r.HandleFunc("/jobs/{jobID}", getJobStatus).Methods("GET")

	// Guest Routes
	r.HandleFunc("/guests", listGuests).Methods("GET")
	r.HandleFunc("/guests", createGuest).Methods("POST") // Special setup

	// Since mux requires all routes to start with "/", can't put this bare
	// one in the guest subrouter cleanly
	r.Handle("/guests/{id}", guestMiddleware.ThenFunc(getGuest)).Methods("GET")

	// Specific guest, but don't need the guest middlewares, so register
	// separately from the subrouter
	r.HandleFunc("/guests/{id}/jobs", getLatestGuestJobs).Queries("limit", "{limit:[0-9]+}").Methods("GET")
	r.HandleFunc("/guests/{id}/jobs", getLatestGuestJobs).Methods("GET")
	r.HandleFunc("/guests/{id}/jobs/{jobID}", getJobStatus).Methods("GET")

	// Guest subrouter
	// Since middleware needs to be applied, a basic subrouter can't be used.
	// Instead, create a new router with the prefix and then register that with
	// the main router.
	gr := mux.NewRouter().PathPrefix("/guests/{id}").Subrouter()
	r.Handle("/guests/{id}/{path:.*}", guestMiddleware.Then(gr))

	gr.HandleFunc("/metadata", getGuestMetadata).Methods("GET")
	gr.HandleFunc("/metadata", setGuestMetadata).Methods("PATCH")

	gr.HandleFunc("/metrics/cpu", getCPUMetrics).Methods("GET")
	gr.HandleFunc("/metrics/disk", getDiskMetrics).Methods("GET")
	gr.HandleFunc("/metrics/nic", getNicMetrics).Methods("GET")

	for _, action := range []string{"shutdown", "reboot", "restart", "poweroff", "start", "suspend", "delete"} {
		gr.HandleFunc(fmt.Sprintf("/%s", action), GenerateGuestAction(action)).Methods("POST")
	}

	for _, prefix := range []string{"", "/disks/{disk}"} {
		gr.HandleFunc(fmt.Sprintf("%s/snapshots", prefix), listSnapshots).Methods("GET")
		gr.HandleFunc(fmt.Sprintf("%s/snapshots", prefix), createSnapshot).Methods("POST")
		gr.HandleFunc(fmt.Sprintf("%s/snapshots/{name}", prefix), getSnapshot).Methods("GET")
		gr.HandleFunc(fmt.Sprintf("%s/snapshots/{name}", prefix), deleteSnapshot).Methods("DELETE")
		gr.HandleFunc(fmt.Sprintf("%s/snapshots/{name}/rollback", prefix), rollbackSnapshot).Methods("POST")
		gr.HandleFunc(fmt.Sprintf("%s/snapshots/{name}/download", prefix), downloadSnapshot).Methods("GET")
	}

	s := &http.Server{
		Addr:           address,
		Handler:        commonMiddleware.Then(r),
		MaxHeaderBytes: 1 << 20,
	}
	return s.ListenAndServe()
}

func getMetadata(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ctx := GetContext(r)

	metadata := make(map[string]string)

	err := ctx.db.Transaction(func(tx *kvite.Tx) error {
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
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}

	hr.JSON(http.StatusOK, metadata)
}

func setMetadata(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ctx := GetContext(r)

	var metadata map[string]string

	err := json.NewDecoder(r.Body).Decode(&metadata)
	if err != nil {
		hr.JSONError(http.StatusBadRequest, err)
		return
	}

	err = ctx.db.Transaction(func(tx *kvite.Tx) error {
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
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}

	getMetadata(w, r)
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
	httpError := NewHTTPError(code, err)
	// Remove this function call from the stack
	httpError.Stack = httpError.Stack[1:]

	hr.JSON(code, httpError)
}

// NewHTTPError prepares an HTTPError with a stack trace
func NewHTTPError(code int, err error) *HTTPError {
	httpError := &HTTPError{
		Message: err.Error(),
		Code:    code,
		Stack:   make([]string, 0, 4),
	}
	// Loop through the callers to build the stack. Skip the first one, which
	// is this function and continue until there are no more callers
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Look up the function name (form of package.Name)
		fnName := runtime.FuncForPC(pc).Name()
		// Add the line to the stack array
		httpError.Stack = append(httpError.Stack, fmt.Sprintf("%s:%d (%s)", file, line, fnName))
	}
	return httpError
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

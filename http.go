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
	"time"

	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/mistifyio/kvite"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/log"
)

type (
	HttpRequest struct {
		ResponseWriter http.ResponseWriter
		Request        *http.Request
		Context        *Context
		vars           map[string]string
		Guest          *client.Guest
		GuestRunner    *GuestRunner
	}

	HttpErrorMessage struct {
		Message string   `json:"message"`
		Code    int      `json:"code"`
		Stack   []string `json:"stack"`
	}

	Chain struct {
		alice.Chain
		ctx *Context
	}
)

var (
	NotFound = errors.New("not found")
)

func AttachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	for _, profile := range runtime_pprof.Profiles() {
		router.Handle(fmt.Sprintf("/debug/pprof/%s", profile.Name()), pprof.Handler(profile.Name()))
	}
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

func Run(ctx *Context, address string) error {
	r := mux.NewRouter()
	r.StrictSlash(true)

	AttachProfiler(r)

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

	r.HandleFunc("/images", chain.RequestWrapper(listImages)).Methods("GET")
	r.HandleFunc("/images", chain.RequestWrapper(fetchImage)).Methods("POST")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(getImage)).Methods("GET")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(deleteImage)).Methods("DELETE")

	r.HandleFunc("/guests", chain.RequestWrapper(listGuests)).Methods("GET")

	// Specific Guest
	r.HandleFunc("/guests", chain.RequestWrapper(createGuest)).Methods("POST") // Special setup
	r.HandleFunc("/guests/{id}", chain.GuestRunnerWrapper(getGuest)).Methods("GET")
	//r.HandleFunc("/guests/{id}", chain.GuestRunnerWrapper(deleteGuest)).Methods("DELETE")
	r.HandleFunc("/guests/{id}/metadata", chain.GuestRunnerWrapper(getGuestMetadata)).Methods("GET")
	r.HandleFunc("/guests/{id}/metadata", chain.GuestRunnerWrapper(setGuestMetadata)).Methods("PATCH")

	r.HandleFunc("/guests/{id}/metrics/cpu", chain.GuestRunnerWrapper(getCpuMetrics)).Methods("GET")
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
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s.ListenAndServe()
}

func (c *Chain) RequestWrapper(fn func(*HttpRequest) *HttpErrorMessage) http.HandlerFunc {
	return c.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := HttpRequest{
			Context:        c.ctx,
			ResponseWriter: w,
			Request:        r,
		}
		if err := fn(&req); err != nil {
			log.Error("%s\n\t%s\n", err.Message, strings.Join(err.Stack, "\t\n\t"))
			req.JSON(err.Code, err)
		}
	})).ServeHTTP
}

func (r *HttpRequest) Parameter(key string) string {
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

func (r *HttpRequest) SetHeader(key, val string) {
	r.ResponseWriter.Header().Set(key, val)
}

func (r *HttpRequest) JSON(code int, obj interface{}) *HttpErrorMessage {
	r.SetHeader("Content-Type", "application/json")
	r.ResponseWriter.WriteHeader(code)
	encoder := json.NewEncoder(r.ResponseWriter)
	if err := encoder.Encode(obj); err != nil {
		return r.NewError(err, 500)
	}
	return nil
}

func (r *HttpRequest) NewError(err error, code int) *HttpErrorMessage {
	if code <= 0 {
		code = 500
	}
	msg := HttpErrorMessage{
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

func getMetadata(r *HttpRequest) *HttpErrorMessage {
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
		return r.NewError(err, 500)
	}

	return r.JSON(200, metadata)
}

func setMetadata(r *HttpRequest) *HttpErrorMessage {
	var metadata map[string]string

	err := json.NewDecoder(r.Request.Body).Decode(&metadata)
	if err != nil {
		return r.NewError(err, 400)
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
		return r.NewError(err, 500)
	}

	return getMetadata(r)
}

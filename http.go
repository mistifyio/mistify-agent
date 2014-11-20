package agent

// the HTTP interface

import (
	"encoding/json"
	"errors"
	_ "expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bakins/net-http-recover"
	"github.com/gorilla/context"
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

const (
	GUEST_KEY = 1 << iota
	CONTEXT_KEY
)

var (
	NotFound = errors.New("not found")
)

func Run(ctx *Context, address string) error {
	r := mux.NewRouter()
	r.StrictSlash(true)

	// default mux will have the profiler handlers
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	// TODO: normal chain, no need for a wrapper
	chain := Chain{
		ctx: ctx,
		Chain: alice.New(
			func(h http.Handler) http.Handler {
				return ContextHandler(ctx, h)
			},
			func(h http.Handler) http.Handler {
				return handlers.CombinedLoggingHandler(os.Stdout, h)
			},
			handlers.CompressHandler,
			func(h http.Handler) http.Handler {
				return recovery.Handler(os.Stderr, h, true)
			}),
	}

	guest_chain := chain.Append(GuestHandler)

	r.Handle("/metadata", chain.ThenFunc(getMetadata)).Methods("GET")
	r.Handle("/metadata", chain.ThenFunc(setMetadata)).Methods("PATCH")

	r.Handle("/guests", chain.ThenFunc(listGuests)).Methods("GET")
	r.Handle("/guests", chain.ThenFunc(createGuest)).Methods("POST")

	r.Handle("/guests/{guest_id}", guest_chain.ThenFunc(getGuest)).Methods("GET")
	r.Handle("/guests/{guest_id}", guest_chain.ThenFunc(deleteGuest)).Methods("DELETE")
	r.Handle("/guests/{guest_id}/metadata", guest_chain.ThenFunc(getGuestMetadata)).Methods("GET")
	r.Handle("/guests/{guest_id}/metadata", guest_chain.ThenFunc(setGuestMetadata)).Methods("PATCH")

	r.Handle("/guests/{guest_id}/metrics/cpu", guest_chain.ThenFunc(getGuestCPUMetrics)).Methods("GET")
	r.Handle("/guests/{guest_id}/metrics/disk", guest_chain.ThenFunc(getGuestDiskMetrics)).Methods("GET")
	r.Handle("/guests/{guest_id}/metrics/nic", guest_chain.ThenFunc(getGuestNicMetrics)).Methods("GET")

	for _, action := range []string{"shutdown", "reboot", "restart", "poweroff", "start", "suspend", "delete"} {
		r.Handle(fmt.Sprintf("/guests/{id}/%s", action), guest_chain.Then(actionWrapper(action))).Methods("POST")
	}

	for _, prefix := range []string{"/guests/{id}", "/guests/{id}/disks/{disk}"} {
		r.HandleFunc(fmt.Sprintf("%s/snapshots", prefix), chain.RequestWrapper(listSnapshots)).Methods("GET")
		r.HandleFunc(fmt.Sprintf("%s/snapshots", prefix), chain.RequestWrapper(createSnapshot)).Methods("POST")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}", prefix), chain.RequestWrapper(getSnapshot)).Methods("GET")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}", prefix), chain.RequestWrapper(deleteSnapshot)).Methods("DELETE")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}/rollback", prefix), chain.RequestWrapper(rollbackSnapshot)).Methods("POST")
		r.HandleFunc(fmt.Sprintf("%s/snapshots/{name}/download", prefix), chain.RequestWrapper(downloadSnapshot)).Methods("GET")
	}
	r.HandleFunc("/images", chain.RequestWrapper(listImages)).Methods("GET")
	r.HandleFunc("/images", chain.RequestWrapper(fetchImage)).Methods("POST")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(getImage)).Methods("GET")
	r.HandleFunc("/images/{id}", chain.RequestWrapper(deleteImage)).Methods("DELETE")

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

func getMetadata(w http.ResponseWriter, r *http.Request) {
	//assume we get called after ContextHandler. If not, this will panic
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
		NewError(err, 500).Serve(w)
		return
	}

	JSON(w, 200, metadata)
}

func setMetadata(w http.ResponseWriter, r *http.Request) {
	//assume we get called after ContextHandler. If not, this will panic
	ctx := GetContext(r)

	var metadata map[string]string

	err := json.NewDecoder(r.Body).Decode(&metadata)
	if err != nil {
		NewError(err, 400).Serve(w)
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
		NewError(err, 500).Serve(w)
		return
	}

	getMetadata(w, r)
}

// ContextHandler is an http.handler "middleware" that add the magic context to request
func ContextHandler(ctx *Context, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, CONTEXT_KEY, ctx)
		h.ServeHTTP(w, r)
	})
}

// GetContext retrieves the magic context that has been placed on the request by
//  ContextHandler
func GetContext(r *http.Request) *Context {
	if rv := context.Get(r, CONTEXT_KEY); rv != nil {
		return rv.(*Context)
	}
	return nil
}

// GuestHandler is an http.handler "middleware" that looks up a guest
// and adds to context. The guest id should be stored in the parameter 'guest_id'.
// This should always be added after ContextHandler
func GuestHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := GetContext(r)
		if ctx == nil {
			NewError(errors.New("no context found"), 500).Serve(w)
			return
		}

		vars := mux.Vars(r)
		id := vars["guest_id"]

		if id == "" {
			NewError(errors.New("no guest id in request"), 400).Serve(w)
			return
		}

		var g client.Guest
		err := ctx.db.Transaction(func(tx *kvite.Tx) error {
			b, err := tx.Bucket("guests")
			if err != nil {
				return err
			}
			data, err := b.Get(id)
			if err != nil {
				return err
			}
			if data == nil {
				return NotFound
			}

			return json.Unmarshal(data, &g)
		})

		if err != nil {
			code := 500
			if err == NotFound {
				code = 404
			}
			NewError(err, code).Serve(w)
		}

		context.Set(r, GUEST_KEY, &g)
		h.ServeHTTP(w, r)
	})
}

// GuestContext retrieves the magic context that has been placed on the request by
//  GuestHandler
func GetGuest(r *http.Request) *client.Guest {
	if rv := context.Get(r, GUEST_KEY); rv != nil {
		return rv.(*client.Guest)
	}
	return nil
}

// JSON encodes the objects and sets the appropriate http headers
func JSON(w http.ResponseWriter, code int, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(obj); err != nil {
		// what can we do with error???
	}
}

// NewError is a convinience method for wrapping an http response
func NewError(err error, code int) *HttpErrorMessage {
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

// Serve returns the error message to the client
func (msg *HttpErrorMessage) Serve(w http.ResponseWriter) {
	JSON(w, msg.Code, msg)
}

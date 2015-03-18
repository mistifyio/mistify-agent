package rpc

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/justinas/alice"
)

type (
	// Server is a basic JSON-RPC over HTTP server.
	Server struct {
		RPCServer  *rpc.Server
		HTTPServer *http.Server
		Router     *mux.Router
		Chain      alice.Chain
	}
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

// NewServer creates an JSON-RPC HTTP server bound to 127.0.0.1.  This answers RPC requests on the Mistify RPC Path.
// This server logs to STDOUT and also presents pprof on /debug/pprof/
func NewServer(port uint) (*Server, error) {
	s := &Server{
		RPCServer: rpc.NewServer(),
		HTTPServer: &http.Server{
			Addr: fmt.Sprintf("127.0.0.1:%d", port),
		},
		Router: mux.NewRouter(),
	}
	s.RPCServer.RegisterCodec(NewCodec(), "application/json")
	s.HTTPServer.Handler = s.Router

	s.Chain = alice.New(
		func(h http.Handler) http.Handler {
			return newLogger(h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		})

	s.Router.Handle(RPCPath, s.Chain.Then(s.RPCServer))
	attachProfiler(s.Router)
	return s, nil
}

// RegisterService is a helper for registering a new RPC service
func (s *Server) RegisterService(receiver interface{}) error {
	return s.RPCServer.RegisterService(receiver, "")
}

// Handle is a helper for registering a handler for a given path.
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.Router.Handle(pattern, s.Chain.Then(handler))
}

// HandleFunc is a helper for registering a handler function for a given path
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.Router.Handle(pattern, s.Chain.ThenFunc(handler))
}

// ListenAndServe is a helper for starting the HTTP service. This generally does not return.
func (s *Server) ListenAndServe() error {
	return s.HTTPServer.ListenAndServe()
}

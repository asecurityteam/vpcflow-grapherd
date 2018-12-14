package grapherd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"bitbucket.org/atlassian/vpcflow-grapherd/pkg/types"
	"github.com/go-chi/chi"
)

// Server is an interface for starting/stopping an HTTP server
type Server interface {
	// ListenAndServe starts the HTTP server in a blocking call.
	ListenAndServe() error
	// Shutdown stops the server from accepting new connections.
	// If the given context expires before shutdown is complete then
	// the context error is returned.
	Shutdown(ctx context.Context) error
}

// Service is a container for all of the pluggable modules used by the service
type Service struct {
	// Middleware is a list of service middleware to install on the router.
	// The set of prepackaged middleware can be found in pkg/plugins.
	Middleware []func(http.Handler) http.Handler

	// Queuer is responsible for queuing graphing jobs which will eventually be consumed
	// by the Produce handler. The built in Queuer POSTs to an HTTP endpoint.
	Queuer types.Queuer

	// Storage provides a mechanism to hook into a persistent store for the digests. The
	// built in Storage uses S3 as the persistent storage for digest blobs.
	Storage types.Storage

	// Marker is responsible for marking which digests jobs are inprogress. The built in
	// Marker uses S3 to hold this state.
	Marker types.Marker
}

func (s *Service) init() error {
	return nil
}

// BindRoutes binds the service handlers to the provided router
func (s *Service) BindRoutes(router chi.Router) error {
	if err := s.init(); err != nil {
		return err
	}
	router.Use(s.Middleware...)
	return nil
}

// Runtime is the app configuration and execution point
type Runtime struct {
	Server      Server
	ExitSignals []types.ExitSignal
}

// Run runs the application
func (r *Runtime) Run() error {
	exit := make(chan error)

	for _, f := range r.ExitSignals {
		go func(f func() chan error) {
			exit <- <-f()
		}(f)
	}

	go func() {
		exit <- r.Server.ListenAndServe()
	}()

	err := <-exit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = r.Server.Shutdown(ctx)

	return err
}

// nolint
func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return val
}

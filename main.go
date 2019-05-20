package main

import (
	"context"
	"os"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings"
	grapherd "github.com/asecurityteam/vpcflow-grapherd/pkg"
	"github.com/go-chi/chi"
)

func main() {
	router := chi.NewRouter()
	service := &grapherd.Service{}
	if err := service.BindRoutes(router); err != nil {
		panic(err.Error())
	}

	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}

	// Load the runtime using the Source and Handler.
	rt, err := runhttp.New(context.Background(), source, router)
	if err != nil {
		panic(err.Error())
	}

	// Run the HTTP server.
	if err := rt.Run(); err != nil {
		panic(err.Error())
	}
}

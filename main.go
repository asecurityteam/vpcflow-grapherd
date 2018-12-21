package main

import (
	"net/http"
	"os"

	"bitbucket.org/atlassian/vpcflow-grapherd/pkg"
	"bitbucket.org/atlassian/vpcflow-grapherd/pkg/plugins"
	"bitbucket.org/atlassian/vpcflow-grapherd/pkg/types"
	"github.com/go-chi/chi"
)

func main() {
	router := chi.NewRouter()
	middleware := []func(http.Handler) http.Handler{
		plugins.DefaultLogMiddleware(),
		plugins.DefaultStatMiddleware(),
	}
	service := &grapherd.Service{
		Middleware: middleware,
	}
	if err := service.BindRoutes(router); err != nil {
		panic(err.Error())
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	r := &grapherd.Runtime{
		Server: server,
		ExitSignals: []types.ExitSignal{
			plugins.OS,
		},
	}

	if err := r.Run(); err != nil {
		panic(err.Error())
	}
}

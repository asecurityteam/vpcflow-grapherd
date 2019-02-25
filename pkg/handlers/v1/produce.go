package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/asecurityteam/vpcflow-grapherd/pkg/logs"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/types"
)

type payload struct {
	ID    string `json:"id"`
	Start string `json:"start"`
	Stop  string `json:"stop"`
}

// Produce is a handler which performs the digest job, and stores the digest
type Produce struct {
	LogProvider  types.LoggerProvider
	StatProvider types.StatsProvider
	Marker       types.Marker
	Digester     types.Digester
	Grapher      types.Grapher
}

// ServeHTTP handles incoming HTTP requests, and creates a vpc flow digest
func (h *Produce) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.LogProvider(r.Context())
	var body payload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Info(logs.InvalidInput{Reason: err.Error()})
		writeTextResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if body.ID == "" {
		msg := "missing ID field"
		logger.Info(logs.InvalidInput{Reason: msg})
		writeTextResponse(w, http.StatusBadRequest, msg)
		return
	}

	start, err := time.Parse(time.RFC3339Nano, body.Start)
	if err != nil {
		logger.Info(logs.InvalidInput{Reason: err.Error()})
		writeTextResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	stop, err := time.Parse(time.RFC3339Nano, body.Stop)
	if err != nil {
		logger.Info(logs.InvalidInput{Reason: err.Error()})
		writeTextResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if !stop.After(start) {
		msg := "invalid time range"
		logger.Info(logs.InvalidInput{Reason: msg})
		writeTextResponse(w, http.StatusBadRequest, msg)
		return
	}

	digest, err := h.Digester.Digest(r.Context(), start, stop)
	if err != nil {
		logger.Error(logs.DependencyFailure{Dependency: logs.DependencyDigester, Reason: err.Error()})
		writeTextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer digest.Close()

	if err := h.Grapher.Graph(r.Context(), body.ID, digest); err != nil {
		logger.Error(logs.DependencyFailure{Dependency: logs.DependencyGrapher, Reason: err.Error()})
		writeTextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// We may want to improve this in the future to be a non-fatal error. Today if unmark fails,
	// fetching the digest will result in a perpetual "in progress" state. To mitigate this, we
	// report a failure to the caller signifying that the operation should be retried. This will
	// hopefully mitigate the amount of invalid state occurrence we may incur
	if err := h.Marker.Unmark(r.Context(), body.ID); err != nil {
		logger.Error(logs.DependencyFailure{Dependency: logs.DependencyMarker, Reason: err.Error()})
		writeTextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeTextResponse(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write([]byte(msg))
}

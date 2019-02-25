package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asecurityteam/vpcflow-grapherd/pkg/logs"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/types"
	"github.com/google/uuid"
)

var graphNamespace = uuid.NewSHA1(uuid.Nil, []byte("graph"))

// GrapherHandler handles incoming HTTP requests for creating and retrieving new graphs
type GrapherHandler struct {
	LogProvider  types.LoggerProvider
	StatProvider types.StatsProvider
	Storage      types.Storage
	Marker       types.Marker
	Queuer       types.Queuer
}

// Post creates a new graph
func (h *GrapherHandler) Post(w http.ResponseWriter, r *http.Request) {
	logger := h.LogProvider(r.Context())
	start, stop, err := extractInput(r)
	if err != nil {
		logger.Info(logs.InvalidInput{Reason: err.Error()})
		writeJSONResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	id := computeID(start, stop)
	exists, err := h.Storage.Exists(r.Context(), id)
	switch err.(type) {
	case nil:
	case types.ErrInProgress:
		logger.Info(logs.Conflict{Reason: err.Error()})
		writeJSONResponse(w, http.StatusConflict, err.Error())
		return
	default:
		logger.Error(logs.DependencyFailure{Dependency: logs.DependencyStorage, Reason: err.Error()})
		writeJSONResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	// if data is returned, a graph already exists. return 409 and exit
	if exists {
		startStr := start.Format(time.RFC3339)
		stopStr := start.Format(time.RFC3339)
		msg := fmt.Sprintf("graph for the time %s to %s already exists", startStr, stopStr)
		logger.Info(logs.Conflict{Reason: msg})
		writeJSONResponse(w, http.StatusConflict, msg)
		return
	}

	if err = h.Queuer.Queue(r.Context(), id, start, stop); err != nil {
		logger.Error(logs.DependencyFailure{Dependency: logs.DependencyQueuer, Reason: err.Error()})
		writeJSONResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = h.Marker.Mark(r.Context(), id)
	if err != nil {
		logger.Error(logs.DependencyFailure{Dependency: logs.DependencyMarker, Reason: err.Error()})
	}

	w.WriteHeader(http.StatusAccepted)
}

// Get retrieves a graph
func (h *GrapherHandler) Get(w http.ResponseWriter, r *http.Request) {
	logger := h.LogProvider(r.Context())
	start, stop, err := extractInput(r)
	if err != nil {
		logger.Info(logs.InvalidInput{Reason: err.Error()})
		writeJSONResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	id := computeID(start, stop)
	body, err := h.Storage.Get(r.Context(), id)
	switch err.(type) {
	case nil:
		defer body.Close()
	case types.ErrInProgress:
		w.WriteHeader(http.StatusNoContent)
		return
	case types.ErrNotFound:
		logger.Info(logs.NotFound{Reason: err.Error()})
		w.WriteHeader(http.StatusNotFound)
		return
	default:
		logger.Error(logs.DependencyFailure{Dependency: logs.DependencyStorage, Reason: err.Error()})
		writeJSONResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, body)
}

// extractInput attempts to extract the start/stop query parameters required by GET and POST.
// If either value is not a valid RFC3339Nano or the input is invalid, an error is returned.
// Otherwise, start and stop times are returned in the respective order. Additionally, it
// truncates the time values to the nearest minute since anything with more precision doesn't
// really fit the vpc flow filter use case
func extractInput(r *http.Request) (time.Time, time.Time, error) {
	startString := r.URL.Query().Get("start")
	stopString := r.URL.Query().Get("stop")
	start, err := time.Parse(time.RFC3339Nano, startString)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	stop, err := time.Parse(time.RFC3339Nano, stopString)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if start.After(stop) {
		return time.Time{}, time.Time{}, errors.New("start should be before stop")
	}
	return start.Truncate(time.Minute), stop.Truncate(time.Minute), nil
}

// computeID generates a UUID v5 from a name composed by appending start and stop time strings
// in that order
func computeID(start, stop time.Time) string {
	name := start.String() + stop.String()
	u := uuid.NewSHA1(graphNamespace, []byte(name))
	return u.String()
}

// write the http response with the given status code and message
func writeJSONResponse(w http.ResponseWriter, statusCode int, message string) {
	msg := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(msg)
}

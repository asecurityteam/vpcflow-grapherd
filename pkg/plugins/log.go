package plugins

import (
	"io/ioutil"
	"net/http"

	"bitbucket.org/atlassian/logevent"
	hlog "bitbucket.org/atlassian/logevent/http"
)

// DefaultLogMiddleware injects the default logger on each incoming request's context
func DefaultLogMiddleware() func(http.Handler) http.Handler {
	return CustomLogMiddleware(logevent.New(logevent.Config{}))
}

// NopLogMiddleware injects a nop logger in each incoming request's context
func NopLogMiddleware() func(http.Handler) http.Handler {
	return CustomLogMiddleware(logevent.New(logevent.Config{Output: ioutil.Discard}))
}

// CustomLogMiddleware injects the provided logger in each incoming request's context
func CustomLogMiddleware(logger logevent.Logger) func(http.Handler) http.Handler {
	return hlog.NewMiddleware(logger)
}

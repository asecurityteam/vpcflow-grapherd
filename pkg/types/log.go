package types

import (
	"github.com/asecurityteam/runhttp"
)

// Logger is the project logger interface.
type Logger = runhttp.Logger

// LogFn is the recommended way to extract a logger from the context.
type LogFn = runhttp.LogFn

// LoggerFromContext is a concrete implementation of the LogFn interface.
var LoggerFromContext LogFn = runhttp.LoggerFromContext

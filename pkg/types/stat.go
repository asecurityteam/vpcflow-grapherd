package types

import (
	"github.com/asecurityteam/runhttp"
)

// Stat is the project metrics client interface.
type Stat = runhttp.Stat

// StatFn is the recommended way to extract a metrics client from the context.
type StatFn = runhttp.StatFn

// StatFromContext is a concrete implementation of the StatFn interface.
var StatFromContext StatFn = runhttp.StatFromContext

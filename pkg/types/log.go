package types

import (
	"context"

	"github.com/asecurityteam/logevent"
)

// LoggerProvider extracts a logger from context.
type LoggerProvider func(ctx context.Context) logevent.Logger

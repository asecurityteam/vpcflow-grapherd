package types

import (
	"context"

	"bitbucket.org/atlassian/logevent"
)

// LoggerProvider extracts a logger from context.
type LoggerProvider func(ctx context.Context) logevent.Logger

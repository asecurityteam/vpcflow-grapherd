package types

import (
	"context"
	"io"
	"time"
)

// Digester provides an interface for creating a digest of VPC flow logs for a given start and end time
type Digester interface {
	Digest(context.Context, time.Time, time.Time) (io.ReadCloser, error)
}

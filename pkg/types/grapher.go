package types

import (
	"context"
	"io"
)

// Grapher provides an interface for creating graphs for a provided digest
type Grapher interface {
	Graph(context.Context, string, io.ReadCloser) error
}

package types

import (
	"context"
	"fmt"
)

// ErrInProgress indicates that a digest is in the process of being created
type ErrInProgress struct {
	Key string
}

func (e ErrInProgress) Error() string {
	return fmt.Sprintf("digest %s is being created", e.Key)
}

// Marker is an interface for indicating that a digest is in progress of being created
type Marker interface {
	// Mark flags the digest identified by key as being "in progress"
	Mark(ctx context.Context, key string) error

	// Unmark flags the digest identified by key as not being "in progress"
	Unmark(ctx context.Context, key string) error
}

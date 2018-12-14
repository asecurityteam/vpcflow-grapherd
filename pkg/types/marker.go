package types

import (
	"context"
	"fmt"
	"io"
)

// ErrInProgress indicates that a digest is in the process of being created
type ErrInProgress struct {
	Key string
}

func (e ErrInProgress) Error() string {
	return fmt.Sprintf("digest %s is being created", e.Key)
}

// ErrNotFound represents a resource lookup that failed due to a missing record.
type ErrNotFound struct {
	ID string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("digest %s was not found", e.ID)
}

// Storage is an interface for accessing created digests. It is the caller's responsibility to call Close on the Reader when done.
type Storage interface {
	// Get returns the digest for the given key.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Exists returns true if the digest exists, but does not download the digest body.
	Exists(ctx context.Context, key string) (bool, error)

	// Store stores the digest
	Store(ctx context.Context, key string, data io.ReadCloser) error
}

// Marker is an interface for indicating that a digest is in progress of being created
type Marker interface {
	// Mark flags the digest identified by key as being "in progress"
	Mark(ctx context.Context, key string) error

	// Unmark flags the digest identified by key as not being "in progress"
	Unmark(ctx context.Context, key string) error
}

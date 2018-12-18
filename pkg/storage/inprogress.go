package storage

import (
	"context"
	"io"

	"bitbucket.org/atlassian/vpcflow-grapherd/pkg/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const inProgressSuffix = "_in_progress"

// InProgress is an implementation of Storage which is intended to decorate the S3 implementation.
//
// The decorator will check if a graph is in progress, and if so, will return types.ErrInProgress.
// On a successful Store operation, the decorator will remove the graph's "in progress" status.
type InProgress struct {
	Bucket string
	Client s3iface.S3API
	types.Storage
}

// Get returns the graph for the given key.
//
// If the graph is in the process of being created, an error will be returned of type types.ErrInProgress
func (s *InProgress) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	err := s.checkInProgress(ctx, key)
	// graph is in progress
	if err == nil {
		return nil, types.ErrInProgress{Key: key}
	}
	// unknown error
	if _, ok := parseNotFound(err, key).(types.ErrNotFound); !ok {
		return nil, err
	}
	return s.Storage.Get(ctx, key)
}

// Exists returns true if the graph exists, but does not download the graph body.
//
// If the graph is in the process of being created, an error will be returned of type types.ErrInProgress
func (s *InProgress) Exists(ctx context.Context, key string) (bool, error) {
	err := s.checkInProgress(ctx, key)
	// graph is in progress
	if err == nil {
		return false, types.ErrInProgress{Key: key}
	}
	// unknown error
	if _, ok := parseNotFound(err, key).(types.ErrNotFound); !ok {
		return false, err
	}
	return s.Storage.Exists(ctx, key)
}

func (s *InProgress) checkInProgress(ctx context.Context, key string) error {
	_, err := s.Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key + inProgressSuffix),
	})
	return err
}

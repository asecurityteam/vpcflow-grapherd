package storage

import (
	"context"
	"io"
	"io/ioutil"
	"time"

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
	Bucket  string
	Timeout time.Duration
	Client  s3iface.S3API
	types.Storage
}

// Get returns the graph for the given key.
//
// If the graph is in the process of being created, an error will be returned of type types.ErrInProgress
func (s *InProgress) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	inProgress, err := s.isInProgress(ctx, key)
	if err != nil {
		return nil, err
	}
	if inProgress {
		return nil, types.ErrInProgress{Key: key}
	}
	return s.Storage.Get(ctx, key)
}

// Exists returns true if the graph exists, but does not download the graph body.
//
// If the graph is in the process of being created, an error will be returned of type types.ErrInProgress
func (s *InProgress) Exists(ctx context.Context, key string) (bool, error) {
	inProgress, err := s.isInProgress(ctx, key)
	if err != nil {
		return false, err
	}
	if inProgress {
		return false, types.ErrInProgress{Key: key}
	}
	return s.Storage.Exists(ctx, key)
}

func (s *InProgress) isInProgress(ctx context.Context, key string) (bool, error) {
	res, err := s.Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key + inProgressSuffix),
	})
	if err != nil && isNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	ts, _ := time.Parse(time.RFC3339Nano, string(b))
	now := time.Now()
	return now.Before(ts.Add(s.Timeout)), nil
}

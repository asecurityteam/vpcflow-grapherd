package storage

import (
	"context"
	"io"
	"sync"

	"github.com/asecurityteam/vpcflow-grapherd/pkg/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

const keySuffix = ".dot"

// S3 implements the Storage interface and uses S3 as the backing store for graph
type S3 struct {
	Bucket   string
	Client   s3iface.S3API
	uploader s3manageriface.UploaderAPI
	once     sync.Once
}

// Get returns the graph for the given key. The graph is returned as a gzipped payload.
// It is the caller's responsibility to call Close on the Reader when done.
func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key + keySuffix),
	}
	res, err := s.Client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, parseNotFound(err, key)
	}
	return res.Body, nil
}

// Exists returns true if the graph exists, but does not download the graph body.
func (s *S3) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key + keySuffix),
	}
	_, err := s.Client.HeadObjectWithContext(ctx, input)
	if err == nil {
		return true, nil
	}
	if _, ok := parseNotFound(err, key).(types.ErrNotFound); ok {
		return false, nil
	}
	return false, err
}

// Store stores the graph. It is the caller's responsibility to call Close on the Reader when done.
func (s *S3) Store(ctx context.Context, key string, data io.ReadCloser) error {
	// lazily initialize uploader with the s3 client
	s.once.Do(s.initUploader)

	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key + keySuffix),
		Body:   data,
	})
	return err
}

func (s *S3) initUploader() {
	if s.uploader == nil {
		s.uploader = s3manager.NewUploaderWithClient(s.Client)
	}
}

func isNotFound(err error) bool {
	aErr, ok := err.(awserr.Error)
	return ok && (aErr.Code() == s3.ErrCodeNoSuchKey || aErr.Code() == "NotFound") // NotFound is an undocumented error code with no provided constant
}

// If a key is not found, transform to our NotFound error, otherwise return original error
func parseNotFound(err error, key string) error {
	if isNotFound(err) {
		return types.ErrNotFound{ID: key}
	}
	return err
}

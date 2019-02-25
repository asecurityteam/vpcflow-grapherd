package storage

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/asecurityteam/vpcflow-grapherd/pkg/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	key    = "foo_key"
	bucket = "foo_bucket"
)

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + ".dot"),
		Bucket: aws.String(bucket),
	}
	expectedBody := []byte("some grapg content")
	output := &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewReader(expectedBody)),
	}

	mockS3 := NewMockS3API(ctrl)
	mockS3.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(output, nil)

	storage := &S3{
		Bucket: bucket,
		Client: mockS3,
	}

	r, err := storage.Get(context.Background(), key)
	assert.Nil(t, err)
	defer r.Close()
	data, _ := ioutil.ReadAll(r)
	assert.Equal(t, string(expectedBody), string(data))
}

func TestGetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + ".dot"),
		Bucket: aws.String(bucket),
	}

	aErr := awserr.New(s3.ErrCodeNoSuchKey, "", errors.New(""))

	mockS3 := NewMockS3API(ctrl)
	mockS3.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(nil, aErr)

	storage := &S3{
		Bucket: bucket,
		Client: mockS3,
	}

	_, err := storage.Get(context.Background(), key)
	assert.NotNil(t, err)

	_, ok := err.(types.ErrNotFound)
	assert.True(t, ok)
}

func TestGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + ".dot"),
		Bucket: aws.String(bucket),
	}

	mockS3 := NewMockS3API(ctrl)
	mockS3.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(nil, errors.New("oops"))

	storage := &S3{
		Bucket: bucket,
		Client: mockS3,
	}

	_, err := storage.Get(context.Background(), key)
	assert.NotNil(t, err)
}

func TestExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + ".dot"),
		Bucket: aws.String(bucket),
	}
	output := &s3.HeadObjectOutput{}

	mockS3 := NewMockS3API(ctrl)
	mockS3.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(output, nil)

	storage := &S3{
		Bucket: bucket,
		Client: mockS3,
	}

	exists, err := storage.Exists(context.Background(), key)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + ".dot"),
		Bucket: aws.String(bucket),
	}

	aErr := awserr.New("NotFound", "", errors.New(""))

	mockS3 := NewMockS3API(ctrl)
	mockS3.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, aErr)

	storage := &S3{
		Bucket: bucket,
		Client: mockS3,
	}

	exists, err := storage.Exists(context.Background(), key)
	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + ".dot"),
		Bucket: aws.String(bucket),
	}

	mockS3 := NewMockS3API(ctrl)
	mockS3.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, errors.New("oops"))

	storage := &S3{
		Bucket: bucket,
		Client: mockS3,
	}

	_, err := storage.Exists(context.Background(), key)
	assert.NotNil(t, err)
}

func TestStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	value := "this is a graph"

	mockUploader := NewMockUploaderAPI(ctrl)
	mockUploader.EXPECT().UploadWithContext(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, input *s3manager.UploadInput) (interface{}, error) {
		data, err := ioutil.ReadAll(input.Body)
		if err != nil {
			return nil, err
		}
		assert.Equal(t, value, string(data))
		return &s3manager.UploadOutput{}, nil
	})

	storage := &S3{
		Bucket:   bucket,
		uploader: mockUploader,
	}

	storage.once.Do(func() {}) // trigger the once call

	input := ioutil.NopCloser(bytes.NewReader([]byte(value)))
	err := storage.Store(context.Background(), key, input)
	assert.Nil(t, err)
}

func TestStoreError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	value := "this is another graph"

	mockUploader := NewMockUploaderAPI(ctrl)
	mockUploader.EXPECT().UploadWithContext(gomock.Any(), gomock.Any()).Return(nil, errors.New("oops"))

	storage := &S3{
		Bucket:   bucket,
		uploader: mockUploader,
	}

	storage.once.Do(func() {}) // trigger the once call

	input := ioutil.NopCloser(bytes.NewReader([]byte(value)))
	err := storage.Store(context.Background(), key, input)
	assert.NotNil(t, err)
}

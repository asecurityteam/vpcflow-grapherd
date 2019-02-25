package storage

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/asecurityteam/vpcflow-grapherd/pkg/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetNotInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	output := []byte("graph")
	aErr := awserr.New(s3.ErrCodeNoSuchKey, "", errors.New(""))

	mockClient := NewMockS3API(ctrl)
	mockStorage := NewMockStorage(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(nil, aErr)
	mockStorage.EXPECT().Get(gomock.Any(), key).Return(ioutil.NopCloser(bytes.NewReader(output)), nil)

	ip := &InProgress{
		Bucket:  bucket,
		Client:  mockClient,
		Storage: mockStorage,
	}
	res, err := ip.Get(context.Background(), key)
	assert.Nil(t, err)
	defer res.Close()
	data, _ := ioutil.ReadAll(res)
	assert.Equal(t, string(output), string(data))
}

func TestGetInProgressUnknownError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(nil, errors.New("oops"))

	ip := &InProgress{
		Bucket: bucket,
		Client: mockClient,
	}
	_, err := ip.Get(context.Background(), key)
	assert.NotNil(t, err)
}

func TestGetInProgressBeforeTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}
	getOutput := &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewBufferString(time.Now().Format(time.RFC3339))),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(getOutput, nil)

	ip := &InProgress{
		Timeout: time.Hour,
		Bucket:  bucket,
		Client:  mockClient,
	}
	_, err := ip.Get(context.Background(), key)
	assert.NotNil(t, err)
	_, ok := err.(types.ErrInProgress)
	assert.True(t, ok)
}

func TestGetInProgressAfterTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}
	getOutput := &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewBufferString(time.Now().Add(-2 * time.Hour).Format(time.RFC3339))),
	}
	output := []byte("graph")

	mockClient := NewMockS3API(ctrl)
	mockStorage := NewMockStorage(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(getOutput, nil)
	mockStorage.EXPECT().Get(gomock.Any(), key).Return(ioutil.NopCloser(bytes.NewReader(output)), nil)

	ip := &InProgress{
		Timeout: time.Hour,
		Bucket:  bucket,
		Client:  mockClient,
		Storage: mockStorage,
	}
	res, err := ip.Get(context.Background(), key)
	assert.Nil(t, err)
	defer res.Close()
	data, _ := ioutil.ReadAll(res)
	assert.Equal(t, string(output), string(data))
}

func TestExistsNotInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	aErr := awserr.New(s3.ErrCodeNoSuchKey, "", errors.New(""))

	mockClient := NewMockS3API(ctrl)
	mockStorage := NewMockStorage(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(nil, aErr)
	mockStorage.EXPECT().Exists(gomock.Any(), key).Return(true, nil)

	ip := &InProgress{
		Bucket:  bucket,
		Client:  mockClient,
		Storage: mockStorage,
	}
	exists, err := ip.Exists(context.Background(), key)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestExistsInProgressUnknownError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(nil, errors.New("oops"))

	ip := &InProgress{
		Bucket: bucket,
		Client: mockClient,
	}
	_, err := ip.Exists(context.Background(), key)
	assert.NotNil(t, err)
}

func TestExistsInProgressBeforeTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}
	getOutput := &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewBufferString(time.Now().Format(time.RFC3339))),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(getOutput, nil)

	ip := &InProgress{
		Timeout: time.Hour,
		Bucket:  bucket,
		Client:  mockClient,
	}
	_, err := ip.Exists(context.Background(), key)
	assert.NotNil(t, err)
	_, ok := err.(types.ErrInProgress)
	assert.True(t, ok)
}

func TestExistsInProgressAfterTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.GetObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}
	getOutput := &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewBufferString(time.Now().Add(-2 * time.Hour).Format(time.RFC3339))),
	}

	mockClient := NewMockS3API(ctrl)
	mockStorage := NewMockStorage(ctrl)
	mockClient.EXPECT().GetObjectWithContext(gomock.Any(), expectedInput).Return(getOutput, nil)
	mockStorage.EXPECT().Exists(gomock.Any(), key).Return(true, nil)

	ip := &InProgress{
		Timeout: time.Hour,
		Bucket:  bucket,
		Client:  mockClient,
		Storage: mockStorage,
	}
	exists, err := ip.Exists(context.Background(), key)
	assert.Nil(t, err)
	assert.True(t, exists)
}

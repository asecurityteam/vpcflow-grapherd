package storage

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"bitbucket.org/atlassian/vpcflow-grapherd/pkg/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetNotInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	output := []byte("graph")
	aErr := awserr.New(s3.ErrCodeNoSuchKey, "", errors.New(""))

	mockClient := NewMockS3API(ctrl)
	mockStorage := NewMockStorage(ctrl)
	mockClient.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, aErr)
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

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, errors.New("oops"))

	ip := &InProgress{
		Bucket: bucket,
		Client: mockClient,
	}
	_, err := ip.Get(context.Background(), key)
	assert.NotNil(t, err)
}

func TestGetInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, nil)

	ip := &InProgress{
		Bucket: bucket,
		Client: mockClient,
	}
	_, err := ip.Get(context.Background(), key)
	assert.NotNil(t, err)
	_, ok := err.(types.ErrInProgress)
	assert.True(t, ok)
}

func TestExistsNotInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	aErr := awserr.New(s3.ErrCodeNoSuchKey, "", errors.New(""))

	mockClient := NewMockS3API(ctrl)
	mockStorage := NewMockStorage(ctrl)
	mockClient.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, aErr)
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

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, errors.New("oops"))

	ip := &InProgress{
		Bucket: bucket,
		Client: mockClient,
	}
	_, err := ip.Exists(context.Background(), key)
	assert.NotNil(t, err)
}

func TestExistsInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedInput := &s3.HeadObjectInput{
		Key:    aws.String(key + "_in_progress"),
		Bucket: aws.String(bucket),
	}

	mockClient := NewMockS3API(ctrl)
	mockClient.EXPECT().HeadObjectWithContext(gomock.Any(), expectedInput).Return(nil, nil)

	ip := &InProgress{
		Bucket: bucket,
		Client: mockClient,
	}
	_, err := ip.Exists(context.Background(), key)
	assert.NotNil(t, err)
	_, ok := err.(types.ErrInProgress)
	assert.True(t, ok)
}

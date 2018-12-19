package queuer

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	endpoint = "http://some.host"
)

func TestGraphQueuer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}, nil)

	endpoint, _ := url.Parse(endpoint)
	client := &http.Client{Transport: mockRT}
	dq := GraphQueuer{
		Client:   client,
		Endpoint: endpoint,
	}
	err := dq.Queue(context.Background(), "graphID", time.Now(), time.Now())
	assert.Nil(t, err)
}

func TestUnexpectedResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{StatusCode: 500, Body: ioutil.NopCloser(nil)}, nil)

	endpoint, _ := url.Parse(endpoint)
	client := &http.Client{Transport: mockRT}
	dq := GraphQueuer{
		Client:   client,
		Endpoint: endpoint,
	}
	err := dq.Queue(context.Background(), "graphID", time.Now(), time.Now())
	assert.NotNil(t, err)
}

func TestRTError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, errors.New(""))

	endpoint, _ := url.Parse(endpoint)
	client := &http.Client{Transport: mockRT}
	dq := GraphQueuer{
		Client:   client,
		Endpoint: endpoint,
	}
	err := dq.Queue(context.Background(), "graphID", time.Now(), time.Now())
	assert.NotNil(t, err)
}

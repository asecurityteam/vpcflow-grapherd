package v1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/asecurityteam/logevent"
	"github.com/golang/mock/gomock"
	"github.com/rs/xstats"
	"github.com/stretchr/testify/assert"
)

const (
	payloadTpl = `{"id":"%s","start":"%s","stop":"%s"}`
	key        = "foo_key"
)

func TestProduceBadRequst(t *testing.T) {
	tc := []struct {
		Name    string
		Payload string
	}{
		{
			Name:    "not_json",
			Payload: "not json",
		},
		{
			Name:    "missing_id",
			Payload: fmt.Sprintf(payloadTpl, "", time.Now().Format(time.RFC3339Nano), time.Now().Format(time.RFC3339Nano)),
		},
		{
			Name:    "invalid_start",
			Payload: fmt.Sprintf(payloadTpl, key, "", time.Now().Format(time.RFC3339Nano)),
		},
		{
			Name:    "invalid_stop",
			Payload: fmt.Sprintf(payloadTpl, key, time.Now().Format(time.RFC3339Nano), ""),
		},
		{
			Name:    "invalid_range",
			Payload: fmt.Sprintf(payloadTpl, key, time.Now().Format(time.RFC3339Nano), time.Now().Add(-1*time.Minute).Format(time.RFC3339Nano)),
		},
	}

	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", ioutil.NopCloser(bytes.NewReader([]byte(tt.Payload))))
			r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))
			w := httptest.NewRecorder()
			handler := &Produce{
				LogProvider:  logevent.FromContext,
				StatProvider: xstats.FromContext,
			}
			handler.ServeHTTP(w, r)
			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		})
	}
}

func TestDigestError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	digesterMock := NewMockDigester(ctrl)
	digesterMock.EXPECT().Digest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("oops"))

	start := time.Now().Add(-1 * time.Minute)
	stop := time.Now()
	payload := []byte(fmt.Sprintf(payloadTpl, key, start.Format(time.RFC3339Nano), stop.Format(time.RFC3339Nano)))
	r, _ := http.NewRequest(http.MethodPost, "/", ioutil.NopCloser(bytes.NewReader(payload)))
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))
	w := httptest.NewRecorder()
	handler := &Produce{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Digester:     digesterMock,
	}
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestGrapherError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	digesterMock := NewMockDigester(ctrl)
	digesterMock.EXPECT().Digest(gomock.Any(), gomock.Any(), gomock.Any()).Return(ioutil.NopCloser(bytes.NewReader([]byte(""))), nil)

	grapherMock := NewMockGrapher(ctrl)
	grapherMock.EXPECT().Graph(gomock.Any(), key, gomock.Any()).Return(errors.New("oops"))

	start := time.Now().Add(-1 * time.Minute)
	stop := time.Now()
	payload := []byte(fmt.Sprintf(payloadTpl, key, start.Format(time.RFC3339Nano), stop.Format(time.RFC3339Nano)))
	r, _ := http.NewRequest(http.MethodPost, "/", ioutil.NopCloser(bytes.NewReader(payload)))
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))
	w := httptest.NewRecorder()
	handler := &Produce{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Grapher:      grapherMock,
		Digester:     digesterMock,
	}
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestMarkerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	digesterMock := NewMockDigester(ctrl)
	digesterMock.EXPECT().Digest(gomock.Any(), gomock.Any(), gomock.Any()).Return(ioutil.NopCloser(bytes.NewReader([]byte(""))), nil)

	grapherMock := NewMockGrapher(ctrl)
	grapherMock.EXPECT().Graph(gomock.Any(), key, gomock.Any()).Return(nil)

	markerMock := NewMockMarker(ctrl)
	markerMock.EXPECT().Unmark(gomock.Any(), key).Return(errors.New("oops"))

	start := time.Now().Add(-1 * time.Minute)
	stop := time.Now()
	payload := []byte(fmt.Sprintf(payloadTpl, key, start.Format(time.RFC3339Nano), stop.Format(time.RFC3339Nano)))
	r, _ := http.NewRequest(http.MethodPost, "/", ioutil.NopCloser(bytes.NewReader(payload)))
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))
	w := httptest.NewRecorder()
	handler := &Produce{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Grapher:      grapherMock,
		Marker:       markerMock,
		Digester:     digesterMock,
	}
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	digesterMock := NewMockDigester(ctrl)
	digesterMock.EXPECT().Digest(gomock.Any(), gomock.Any(), gomock.Any()).Return(ioutil.NopCloser(bytes.NewReader([]byte(""))), nil)

	grapherMock := NewMockGrapher(ctrl)
	grapherMock.EXPECT().Graph(gomock.Any(), key, gomock.Any()).Return(nil)

	markerMock := NewMockMarker(ctrl)
	markerMock.EXPECT().Unmark(gomock.Any(), key).Return(nil)

	start := time.Now().Add(-1 * time.Minute)
	stop := time.Now()
	payload := []byte(fmt.Sprintf(payloadTpl, key, start.Format(time.RFC3339Nano), stop.Format(time.RFC3339Nano)))
	r, _ := http.NewRequest(http.MethodPost, "/", ioutil.NopCloser(bytes.NewReader(payload)))
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))
	w := httptest.NewRecorder()
	handler := &Produce{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Grapher:      grapherMock,
		Marker:       markerMock,
		Digester:     digesterMock,
	}
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

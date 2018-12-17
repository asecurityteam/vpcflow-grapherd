package v1

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bitbucket.org/atlassian/logevent"
	"bitbucket.org/atlassian/vpcflow-grapherd/pkg/types"
	"github.com/golang/mock/gomock"
	"github.com/rs/xstats"
	"github.com/stretchr/testify/assert"
)

type timeMatcher struct {
	T time.Time
}

// Matches implements matcher and matches whether or not two times are equal
// using the preferred .Equal function from the time package.
func (m *timeMatcher) Matches(x interface{}) bool {
	t, ok := x.(time.Time)
	if !ok {
		return false
	}
	return m.T.Equal(t)
}

func (m *timeMatcher) String() string {
	return "matches two time.Time instances based on the evaluation of time.Equal()"
}

func newHandlerFunc(storage types.Storage, queuer types.Queuer, method string) http.HandlerFunc {
	handler := &GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storage,
		Queuer:       queuer,
	}
	if method == http.MethodGet {
		return handler.Get
	}
	return handler.Post
}

func TestHTTPBadRequest(t *testing.T) {
	tc := []struct {
		Name   string
		Start  string
		Stop   string
		Method string
	}{
		{
			Name:   "GET_bad_start",
			Start:  "invalid ts",
			Stop:   time.Now().Format(time.RFC3339Nano),
			Method: http.MethodGet,
		},
		{
			Name:   "GET_bad_stop",
			Start:  time.Now().Format(time.RFC3339Nano),
			Stop:   "invalid ts",
			Method: http.MethodGet,
		},
		{
			Name:   "GET_bad_range",
			Start:  time.Now().Format(time.RFC3339Nano),
			Stop:   time.Now().Add(-1 * time.Minute).Format(time.RFC3339Nano),
			Method: http.MethodGet,
		},
		{
			Name:   "POST_bad_start",
			Start:  "invalid ts",
			Stop:   time.Now().Format(time.RFC3339Nano),
			Method: http.MethodPost,
		},
		{
			Name:   "POST_bad_stop",
			Start:  time.Now().Format(time.RFC3339Nano),
			Stop:   "invalid ts",
			Method: http.MethodPost,
		},
		{
			Name:   "POST_bad_range",
			Start:  time.Now().Format(time.RFC3339Nano),
			Stop:   time.Now().Add(-1 * time.Minute).Format(time.RFC3339Nano),
			Method: http.MethodPost,
		},
	}

	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			r, _ := http.NewRequest(tt.Method, "/", nil)
			w := httptest.NewRecorder()

			q := r.URL.Query()
			q.Set("start", tt.Start)
			q.Set("stop", tt.Stop)
			r.URL.RawQuery = q.Encode()
			r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

			newHandlerFunc(nil, nil, tt.Method)(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		})
	}
}

func TestGetStorageErrors(t *testing.T) {
	tc := []struct {
		Name               string
		Error              error
		ExpectedStatusCode int
	}{
		{
			Name:               "GET_in_progress",
			Error:              types.ErrInProgress{},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Name:               "GET_not_found",
			Error:              types.ErrNotFound{},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Name:               "GET_unknown",
			Error:              errors.New("oops"),
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			start := time.Now().Add(-1 * time.Minute).Format(time.RFC3339Nano)
			stop := time.Now().Format(time.RFC3339Nano)
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			q := r.URL.Query()
			q.Set("start", start)
			q.Set("stop", stop)
			r.URL.RawQuery = q.Encode()
			r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

			storageMock := NewMockStorage(ctrl)
			storageMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, tt.Error)

			h := GrapherHandler{
				LogProvider:  logevent.FromContext,
				StatProvider: xstats.FromContext,
				Storage:      storageMock,
			}
			h.Get(w, r)

			assert.Equal(t, tt.ExpectedStatusCode, w.Result().StatusCode)
		})
	}
}

func TestGetHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	start := time.Now().Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	q.Set("start", start)
	q.Set("stop", stop)
	r.URL.RawQuery = q.Encode()
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

	data := "this is the graph you're looking for"
	readCloser := ioutil.NopCloser(bytes.NewReader([]byte(data)))

	storageMock := NewMockStorage(ctrl)
	storageMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(readCloser, nil)

	h := GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storageMock,
	}
	h.Get(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	body := w.Result().Body
	defer body.Close()

	result, _ := ioutil.ReadAll(body)
	assert.Equal(t, data, string(result))
}

func TestPostConflictInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	start := time.Now().Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	q.Set("start", start)
	q.Set("stop", stop)
	r.URL.RawQuery = q.Encode()
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

	storageMock := NewMockStorage(ctrl)
	storageMock.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, types.ErrInProgress{})

	h := GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storageMock,
	}
	h.Post(w, r)

	assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
}

func TestPostConflictGraphCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	start := time.Now().Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	q.Set("start", start)
	q.Set("stop", stop)
	r.URL.RawQuery = q.Encode()
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

	storageMock := NewMockStorage(ctrl)
	storageMock.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)

	h := GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storageMock,
	}
	h.Post(w, r)

	assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
}

func TestPostStorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	start := time.Now().Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	q.Set("start", start)
	q.Set("stop", stop)
	r.URL.RawQuery = q.Encode()
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

	storageMock := NewMockStorage(ctrl)
	storageMock.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, errors.New("oops"))

	h := GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storageMock,
	}
	h.Post(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestPostQueueError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	start := time.Now()
	stop := time.Now()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	q.Set("start", start.Format(time.RFC3339Nano))
	q.Set("stop", stop.Format(time.RFC3339Nano))
	r.URL.RawQuery = q.Encode()
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

	expectedStart := &timeMatcher{start.Truncate(time.Minute)}
	expectedStop := &timeMatcher{stop.Truncate(time.Minute)}

	storageMock := NewMockStorage(ctrl)
	storageMock.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil)
	queuerMock := NewMockQueuer(ctrl)
	queuerMock.EXPECT().Queue(gomock.Any(), gomock.Any(), expectedStart, expectedStop).Return(errors.New("oops"))

	h := GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storageMock,
		Queuer:       queuerMock,
	}
	h.Post(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestPostHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	start := time.Now()
	stop := time.Now()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	q.Set("start", start.Format(time.RFC3339Nano))
	q.Set("stop", stop.Format(time.RFC3339Nano))
	r.URL.RawQuery = q.Encode()
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

	expectedStart := &timeMatcher{start.Truncate(time.Minute)}
	expectedStop := &timeMatcher{stop.Truncate(time.Minute)}

	storageMock := NewMockStorage(ctrl)
	storageMock.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil)
	queuerMock := NewMockQueuer(ctrl)
	queuerMock.EXPECT().Queue(gomock.Any(), gomock.Any(), expectedStart, expectedStop).Return(nil)
	markerMock := NewMockMarker(ctrl)
	markerMock.EXPECT().Mark(gomock.Any(), gomock.Any()).Return(nil)

	h := GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storageMock,
		Queuer:       queuerMock,
		Marker:       markerMock,
	}
	h.Post(w, r)

	assert.Equal(t, http.StatusAccepted, w.Result().StatusCode)
}

func TestPostUnsuccessfulMark(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	start := time.Now()
	stop := time.Now()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	q.Set("start", start.Format(time.RFC3339Nano))
	q.Set("stop", stop.Format(time.RFC3339Nano))
	r.URL.RawQuery = q.Encode()
	r = r.WithContext(logevent.NewContext(context.Background(), logevent.New(logevent.Config{Output: ioutil.Discard})))

	expectedStart := &timeMatcher{start.Truncate(time.Minute)}
	expectedStop := &timeMatcher{stop.Truncate(time.Minute)}

	storageMock := NewMockStorage(ctrl)
	storageMock.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil)
	queuerMock := NewMockQueuer(ctrl)
	queuerMock.EXPECT().Queue(gomock.Any(), gomock.Any(), expectedStart, expectedStop).Return(nil)
	markerMock := NewMockMarker(ctrl)
	markerMock.EXPECT().Mark(gomock.Any(), gomock.Any()).Return(errors.New("OOPS"))

	h := GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Storage:      storageMock,
		Queuer:       queuerMock,
		Marker:       markerMock,
	}
	h.Post(w, r)

	// Shouldn't blow up
	assert.Equal(t, http.StatusAccepted, w.Result().StatusCode)
}

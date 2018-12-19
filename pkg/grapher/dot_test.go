package grapher

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const key = "foo"

func TestFailedConvert(t *testing.T) {
	input := []byte(`input data`)
	d := DOT{
		Converter: func(_ io.ReadCloser) (io.ReadCloser, error) {
			return nil, errors.New("")
		},
	}
	err := d.Graph(context.Background(), key, ioutil.NopCloser(bytes.NewReader(input)))
	assert.NotNil(t, err)
}

func TestFailedStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockStorage.EXPECT().Store(gomock.Any(), key, gomock.Any()).Return(errors.New(""))
	input := []byte(`input data`)
	d := DOT{
		Storage: mockStorage,
		Converter: func(_ io.ReadCloser) (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader([]byte("converted graph"))), nil
		},
	}
	err := d.Graph(context.Background(), key, ioutil.NopCloser(bytes.NewReader(input)))
	assert.NotNil(t, err)
}

func TestHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockStorage.EXPECT().Store(gomock.Any(), key, gomock.Any()).Return(nil)
	input := []byte(`input data`)
	d := DOT{
		Storage: mockStorage,
		Converter: func(_ io.ReadCloser) (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader([]byte("converted graph"))), nil
		},
	}
	err := d.Graph(context.Background(), key, ioutil.NopCloser(bytes.NewReader(input)))
	assert.Nil(t, err)
}

package grapherd

import (
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

func TestMustEnv(t *testing.T) {
	tc := []struct {
		Name  string
		Key   string
		Value string
		Set   bool
	}{
		{
			Name:  "var unset",
			Key:   "key",
			Value: "",
			Set:   false,
		},
		{
			Name:  "var set",
			Key:   "key",
			Value: "value",
			Set:   true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			originalValue, existing := os.LookupEnv(tt.Key)
			if existing {
				defer os.Setenv(tt.Key, originalValue)
			}
			os.Unsetenv(tt.Key)
			if tt.Set {
				os.Setenv(tt.Key, tt.Value)
				require.Equal(t, tt.Value, mustEnv(tt.Key))
				return
			}
			require.Panics(t, func() {
				mustEnv(tt.Key)
			})

		})
	}
}

func TestServiceInitSuccess(t *testing.T) {
	// save current environment variables, and restore them
	// after the test ends
	environ := os.Environ()
	os.Clearenv()
	defer func() {
		for _, e := range environ {
			envPair := strings.Split(e, "=")
			os.Setenv(envPair[0], envPair[1])
		}
	}()

	// set required test environment variables
	os.Setenv("USE_IAM", "true")
	os.Setenv("GRAPH_STORAGE_BUCKET_REGION", "n/a")
	os.Setenv("GRAPH_PROGRESS_BUCKET_REGION", "n/a")
	os.Setenv("STREAM_APPLIANCE_ENDPOINT", "n/a")
	os.Setenv("GRAPH_PROGRESS_TIMEOUT", "1")
	os.Setenv("GRAPH_PROGRESS_BUCKET", "n/a")
	os.Setenv("GRAPH_STORAGE_BUCKET", "n/a")
	os.Setenv("DIGESTER_ENDPOINT", "n/a")
	os.Setenv("DIGESTER_POLLING_TIMEOUT", "1")
	os.Setenv("DIGESTER_POLLING_INTERVAL", "1")

	s := &Service{}
	require.Nil(t, s.init())
}

func TestServiceBindRoutesSuccess(t *testing.T) {
	environ := os.Environ()
	os.Clearenv()
	defer func() {
		for _, e := range environ {
			envPair := strings.Split(e, "=")
			os.Setenv(envPair[0], envPair[1])
		}
	}()

	// set required test environment variables
	os.Setenv("USE_IAM", "true")
	os.Setenv("GRAPH_STORAGE_BUCKET_REGION", "n/a")
	os.Setenv("GRAPH_PROGRESS_BUCKET_REGION", "n/a")
	os.Setenv("STREAM_APPLIANCE_ENDPOINT", "n/a")
	os.Setenv("GRAPH_PROGRESS_TIMEOUT", "1")
	os.Setenv("GRAPH_PROGRESS_BUCKET", "n/a")
	os.Setenv("GRAPH_STORAGE_BUCKET", "n/a")
	os.Setenv("DIGESTER_ENDPOINT", "n/a")
	os.Setenv("DIGESTER_POLLING_TIMEOUT", "1")
	os.Setenv("DIGESTER_POLLING_INTERVAL", "1")

	router := chi.NewMux()
	s := &Service{}
	require.Nil(t, s.BindRoutes(router))
}

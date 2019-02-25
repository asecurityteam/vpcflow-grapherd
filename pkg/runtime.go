package grapherd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/asecurityteam/go-vpcflow"
	"github.com/asecurityteam/logevent"
	"github.com/asecurityteam/transport"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/digester"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/grapher"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/handlers/v1"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/marker"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/queuer"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/storage"
	"github.com/asecurityteam/vpcflow-grapherd/pkg/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi"
	"github.com/rs/xstats"
)

// Server is an interface for starting/stopping an HTTP server
type Server interface {
	// ListenAndServe starts the HTTP server in a blocking call.
	ListenAndServe() error
	// Shutdown stops the server from accepting new connections.
	// If the given context expires before shutdown is complete then
	// the context error is returned.
	Shutdown(ctx context.Context) error
}

// Service is a container for all of the pluggable modules used by the service
type Service struct {
	// QueuerHTTPClient is the client to be used with the default Queuer module.
	// If no client is provided, the default client will be used.
	QueuerHTTPClient *http.Client

	// DigesterHTTPClient is the client to be used with the default Digester module.
	// If no client is provided, the default client will be used.
	DigesterHTTPClient *http.Client

	// Middleware is a list of service middleware to install on the router.
	// The set of prepackaged middleware can be found in pkg/plugins.
	Middleware []func(http.Handler) http.Handler

	// Queuer is responsible for queuing graphing jobs which will eventually be consumed
	// by the Produce handler. The built in Queuer POSTs to an HTTP endpoint.
	Queuer types.Queuer

	// Storage provides a mechanism to hook into a persistent store for the graphs. The
	// built in Storage uses S3 as the persistent storage for graph content.
	Storage types.Storage

	// Marker is responsible for marking which graph jobs are inprogress. The built in
	// Marker uses S3 to hold this state.
	Marker types.Marker

	// Digester is responsible for creating a digest of VPC logs for a given time range.
	// The built in digester calls out to a digester service.
	Digester types.Digester
}

func (s *Service) init() error {
	var err error
	storageClient, err := createS3Client(mustEnv("GRAPH_STORAGE_BUCKET_REGION"))
	if err != nil {
		return err
	}
	progressClient, err := createS3Client(mustEnv("GRAPH_PROGRESS_BUCKET_REGION"))
	if err != nil {
		return err
	}

	if s.Queuer == nil {
		streamApplianceEndpoint := mustEnv("STREAM_APPLIANCE_ENDPOINT")
		streamApplianceURL, err := url.Parse(streamApplianceEndpoint)
		if err != nil {
			return err
		}
		if s.QueuerHTTPClient == nil {
			s.QueuerHTTPClient = defaultHTTPClient()
		}
		s.Queuer = &queuer.GraphQueuer{
			Client:   s.QueuerHTTPClient,
			Endpoint: streamApplianceURL,
		}
	}
	if s.Storage == nil {
		progressTimeoutStr := mustEnv("GRAPH_PROGRESS_TIMEOUT")
		progressTimeoutInt, err := strconv.Atoi(progressTimeoutStr)
		if err != nil {
			return err
		}
		s.Storage = &storage.InProgress{
			Bucket: mustEnv("GRAPH_PROGRESS_BUCKET"),
			Client: progressClient,
			Storage: &storage.S3{
				Bucket: mustEnv("GRAPH_STORAGE_BUCKET"),
				Client: storageClient,
			},
			Timeout: time.Millisecond * time.Duration(progressTimeoutInt),
		}
	}
	if s.Marker == nil {
		s.Marker = &marker.ProgressMarker{
			Bucket: mustEnv("GRAPH_PROGRESS_BUCKET"),
			Client: progressClient,
		}
	}
	if s.Digester == nil {
		digesterEndpoint := mustEnv("DIGESTER_ENDPOINT")
		digesterURL, err := url.Parse(digesterEndpoint)
		if err != nil {
			return err
		}
		durationStr := mustEnv("DIGESTER_POLLING_TIMEOUT")
		durationMs, err := strconv.Atoi(durationStr)
		if err != nil {
			return err
		}
		intervalStr := mustEnv("DIGESTER_POLLING_INTERVAL")
		intervalMs, err := strconv.Atoi(intervalStr)
		if err != nil {
			return err
		}
		if s.DigesterHTTPClient == nil {
			s.DigesterHTTPClient = defaultHTTPClient()
		}
		s.Digester = &digester.HTTP{
			Client:          s.DigesterHTTPClient,
			Endpoint:        digesterURL,
			PollTimeout:     time.Duration(durationMs) * time.Millisecond,
			PollingInterval: time.Duration(intervalMs) * time.Millisecond,
		}
	}
	return nil
}

// BindRoutes binds the service handlers to the provided router
func (s *Service) BindRoutes(router chi.Router) error {
	if err := s.init(); err != nil {
		return err
	}
	grapherHandler := &v1.GrapherHandler{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Queuer:       s.Queuer,
		Storage:      s.Storage,
		Marker:       s.Marker,
	}
	produceHandler := &v1.Produce{
		LogProvider:  logevent.FromContext,
		StatProvider: xstats.FromContext,
		Digester:     s.Digester,
		Marker:       s.Marker,
		Grapher: &grapher.DOT{
			Converter: vpcflow.DOTConverter,
			Storage:   s.Storage,
		},
	}
	router.Use(s.Middleware...)
	router.Post("/", grapherHandler.Post)
	router.Get("/", grapherHandler.Get)
	router.Post("/{topic}/{event}", produceHandler.ServeHTTP)
	return nil
}

// Runtime is the app configuration and execution point
type Runtime struct {
	Server      Server
	ExitSignals []types.ExitSignal
}

// Run runs the application
func (r *Runtime) Run() error {
	exit := make(chan error)

	for _, f := range r.ExitSignals {
		go func(f func() chan error) {
			exit <- <-f()
		}(f)
	}

	go func() {
		exit <- r.Server.ListenAndServe()
	}()

	err := <-exit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = r.Server.Shutdown(ctx)

	return err
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return val
}

func createS3Client(region string) (*s3.S3, error) {
	useIAM := mustEnv("USE_IAM")
	useIAMFlag, err := strconv.ParseBool(useIAM)
	if err != nil {
		return nil, err
	}
	cfg := aws.NewConfig()
	cfg.Region = aws.String(region)
	if !useIAMFlag {
		cfg.Credentials = credentials.NewChainCredentials([]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Filename: os.Getenv("AWS_CREDENTIALS_FILE"),
				Profile:  os.Getenv("AWS_CREDENTIALS_PROFILE"),
			},
		})
	}
	awsSession, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	return s3.New(awsSession), nil
}

func defaultHTTPClient() *http.Client {
	retrier := transport.NewRetrier(
		transport.NewFixedBackoffPolicy(50*time.Millisecond),
		transport.NewLimitedRetryPolicy(3),
		transport.NewStatusCodeRetryPolicy(500, 502, 503),
	)
	base := transport.NewFactory(
		transport.OptionDefaultTransport,
		transport.OptionTLSHandshakeTimeout(time.Second),
		transport.OptionMaxIdleConns(100),
	)
	recycler := transport.NewRecycler(
		transport.Chain{retrier}.ApplyFactory(base),
		transport.RecycleOptionTTL(10*time.Minute),
		transport.RecycleOptionTTLJitter(time.Minute),
	)
	return &http.Client{Transport: recycler}
}

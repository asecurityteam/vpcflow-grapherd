package plugins

import (
	"net/http"
	"time"

	"bitbucket.org/atlassian/transport"
)

// HTTPClientProvider is a factory function which creates a new http client
func HTTPClientProvider() *http.Client {
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

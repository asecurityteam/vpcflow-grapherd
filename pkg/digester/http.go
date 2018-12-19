package digester

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	queryStart = "start"
	queryStop  = "stop"
)

// HTTP is used to create a new digest
type HTTP struct {
	Client          *http.Client
	Endpoint        *url.URL
	PollAttempts    int
	PollingInterval time.Duration
}

// Digest starts a new digest job, and waits for its completion. On successful completion, Digest will return the
// digested content.
func (c *HTTP) Digest(ctx context.Context, start, stop time.Time) (io.ReadCloser, error) {
	req, err := newDigestRequest(c.Endpoint, http.MethodPost, start, stop)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	// If the status code is 202 or 409 the job is either a) scheduled b) in progress or c) created.
	// In all of these cases, we want to poll the GET endpoint for a 200. Otherwise, report an error.
	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusConflict {
		data, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("Received unexpected response from digester %d: %s", res.StatusCode, data)
	}

	return c.waitForDigest(ctx, start, stop)
}

func (c *HTTP) waitForDigest(ctx context.Context, start, stop time.Time) (io.ReadCloser, error) {
	req, err := newDigestRequest(c.Endpoint, http.MethodGet, start, stop)
	if err != nil {
		return nil, err
	}
	for attempts := 1; attempts <= c.PollAttempts; attempts++ {
		res, err := c.Client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusOK { // digest is ready
			return extractDigest(res.Body)
		}
		if res.StatusCode != http.StatusNoContent {
			data, _ := ioutil.ReadAll(res.Body)
			return nil, fmt.Errorf("Received unexpected response while polling digester %d: %s", res.StatusCode, data)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("request time out reached after %d attempt(s): %s", attempts, ctx.Err().Error())
		case <-time.After(c.PollingInterval):
		}
	}
	return nil, fmt.Errorf("Max digester poll attempts reached: %d", c.PollAttempts)
}

func extractDigest(r io.Reader) (io.ReadCloser, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

func newDigestRequest(endpoint *url.URL, method string, start, stop time.Time) (*http.Request, error) {
	u, _ := url.Parse(endpoint.String())
	q := u.Query()
	q.Set(queryStart, start.Format(time.RFC3339Nano))
	q.Set(queryStop, stop.Format(time.RFC3339Nano))
	u.RawQuery = q.Encode()
	return http.NewRequest(method, u.String(), nil)
}

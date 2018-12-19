package queuer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type payload struct {
	ID    string `json:"id"`
	Start string `json:"start"`
	Stop  string `json:"stop"`
}

// GraphQueuer is a Queuer implementation which queues graph jobs onto a streaming appliance
type GraphQueuer struct {
	Endpoint *url.URL
	Client   *http.Client
}

// Queue enqueues a graph job onto a streaming appliance
func (q *GraphQueuer) Queue(ctx context.Context, id string, start, stop time.Time) error {
	body := payload{
		ID:    id,
		Start: start.Format(time.RFC3339Nano),
		Stop:  stop.Format(time.RFC3339Nano),
	}
	rawBody, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, q.Endpoint.String(), bytes.NewReader(rawBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := q.Client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from streaming appliance: %d", res.StatusCode)
	}
	return nil
}

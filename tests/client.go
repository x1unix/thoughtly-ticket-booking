package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
	"github.com/x1unix/thoughtly-ticket-booking/internal/server"
)

type Client struct {
	addr       string
	httpClient *http.Client
}

func NewClient(addr string) (*Client, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("bad listen address: %w", err)
	}

	if host == "" {
		host = "localhost"
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, port)
	return &Client{
		addr:       baseURL,
		httpClient: http.DefaultClient,
	}, nil
}

func (c *Client) WaitForServer(retryCount int, interval time.Duration) error {
	var lastErr error
	for i := range retryCount {
		if i != 0 {
			time.Sleep(interval)
		}

		r, err := c.newGetRequest("/api/ping")
		if err != nil {
			return err
		}

		if err := c.doRequest(r, nil); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return lastErr
}

func (c *Client) CreateEvent(t *testing.T, body booking.EventCreateParams) *booking.EventCreateResult {
	req, err := c.newJSONRequest("/api/events", body)
	require.NoError(t, err)

	rsp := &booking.EventCreateResult{}
	require.NoError(t, c.doRequest(req, rsp))
	return rsp
}

func (c *Client) GetEvents(t *testing.T, body booking.EventCreateParams) *server.ListEventsResponse {
	req, err := c.newGetRequest("/api/events")
	require.NoError(t, err)

	rsp := &server.ListEventsResponse{}
	require.NoError(t, c.doRequest(req, rsp))
	return rsp
}

func (c *Client) GetTicketTiers(t *testing.T, eventID uuid.UUID) *server.ListTiersResponse {
	req, err := c.newGetRequest("/api/events/", eventID.String(), "/tiers")
	require.NoError(t, err)

	rsp := &server.ListTiersResponse{}
	require.NoError(t, c.doRequest(req, rsp))
	return rsp
}

func (c *Client) newGetRequest(parts ...string) (*http.Request, error) {
	uri := c.addr + strings.Join(parts, "")
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("%q: cannot create request: %w", uri, err)
	}

	return req, nil
}

func (c *Client) newJSONRequest(rpath string, body any) (*http.Request, error) {
	uri := c.addr + rpath

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%q: failed to marshal request: %w", uri, err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("%q: cannot create request: %w", uri, err)
	}

	return req, nil
}

func (c *Client) doRequest(req *http.Request, out any) error {
	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%q: failed to send request: %w", req.URL, err)
	}

	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("%q: bad status code: %q", req.URL, rsp.Status)
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(rsp.Body).Decode(out); err != nil {
		return fmt.Errorf("%q: failed to unmarshal request: %w", req.URL, err)
	}

	return nil
}

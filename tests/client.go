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
	t.Helper()
	req, err := c.newJSONRequest("/api/events", body)
	require.NoError(t, err)

	rsp := &booking.EventCreateResult{}
	require.NoError(t, c.doRequest(req, rsp))
	return rsp
}

func (c *Client) GetEvents(t *testing.T, body booking.EventCreateParams) *server.ListEventsResponse {
	t.Helper()
	req, err := c.newGetRequest("/api/events")
	require.NoError(t, err)

	rsp := &server.ListEventsResponse{}
	require.NoError(t, c.doRequest(req, rsp))
	return rsp
}

func (c *Client) GetTicketTiers(t *testing.T, eventID uuid.UUID) *server.ListTiersResponse {
	t.Helper()
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
		return nil, fmt.Errorf("%s %q: cannot create request: %w", req.Method, uri, err)
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
		return nil, fmt.Errorf("%s %q: cannot create request: %w", req.Method, uri, err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) doRequest(req *http.Request, out any) error {
	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %q: failed to send request: %w", req.Method, req.URL, err)
	}

	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return tryReadError(req, rsp)
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(rsp.Body).Decode(out); err != nil {
		return fmt.Errorf("%q: failed to unmarshal request: %w", req.URL, err)
	}

	return nil
}

type ResponseError struct {
	Code     int
	Response *server.ErrorResponse
}

func newResponseError(code int, body *server.ErrorResponse) *ResponseError {
	return &ResponseError{
		Code:     code,
		Response: body,
	}
}

func (err *ResponseError) Error() string {
	return fmt.Sprintf("%s (status: %d)", err.Response.Error, err.Code)
}

func tryReadError(req *http.Request, rsp *http.Response) error {
	ctype := rsp.Header.Get("Content-Type")
	if strings.HasPrefix(ctype, "application/json") {
		errBody := &server.ErrorResponse{}
		err := json.NewDecoder(rsp.Body).Decode(errBody)
		if err == nil {
			return newResponseError(rsp.StatusCode, errBody)
		}
	}

	return fmt.Errorf("%s %q: bad status code: %q", req.Method, req.URL, rsp.Status)
}

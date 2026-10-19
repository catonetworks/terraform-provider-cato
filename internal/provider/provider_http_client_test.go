package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type requestTransportFunc func(*http.Request) (*http.Response, error)

func (f requestTransportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAPIRequestDeadlineCoversRetries(t *testing.T) {
	client := buildRetryHTTPClient(&retryClientConfig{retryMax: 1})
	var deadlines []time.Time
	client.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport = requestTransportFunc(
		func(req *http.Request) (*http.Response, error) {
			deadline, _ := req.Context().Deadline()
			deadlines = append(deadlines, deadline)
			status := http.StatusOK
			if len(deadlines) == 1 {
				status = http.StatusTooManyRequests
			}
			return &http.Response{StatusCode: status, Header: make(http.Header),
				Body: io.NopCloser(strings.NewReader("{}")), Request: req}, nil
		})

	response, err := client.Get("https://example.invalid/graphql")
	if err != nil {
		t.Fatal(err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if len(deadlines) != 2 {
		t.Fatalf("expected one retry, got %d attempts", len(deadlines))
	}
	if deadlines[0].IsZero() || time.Until(deadlines[0]) > 2*time.Minute {
		t.Fatal("outbound API request has no bounded deadline")
	}
	if !deadlines[0].Equal(deadlines[1]) {
		t.Fatal("retry reset the API request deadline")
	}
}

func TestAPIRequestHonorsEarlierCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		<-req.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	client := buildRetryHTTPClient(&retryClientConfig{retryMax: 5})
	response, err := client.Do(req)
	if response != nil {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error for stalled response, got %v", err)
	}
}

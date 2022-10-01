package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/anzx/fabric-cards/pkg/middleware/httplogging"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"
	"github.com/hashicorp/go-retryablehttp"
)

// NewHTTPClient returns a preconfigured default http client
func NewHTTPClient(service names.Service) *http.Client {
	return &http.Client{
		Transport: extractor.NewHTTPTransport(http.DefaultTransport, service),
		Timeout:   30 * time.Second,
	}
}

func NewHTTPClientWithLog(transport http.RoundTripper, payloadLoggingDecider httplogging.PayloadLoggingDecider, service names.Service) *http.Client {
	return &http.Client{
		Transport: httplogging.NewLoggedTransport(
			extractor.NewHTTPTransport(transport, service),
			httplogging.NewLogger(payloadLoggingDecider)),
		Timeout: 30 * time.Second,
	}
}

// NewHTTPClientWithLogAndRetry returns a preconfigured default http client, ochttp.Transport should be the outermost transport
func NewHTTPClientWithLogAndRetry(maxRetries int, payloadLoggingDecider httplogging.PayloadLoggingDecider, service names.Service) *http.Client {
	client := NewHTTPClientWithLog(http.DefaultTransport, payloadLoggingDecider, service)
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	retryClient.HTTPClient = client
	return retryClient.StandardClient()
}

type ClientTagExtKey struct{}

type ClientTagExt struct {
	Operation string
}

func getClientTagExt(ctx context.Context) *ClientTagExt {
	if data, ok := ctx.Value(ClientTagExtKey{}).(ClientTagExt); ok {
		return &data
	}
	return nil
}

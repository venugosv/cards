package httplogging

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/sanitize"

	"github.com/anzx/pkg/log"
)

type loggedRoundTripper struct {
	rt  http.RoundTripper
	log HTTPLogger
}

func (c *loggedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	c.log.LogRequest(request)
	startTime := time.Now()
	response, err := c.rt.RoundTrip(request)
	duration := time.Since(startTime)
	c.log.LogResponse(request, response, err, duration)
	return response, err
}

// NewLoggedTransport takes an http.RoundTripper and returns a new one that logs requests and responses
func NewLoggedTransport(rt http.RoundTripper, log HTTPLogger) http.RoundTripper {
	return &loggedRoundTripper{rt: rt, log: log}
}

// HTTPLogger defines the interface to log http request and responses
type HTTPLogger interface {
	LogRequest(*http.Request)
	LogResponse(*http.Request, *http.Response, error, time.Duration)
}

type PayloadLoggingDecider func(url *url.URL) bool

var defaultLogPayloadDecider PayloadLoggingDecider = func(url *url.URL) bool { return true }

type httpLogger struct {
	decider PayloadLoggingDecider
}

func NewLogger(payloadLoggingDecider PayloadLoggingDecider) *httpLogger {
	if payloadLoggingDecider == nil {
		payloadLoggingDecider = defaultLogPayloadDecider
	}
	return &httpLogger{
		decider: payloadLoggingDecider,
	}
}

func (l *httpLogger) LogRequest(req *http.Request) {
	var bodyBytes []byte
	if req.Body != nil {
		// Read the content
		bodyBytes, _ = io.ReadAll(req.Body)
		defer req.Body.Close()
		// Restore the io.ReadCloser to its original state
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Always append headers
	requestHeaders := sanitize.GetHeaders(req.Header)
	attr := []log.Attribute{
		log.Any("headers", requestHeaders),
	}

	if l.decider(req.URL) {
		sanitizedRequestBody := sanitize.ConvertToLoggableFieldValueWithType(bodyBytes, req.Header.Get("Content-type"))
		attr = append(attr, log.Any("requestBody", sanitizedRequestBody))
	}

	msg := fmt.Sprintf("Request %s %s", req.Method, req.URL.String())
	log.Info(req.Context(), msg, attr...)
}

func (l *httpLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	duration /= time.Millisecond

	if err != nil {
		logf.Err(req.Context(), err)
		return
	}

	var bodyBytes []byte
	if res.Body != nil {
		// Read the content
		bodyBytes, _ = io.ReadAll(res.Body)
		defer res.Body.Close()
		// Restore the io.ReadCloser to its original state
		res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Always append headers
	requestHeaders := sanitize.GetHeaders(req.Header)
	attr := []log.Attribute{
		log.Any("headers", requestHeaders),
	}

	if l.decider(req.URL) {
		sanitizedResponseBody := sanitize.ConvertToLoggableFieldValueWithType(bodyBytes, req.Header.Get("Content-type"))
		attr = append(attr, log.Any("requestBody", sanitizedResponseBody))
	}

	msg := fmt.Sprintf("Response method=%s status=%d durationMs=%d %s", req.Method, res.StatusCode, duration, req.URL.String())
	log.Info(req.Context(), msg, attr...)
}

package httplogging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	defaultLog "log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/anzx/pkg/log/fabriclog"

	"github.com/stretchr/testify/require"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func Test_httpLogger_LogRequest(t *testing.T) {
	tests := []struct {
		name   string
		method string
		url    string
		body   []byte
		want   string
	}{
		{
			name:   "masked successfully",
			method: http.MethodPost,
			url:    "https://apisit.corp.dev.anz",
			body:   []byte("{\"PAN\":\"4514170000000001\"}"),
			want:   "451417******0001",
		}, {
			name:   "mask ignored",
			method: http.MethodPut,
			url:    "https://apisit.corp.dev.anz/test",
			body:   []byte("{\"PAN\":\"notaPAN\"}"),
			want:   "notaPAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			fabriclog.Init(fabriclog.WithConsoleWriter(buf))
			l := &httpLogger{
				decider: defaultLogPayloadDecider,
			}
			request, _ := http.NewRequestWithContext(context.Background(), tt.method, tt.url, bytes.NewBuffer(tt.body))
			l.LogRequest(request)

			assert.Contains(t, buf.String(), tt.method)
			assert.Contains(t, buf.String(), tt.url)
			assert.Contains(t, buf.String(), tt.want)
		})
	}
}

func Test_httpLogger_LogResponse(t *testing.T) {
	type req struct {
		method string
		url    string
	}
	type res struct {
		statusCode int
		body       []byte
		want       string
	}
	tests := []struct {
		name string
		req  req
		res  res
		err  error
	}{
		{
			name: "successfully mask pan",
			req: req{
				method: http.MethodPost,
				url:    "https://apisit.corp.dev.anz",
			},
			res: res{
				statusCode: http.StatusOK,
				body:       []byte("{\"PAN\":\"4514170000000001\"}"),
				want:       "451417******0001",
			},
		},
		{
			name: "successfully mask pii data",
			req: req{
				method: http.MethodPost,
				url:    "https://apisit.corp.dev.anz",
			},
			res: res{
				statusCode: http.StatusOK,
				body:       []byte("{\"firstName\":\"Joe\"}"),
				want:       "\"firstName\":\"*\"",
			},
		},
		{
			name: "display unmasked successfully",
			req: req{
				method: http.MethodPut,
				url:    "https://apisit.corp.dev.anz/test",
			},
			res: res{
				statusCode: http.StatusUnauthorized,
				body:       []byte("{\"PAN\":\"notapan\"}"),
				want:       "notapan",
			},
		},
		{
			name: "has error",
			err:  errors.New("oh no"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			fabriclog.Init(fabriclog.WithConsoleWriter(buf))
			l := &httpLogger{
				decider: defaultLogPayloadDecider,
			}

			request, _ := http.NewRequestWithContext(context.Background(), tt.req.method, tt.req.url, nil)
			response := &http.Response{
				Body:       io.NopCloser(bytes.NewReader(tt.res.body)),
				StatusCode: tt.res.statusCode,
			}

			l.LogResponse(request, response, tt.err, 10)

			if tt.err == nil {
				assert.Contains(t, buf.String(), fmt.Sprintf("method=%s", tt.req.method))
				assert.Contains(t, buf.String(), fmt.Sprintf("status=%d", tt.res.statusCode))
				assert.Contains(t, buf.String(), tt.req.url)
				assert.Contains(t, buf.String(), tt.res.want)
			} else {
				assert.Contains(t, buf.String(), tt.err.Error())
			}
		})
	}
}

func Test_httpLogger_LogResponseWithoutPayloadLogging(t *testing.T) {
	type req struct {
		method string
		url    string
	}
	type res struct {
		statusCode int
		body       []byte
	}
	tests := []struct {
		name string
		req  req
		res  res
		err  error
	}{
		{
			name: "successfully mask pan",
			req: req{
				method: http.MethodPost,
				url:    "https://apisit.corp.dev.anz",
			},
			res: res{
				statusCode: http.StatusOK,
				body:       []byte("{\"PAN\":\"4514170000000001\"}"),
			},
		},
		{
			name: "successfully mask pii data",
			req: req{
				method: http.MethodPost,
				url:    "https://apisit.corp.dev.anz",
			},
			res: res{
				statusCode: http.StatusOK,
				body:       []byte("{\"firstName\":\"Joe\"}"),
			},
		},
		{
			name: "display unmasked successfully",
			req: req{
				method: http.MethodPut,
				url:    "https://apisit.corp.dev.anz/test",
			},
			res: res{
				statusCode: http.StatusUnauthorized,
				body:       []byte("{\"PAN\":\"notapan\"}"),
			},
		},
		{
			name: "has error",
			err:  errors.New("oh no"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			fabriclog.Init(fabriclog.WithConsoleWriter(buf))
			l := &httpLogger{
				decider: func(url *url.URL) bool {
					return false
				},
			}

			request, _ := http.NewRequestWithContext(context.Background(), tt.req.method, tt.req.url, nil)
			response := &http.Response{
				Body:       io.NopCloser(bytes.NewReader(tt.res.body)),
				StatusCode: tt.res.statusCode,
			}

			l.LogResponse(request, response, tt.err, 10)

			if tt.err == nil {
				assert.Contains(t, buf.String(), fmt.Sprintf("method=%s", tt.req.method))
				assert.Contains(t, buf.String(), fmt.Sprintf("status=%d", tt.res.statusCode))
				assert.Contains(t, buf.String(), tt.req.url)
				require.NotContains(t, buf.String(), string(tt.res.body))
			} else {
				assert.Contains(t, buf.String(), tt.err.Error())
			}
		})
	}
}

func Test_httpLogger_newLogger(t *testing.T) {
	t.Run("", func(t *testing.T) {
		logger := NewLogger(nil)
		assert.NotNil(t, logger)
	})
	t.Run("", func(t *testing.T) {
		logger := NewLogger(nil)
		assert.NotNil(t, logger)
	})
}

func TestHttpTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		method string
		url    string
	}{
		{
			name:   "masked successfully",
			method: http.MethodPost,
			url:    "apisit.corp.dev.anz",
		}, {
			name:   "mask ignored",
			method: http.MethodPut,
			url:    "apisit.corp.dev.anz/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			defaultLog.SetOutput(&buf)
			defer func() {
				defaultLog.SetOutput(os.Stderr)
			}()

			transport := NewLoggedTransport(http.DefaultTransport, TestingLogger{})

			request, _ := http.NewRequest(tt.method, tt.url, nil)

			_, _ = transport.RoundTrip(request)
			output := buf.String()
			assert.Contains(t, output, fmt.Sprintf("method=%s", tt.method))
			assert.Contains(t, output, tt.url)
		})
	}
}

type TestingLogger struct{}

func (dl TestingLogger) LogRequest(*http.Request) {
}

func (dl TestingLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	duration /= time.Millisecond
	if err != nil {
		defaultLog.Printf("HTTP Request method=%s host=%s path=%s status=error durationMs=%d error=%q", req.Method, req.Host, req.URL.Path, duration, err.Error())
	} else {
		defaultLog.Printf("HTTP Request method=%s host=%s path=%s status=%d durationMs=%d", req.Method, req.Host, req.URL.Path, res.StatusCode, duration)
	}
}

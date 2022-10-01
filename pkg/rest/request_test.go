package rest

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anzx/pkg/monitoring/names"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultHTTPClient_TimeoutValues(t *testing.T) {
	test := struct {
		name string
		want http.Client
	}{
		name: "timeout values",
		want: http.Client{Timeout: 30 * time.Second},
	}
	got := NewHTTPClient(names.Unknown)
	assert.EqualValues(t, test.want.Timeout, got.Timeout)
}

func TestGetClientTagExt(t *testing.T) {
	// test with empty context
	assert.Nil(t, getClientTagExt(context.Background()))
}

func TestNewHTTPClientWithLogAndRetry(t *testing.T) {
	t.Run("retry until reaches maxRetries, return error", func(t *testing.T) {
		var handler http.HandlerFunc = func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		server := httptest.NewServer(handler)

		client := NewHTTPClientWithLogAndRetry(2, nil, names.Unknown)
		request, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer([]byte("haha")))
		_, err := client.Do(request)

		assert.Contains(t, err.Error(), "giving up after 3 attempt(s)")
	})

	t.Run("retry with success", func(t *testing.T) {
		count := 1
		var handler http.HandlerFunc = func(rw http.ResponseWriter, req *http.Request) {
			if count == 0 {
				_, _ = rw.Write(nil)
			} else {
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				count--
			}
		}

		server := httptest.NewServer(handler)
		client := NewHTTPClientWithLogAndRetry(2, nil, names.Unknown)
		request, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer([]byte("haha")))
		_, err := client.Do(request)

		assert.Nil(t, err)
	})
}

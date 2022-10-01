package servers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatcherToGRPCServer(t *testing.T) {
	t.Run("successfully dispatch to GRPC Server", func(t *testing.T) {
		status := http.StatusFound
		body := `{"dispatcher":"grpc"}`

		expectedHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			err := json.NewEncoder(w).Encode(body)
			require.NoError(t, err)
		}

		handler := grpcDispatcher(context.Background(),
			http.HandlerFunc(expectedHandler),
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

		req, err := http.NewRequest("GET", "/grpc", nil)
		require.NoError(t, err)
		req.Header.Set(contentType, grpcContentType)
		req.ProtoMajor = 2

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, status, recorder.Code)
		assert.Contains(t, strings.ReplaceAll(recorder.Body.String(), `\`, ""), body)
	})
}

func TestDispatcherToHTTPServer(t *testing.T) {
	t.Run("successfully dispatch to HTTP Server", func(t *testing.T) {
		status := http.StatusFound
		body := `{"dispatcher":"http"}`

		expectedHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			err := json.NewEncoder(w).Encode(body)
			require.NoError(t, err)
		}

		handler := grpcDispatcher(context.Background(),
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
			http.HandlerFunc(expectedHandler))

		req, err := http.NewRequest("GET", "/http", nil)
		require.NoError(t, err)
		req.Header.Set(contentType, "application/json")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, status, recorder.Code)
		assert.Contains(t, strings.ReplaceAll(recorder.Body.String(), `\`, ""), body)
	})
}

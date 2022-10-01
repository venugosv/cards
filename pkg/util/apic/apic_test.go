package apic

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/util/testutil"
	"github.com/stretchr/testify/assert"
)

type mockSecretManager struct {
	name    string
	payload string
	err     error
}

func (m mockSecretManager) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name:    m.name,
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(m.payload)},
	}, m.err
}

func TestAPIC_addStandardHeaders(t *testing.T) {
	t.Run("add standard apic request Headers", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "localhost:8080", bytes.NewBuffer([]byte{}))
		addStandardHeaders(r, "clientID")
		assert.Equal(t, applicationjson, r.Header.Get("Accept"))
		assert.Equal(t, applicationjson, r.Header.Get("content-Type"))
		assert.Equal(t, fabric, r.Header.Get("anz-application-id"))
		assert.Equal(t, ax2, r.Header.Get("X-Originating-App"))
		assert.Equal(t, "clientID", r.Header.Get("x-ibm-client-id"))
	})
	t.Run("add standard apic request Headers for get request", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "localhost:8080", nil)
		addStandardHeaders(r, "clientID")
		assert.Equal(t, applicationjson, r.Header.Get("Accept"))
		assert.Equal(t, fabric, r.Header.Get("anz-application-id"))
		assert.Equal(t, ax2, r.Header.Get("X-Originating-App"))
		assert.Equal(t, "clientID", r.Header.Get("x-ibm-client-id"))
		assert.NotContains(t, r.Header, "content-Type")
	})
}

func TestAPIC_propagateAuthorisationHeader(t *testing.T) {
	t.Run("propagate authorisation header", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "localhost:8080", bytes.NewBuffer([]byte{}))
		err := propagateAuthorisationHeader(testutil.GetContext(true), r)
		assert.Nil(t, err)
		assert.Contains(t, r.Header.Get("authorization"), "Bearer")
	})
}

func TestAPIC_noAuthorisationHeader(t *testing.T) {
	t.Run("propagate authorisation header", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "localhost:8080", bytes.NewBuffer([]byte{}))
		err := propagateAuthorisationHeader(testutil.GetContext(false), r)
		assert.NotNil(t, err)
	})
}

func TestAPIC_Client(t *testing.T) {
	key := "EchidnaKey"

	gsmClient := &gsm.Client{
		SM: mockSecretManager{
			name:    "testName",
			payload: "THIS-IS-A-CLIENT-ID",
		},
	}

	gsmNoSecretReturned := &gsm.Client{
		SM: mockSecretManager{
			name:    "testName",
			payload: "",
		},
	}

	t.Run("successfully create a new client", func(t *testing.T) {
		apicClient, err := NewAPICClient(context.Background(), key, &http.Client{}, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, apicClient)
	})

	t.Run("url error", func(t *testing.T) {
		apic, err := NewAPICClient(context.Background(), key, &http.Client{}, gsmClient)
		require.NoError(t, err)
		_, err = apic.Do(context.Background(), NewRequest(http.MethodPost, "%%", []byte{}), "test")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unable to create request")
	})

	t.Run("want: error when key passed in is empty string", func(t *testing.T) {
		apicClient, err := NewAPICClient(context.Background(), "", &http.Client{}, gsmClient)
		require.Error(t, err)
		assert.Nil(t, apicClient)
	})

	t.Run("want: error when gsm returns empty string for secret", func(t *testing.T) {
		apicClient, err := NewAPICClient(context.Background(), key, &http.Client{}, gsmNoSecretReturned)
		require.Error(t, err)
		assert.Nil(t, apicClient)
	})

	tests := []struct {
		name           string
		context        context.Context
		wantErr        string
		want           []byte
		requestHandler http.HandlerFunc
	}{
		{
			name:    "want: error when downstream service errors",
			context: testutil.GetContext(true),
			want:    nil,
			wantErr: "unexpected response from downstream",
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			},
		},
		{
			name:    "successfully get response",
			context: testutil.GetContext(true),
			want:    []byte("haha"),
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte("haha"))
			},
		},
		{
			name:    "want: error when no auth header",
			context: testutil.GetContext(false),
			wantErr: "failed to extract auth token",
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte("haha"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.requestHandler)
			// Close the server when test finishes
			defer server.Close()
			apic, err := NewAPICClient(tt.context, key, server.Client(), gsmClient)
			require.NoError(t, err)

			got, err := apic.Do(tt.context, NewRequest(http.MethodPost, server.URL, []byte{}), "test")
			if tt.wantErr != "" {
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCodeFromHTTPStatus(t *testing.T) {
	tests := []struct {
		want codes.Code
		in   int
	}{
		{
			in:   http.StatusOK,
			want: codes.OK,
		},
		{
			in:   http.StatusRequestTimeout,
			want: codes.Canceled,
		},
		{
			in:   http.StatusBadRequest,
			want: codes.InvalidArgument,
		},
		{
			in:   http.StatusGatewayTimeout,
			want: codes.DeadlineExceeded,
		},
		{
			in:   http.StatusNotFound,
			want: codes.NotFound,
		},
		{
			in:   http.StatusConflict,
			want: codes.AlreadyExists,
		},
		{
			in:   http.StatusForbidden,
			want: codes.PermissionDenied,
		},
		{
			in:   http.StatusUnauthorized,
			want: codes.Unauthenticated,
		},
		{
			in:   http.StatusTooManyRequests,
			want: codes.ResourceExhausted,
		},
		{
			in:   http.StatusNotImplemented,
			want: codes.Unimplemented,
		},
		{
			in:   http.StatusInternalServerError,
			want: codes.Internal,
		},
		{
			in:   http.StatusServiceUnavailable,
			want: codes.Unavailable,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v == %v", test.in, test.want), func(t *testing.T) {
			got := CodeFromHTTPStatus(test.in)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestRequestAddCustomHeaders(t *testing.T) {
	in := map[string]string{
		"never": "gonna",
		"give":  "you",
		"up":    "never",
		"gonna": "let",
		"you":   "down",
	}
	r := &Request{
		Headers: in,
	}
	got := &http.Request{}
	r.addCustomHeaders(got)

	for key, value := range in {
		assert.Equal(t, value, got.Header.Get(key))
	}
}

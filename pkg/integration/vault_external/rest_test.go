package vault_external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVaultAPI_Run(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		status        int
		address       string
		path          string
		method        string
		errMessage    string
		expectedError string
	}{
		{
			name:     "happy path",
			response: "foobar",
			address:  "http://foo.local",
			status:   200,
			method:   "GET",
		},
		{
			name:          "request 404",
			status:        404,
			method:        "PUT",
			address:       "http://foo.local",
			expectedError: "fabric error: status_code=NotFound, error_code=2, message=vault API request failed, reason=vault API response status 404",
		},
		{
			name:          "request 500",
			status:        500,
			method:        "POST",
			address:       "http://foo.local",
			expectedError: "fabric error: status_code=Internal, error_code=2, message=vault API request failed, reason=vault API response status 500",
		},
		{
			name:          "fails with invalid path",
			method:        "PUT",
			path:          "%20%",
			expectedError: "fabric error: status_code=Internal, error_code=0, message=vault API request failed, reason=failed to create request",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				response.WriteHeader(test.status)
				if test.errMessage != "" {
					_, _ = response.Write([]byte(test.errMessage))
				} else {
					_, _ = response.Write([]byte(test.response))
				}
			}))

			v := VaultAPI{
				httpClient: server.Client(),
				address:    fmt.Sprintf("%s%s", server.URL, test.path),
			}

			resp, err := v.run(ctx, test.method, test.path, "abc", []byte(""))
			respString := string(resp)

			if test.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.response, respString)
			}
		})
	}
}

func TestVaultAPI_VaultLogin(t *testing.T) {
	validResponse := Secret{
		Auth: &SecretAuth{
			ClientToken: "ok",
		},
	}
	validResponseBytes, _ := json.Marshal(validResponse)
	tests := []struct {
		name          string
		httpResponse  []byte
		httpStatus    int
		expectedError string
	}{
		{
			name:         "happy path",
			httpResponse: validResponseBytes,
		},
		{
			name:          "invalid response",
			httpResponse:  []byte("abcde"),
			expectedError: "fabric error: status_code=Internal, error_code=2, message=failed to get auth from vault login, reason=failed to unmarshal JSON: invalid character 'a' looking for beginning of value",
		},
		{
			name:          "http error",
			httpStatus:    500,
			expectedError: "fabric error: status_code=Internal, error_code=2, message=vault API request failed, reason=vault API response status 500",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if test.httpStatus > 0 {
					response.WriteHeader(test.httpStatus)
				}
				_, _ = response.Write(test.httpResponse)
			}))

			v := VaultAPI{
				httpClient: server.Client(),
				address:    fmt.Sprintf("%s%s", server.URL, authLoginPath),
			}
			_, err := v.vaultLogin(context.Background(), "foo", ".")
			if test.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

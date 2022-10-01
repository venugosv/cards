package vault_external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	secret := Secret{
		Auth: &SecretAuth{
			ClientToken: "foobar",
		},
	}
	secretBytes, _ := json.Marshal(secret)
	tests := []struct {
		name          string
		localToken    string
		email         string
		address       string
		httpResponse  []byte
		httpError     error
		expectedError string
	}{
		{
			name:         "happy",
			localToken:   "foo",
			email:        "fabric@anz.com",
			httpResponse: secretBytes,
		},
		{
			name:         "with real jwt",
			email:        "fabric@anz.com",
			httpResponse: secretBytes,
		},
		{
			name:          "with invalid login",
			email:         "fabric@anz.com",
			httpResponse:  []byte("abcde"),
			expectedError: "fabric error: status_code=Internal, error_code=2, message=failed to get auth from vault login, reason=failed to unmarshal JSON: invalid character 'a' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if test.httpError != nil {
					response.WriteHeader(500)
				}
				_, _ = response.Write(test.httpResponse)
			}))

			config := &Config{
				LocalToken:                test.localToken,
				Address:                   server.URL,
				NoGoogleCredentialsClient: true,
				OverrideServiceEmail:      test.email,
				TokenLifetime:             1 * time.Hour,
				TokenErrorRetryFirstTime:  10 * time.Minute,
				TokenErrorRetryMaxTime:    100 * time.Minute,
			}

			client, err := NewClient(context.Background(), server.Client(), config)
			if test.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, client.config)
				require.NotNil(t, client.metadataHttpClient)
				require.NotNil(t, client.jwtSigner)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expectedError, err.Error())
			}
		})
	}
}

func TestNewClient_WithNilConfig(t *testing.T) {
	_, err := NewClient(context.Background(), &http.Client{}, nil)
	require.Error(t, err)
}

func TestNewClient_WithNilHttp(t *testing.T) {
	_, err := NewClient(context.Background(), nil, &Config{
		LocalToken:               "foobar",
		Address:                  "somewhere",
		TokenLifetime:            1 * time.Hour,
		TokenErrorRetryFirstTime: 10 * time.Minute,
		TokenErrorRetryMaxTime:   100 * time.Minute,
	})
	require.Error(t, err)
}

func TestNewClient_HasValidBackoff(t *testing.T) {
	secret := Secret{
		Auth: &SecretAuth{
			ClientToken: "foobar",
		},
	}
	secretBytes, _ := json.Marshal(secret)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write(secretBytes)
	}))
	c, err := NewClient(context.Background(), server.Client(), &Config{
		LocalToken:                "foobar",
		Address:                   server.URL,
		OverrideServiceEmail:      "foo@local",
		NoGoogleCredentialsClient: true,
		TokenLifetime:             1 * time.Hour,
		TokenErrorRetryFirstTime:  10 * time.Minute,
		TokenErrorRetryMaxTime:    100 * time.Minute,
	})
	require.NoError(t, err)

	require.NotEqual(t, time.Duration(0), c.backoff.NextBackOff())
}

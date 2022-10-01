package vault

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/vault_external"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FakeVault struct {
	fixedError    string
	fixedResponse string
}

func (f *FakeVault) Transform(_ context.Context, _ vault_external.TransformKind, _ string, values []*vault_external.TransformRequest) ([]*vault_external.TransformResult, error) {
	if f.fixedError != "" {
		return nil, errors.New(f.fixedError)
	}
	if f.fixedResponse != "" {
		return []*vault_external.TransformResult{
			{
				DecodedValue: f.fixedResponse,
				EncodedValue: f.fixedResponse,
				Reference:    "foobarbaz",
			},
		}, nil
	}
	var response []*vault_external.TransformResult
	for _, v := range values {
		rv := vault_external.TransformResult{
			DecodedValue: v.Value,
			EncodedValue: v.Value,
			Reference:    v.Reference,
		}
		response = append(response, &rv)
	}
	return response, nil
}

func TestNewClient_WithNilHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		auth := vault_external.Secret{
			Auth: &vault_external.SecretAuth{
				ClientToken:   "foo",
				LeaseDuration: 100,
			},
		}
		b, _ := json.Marshal(auth)
		_, _ = rw.Write(b)
	}))
	config := &vault_external.Config{
		NoGoogleCredentialsClient: true,
		OverrideServiceEmail:      "foo@local",
		Address:                   server.URL,
	}
	_, err := NewClient(context.Background(), nil, config)
	require.NoError(t, err)
}

// We quietly return nothing when no config is provided, so assert this in a test
func TestNewClient_WithNilConfig(t *testing.T) {
	client, err := NewClient(context.Background(), nil, nil)
	require.NoError(t, err)
	require.Nil(t, client)
}

func TestClient_EncodeCardNumbers(t *testing.T) {
	tests := []struct {
		name           string
		cardNumbers    []string
		transformError string
		expectedError  string
	}{
		{
			name:        "happy path",
			cardNumbers: []string{"123"},
		},
		{
			name:           "error calling transform",
			transformError: "bad",
			expectedError:  "fabric error: status_code=Unknown, error_code=20004, message=failed to ",
			cardNumbers:    []string{"123"},
		},
		{
			name:          "no card numbers",
			cardNumbers:   []string{},
			expectedError: "fabric error: status_code=Internal, error_code=20004, message=failed to ",
		},
		{
			name:          "reject empty strings",
			cardNumbers:   []string{"123", ""},
			expectedError: "fabric error: status_code=Internal, error_code=20004, message=failed to ",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			externalVaultClient := FakeVault{
				fixedError: test.transformError,
			}
			cc := client{
				Client: &externalVaultClient,
			}
			encoded, err := cc.EncodeCardNumbers(context.Background(), test.cardNumbers)
			if test.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				for _, k := range test.cardNumbers {
					_, ok := encoded[k]
					require.True(t, ok, "card number %v not in encode response", k)
				}
			}
			decoded, err := cc.DecodeCardNumbers(context.Background(), test.cardNumbers)
			if test.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				for _, k := range test.cardNumbers {
					_, ok := decoded[k]
					require.True(t, ok, "card number %v not in decode response", k)
				}
			}
		})
	}
}

func TestClient_EncodeCardNumber(t *testing.T) {
	tests := []struct {
		name           string
		cardNumber     string
		transformError string
		expectedError  string
		fixedResponse  string
	}{
		{
			name:       "happy path",
			cardNumber: "123",
		},
		{
			name:           "error calling transform",
			transformError: "bad",
			expectedError:  "fabric error: status_code=Unknown, error_code=20004, message=failed to ",
			cardNumber:     "123",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			externalVaultClient := FakeVault{
				fixedError:    test.transformError,
				fixedResponse: test.fixedResponse,
			}
			cc := client{
				Client: &externalVaultClient,
			}
			encoded, err := cc.EncodeCardNumber(context.Background(), test.cardNumber)
			if test.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.cardNumber, encoded)
			}
			decoded, err := cc.DecodeCardNumber(context.Background(), test.cardNumber)
			if test.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.cardNumber, decoded)
			}
		})
	}
}

const (
	token      = "abcde"
	cardNumber = "12345"
)

func TestMapResults(t *testing.T) {
	tests := []struct {
		name    string
		request []string
		data    *vault_external.BatchTransformResults
		want    map[string]string
		wantErr string
	}{
		{
			name:    "token correctly mapped to card number",
			request: []string{token},
			data: &vault_external.BatchTransformResults{
				Results: []*vault_external.TransformResult{
					{DecodedValue: cardNumber},
				},
			},
			want: map[string]string{token: cardNumber},
		},
		{
			name:    "card number correctly mapped to token",
			request: []string{cardNumber},
			data: &vault_external.BatchTransformResults{
				Results: []*vault_external.TransformResult{
					{EncodedValue: token},
				},
			},
			want: map[string]string{cardNumber: token},
		},
		{
			name:    "error in response contains error",
			request: []string{"123", "123", "123"},
			data: &vault_external.BatchTransformResults{
				Results: []*vault_external.TransformResult{
					{EncodedValue: token},
					{
						EncodedValue: token,
						Errors:       "foobar",
					},
					{EncodedValue: token},
				},
			},
			wantErr: "fabric error: status_code=Internal, error_code=20004, message=failed to tokenize card numbers, reason=error in transform response: foobar",
		},
		{
			name:    "error when lengths don't match",
			request: []string{"123"},
			data: &vault_external.BatchTransformResults{
				Results: []*vault_external.TransformResult{
					{EncodedValue: token},
					{EncodedValue: token},
				},
			},
			wantErr: "fabric error: status_code=Internal, error_code=20004, message=failed to tokenize card numbers, reason=cannot map card numbers with transform response, lengths do not match",
		},
		{
			name:    "error when results are empty",
			request: []string{"123"},
			data: &vault_external.BatchTransformResults{
				Results: []*vault_external.TransformResult{
					{},
				},
			},
			wantErr: "fabric error: status_code=Internal, error_code=20004, message=unable to parse vault request string, reason=failed to map vault transform response values",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got, err := mapResults(context.Background(), test.request, test.data.Results)
			if test.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.wantErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

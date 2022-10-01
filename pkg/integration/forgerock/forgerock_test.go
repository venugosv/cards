package forgerock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	anzerrors "github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/metadata"

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

func TestNewClient(t *testing.T) {
	key := "test_client_secret_key"
	secret := "test_client_secret"
	gsmClient := &gsm.Client{
		SM: mockSecretManager{
			name:    key,
			payload: secret,
		},
	}
	t.Run("success", func(t *testing.T) {
		id := "test_client_id"
		require.NoError(t, os.Setenv(key, secret))
		defer os.Unsetenv(key)
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, "{}")
		}))
		config := &Config{
			BaseURL:         httpServer.URL,
			ClientID:        id,
			ClientSecretKey: key,
		}
		c, err := ClientFromConfig(context.Background(), httpServer.Client(), config, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})
	t.Run("null client secret", func(t *testing.T) {
		bad_key := "fake_key"
		config := &Config{
			ClientSecretKey: bad_key,
		}
		gsm := &gsm.Client{
			SM: mockSecretManager{
				err: fmt.Errorf("oh no"),
			},
		}
		c, err := ClientFromConfig(context.Background(), nil, config, gsm)
		require.Error(t, err)
		assert.Nil(t, c)
		assert.EqualError(t, err, "fabric error: status_code=Internal, error_code=1, message=failed to create forgerock client, reason=unable to find client secret")
	})
	t.Run("nil config", func(t *testing.T) {
		c, err := ClientFromConfig(context.Background(), nil, nil, gsmClient)
		assert.Nil(t, c)
		assert.Nil(t, err)
	})
}

func TestClient_GetToken(t *testing.T) {
	clientID := "test_client_id"
	clientSecret := "test_client_secret"
	tokenURL := "/token"
	const xRequestID = "x-request-id"

	type requestArgs struct {
		ctx   context.Context
		scope string
	}

	tests := []struct {
		description string
		request     requestArgs
		expectedErr error
		expectedRes *TokenResp
	}{
		{
			description: "success",
			request: requestArgs{
				ctx:   context.Background(),
				scope: "4567",
			},
			expectedErr: nil,
			expectedRes: &TokenResp{AccessToken: "1234", TokenType: "SYSTEM", ExpiresIn: 300},
		},
		{
			description: "x-request-id on context",
			request: requestArgs{
				ctx:   metadata.AppendToOutgoingContext(context.Background(), xRequestID, "x12345"),
				scope: "4567",
			},
			expectedErr: nil,
			expectedRes: &TokenResp{AccessToken: "1234", TokenType: "SYSTEM", ExpiresIn: 300},
		},
		{
			description: "success on empty scope",
			request: requestArgs{
				ctx: context.Background(),
			},
			expectedErr: nil,
			expectedRes: &TokenResp{AccessToken: "1234", TokenType: "SYSTEM", ExpiresIn: 300},
		},
		{
			description: "on client error",
			request: requestArgs{
				ctx:   metadata.AppendToOutgoingContext(context.Background(), xRequestID, "x12345"),
				scope: "4567",
			},
			expectedErr: anzerrors.Wrap(fmt.Errorf("received status code 400 from downstream service with body Bad request"), codes.Internal, failure,
				anzerrors.NewErrorInfo(context.Background(), errcodes.Unknown, downstreamFailure)),
			expectedRes: nil,
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.expectedErr != nil {
					http.Error(w, "Bad request", 400)
					return
				}

				assert.Equal(t, tokenURL, r.URL.Path)

				incoming, ok := metadata.FromIncomingContext(test.request.ctx)
				if !ok {
					assert.True(t, r.Header.Get(xRequestID) != "")
				} else {
					assert.Equal(t, incoming[xRequestID], r.Header.Get(xRequestID))
				}

				_ = r.ParseForm()
				assert.Equal(t, clientID, r.Form.Get("client_id"))
				assert.Equal(t, clientSecret, r.Form.Get("client_secret"))
				assert.Equal(t, test.request.scope, r.Form.Get("scope"))

				res, _ := json.Marshal(test.expectedRes)
				_, _ = w.Write(res)
			}))

			c := Client{
				clientURL:    httpServer.URL,
				clientID:     clientID,
				clientSecret: clientSecret,
				client:       httpServer.Client(),
			}

			got, err := c.SystemJWT(test.request.ctx, test.request.scope)
			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
				if exp := errors.Unwrap(test.expectedErr); exp != nil {
					assert.EqualError(t, errors.Unwrap(err), exp.Error())
				} else {
					assert.NoError(t, errors.Unwrap(err))
				}
			} else {
				md, _ := metadata.FromIncomingContext(got)
				token := md.Get("authorization")
				want := fmt.Sprintf("Bearer %s", test.expectedRes.AccessToken)
				assert.Equal(t, want, token[0])
				assert.NoError(t, err)
			}
		})
	}
}

package ctm

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/stretchr/testify/require"

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

func gsmClient() *gsm.Client {
	return &gsm.Client{
		SM: mockSecretManager{
			name:    "name",
			payload: "redispassword",
		},
	}
}

func TestClientFromConfig(t *testing.T) {
	gsmClient := gsmClient()
	key := "ClientIDKey"

	t.Run("New Client with httpClient supplied", func(t *testing.T) {
		server := httptest.NewServer(nil)
		got, err := ClientFromConfig(context.Background(), server.Client(), &Config{BaseURL: "localhost:8000", ClientIDEnvKey: key}, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("New Client without httpClient supplied", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, &Config{BaseURL: "localhost:8000", ClientIDEnvKey: key}, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("New Client without config supplied", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, nil, nil)
		require.Nil(t, err)
		require.Nil(t, got)
	})
	t.Run("no endpoint provided in config", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "%",
			ClientIDEnvKey: key,
		}
		got, err := ClientFromConfig(context.Background(), server.Client(), config, nil)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

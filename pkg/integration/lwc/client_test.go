package lwc

import (
	"context"
	"testing"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type mockSecretManagerServiceClient struct {
	accessSecretVersionFunc func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error)
}

func (m mockSecretManagerServiceClient) AccessSecretVersion(ctx context.Context, req *smpb.AccessSecretVersionRequest, _ ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
	return m.accessSecretVersionFunc(ctx, req)
}

func TestNewClient(t *testing.T) {
	const invalidURL = "%%"
	tests := []struct {
		name    string
		cfg     *Config
		sm      gsm.SecretManager
		wantErr string
	}{
		{
			name: "New LWC Client with httpClient supplied",
			cfg: &Config{
				"",
				&ProxyConfig{},
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return &smpb.AccessSecretVersionResponse{
						Name: req.GetName(),
						Payload: &smpb.SecretPayload{
							Data: []byte(`password`),
						},
					}, nil
				},
			},
		},
		{
			name: "New LWC Client without httpClient supplied",
			cfg: &Config{
				"",
				&ProxyConfig{},
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return &smpb.AccessSecretVersionResponse{
						Name: req.GetName(),
						Payload: &smpb.SecretPayload{
							Data: []byte(`password`),
						},
					}, nil
				},
			},
		},
		{
			name: "handle nil config being passed in",
			cfg:  nil,
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return &smpb.AccessSecretVersionResponse{
						Name: req.GetName(),
						Payload: &smpb.SecretPayload{
							Data: []byte(`password`),
						},
					}, nil
				},
			},
		},
		{
			name: "handle nil proxy config being passed in",
			cfg: &Config{
				"path/",
				nil,
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return &smpb.AccessSecretVersionResponse{
						Name: req.GetName(),
						Payload: &smpb.SecretPayload{
							Data: []byte(`password`),
						},
					}, nil
				},
			},
		},
		{
			name: "unable to access secret from GSM",
			cfg: &Config{
				"path/",
				&ProxyConfig{},
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return nil, errors.New("unable to fetch secretID")
				},
			},
			wantErr: "unable to access secret: failed to access secret : unable to fetch secretID",
		},
		{
			name: "handle incorrect base URL",
			cfg: &Config{
				invalidURL,
				&ProxyConfig{},
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return &smpb.AccessSecretVersionResponse{
						Name: req.GetName(),
						Payload: &smpb.SecretPayload{
							Data: []byte(`password`),
						},
					}, nil
				},
			},
			wantErr: "failed to configure proxy URL: parse \"%%\": invalid URL escape \"%%\"",
		},
		{
			name: "handle invalid proxy URL",
			cfg: &Config{
				"/path",
				&ProxyConfig{
					Username: "admin",
					Host:     invalidURL,
				},
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return &smpb.AccessSecretVersionResponse{
						Name: req.GetName(),
						Payload: &smpb.SecretPayload{
							Data: []byte(`password`),
						},
					}, nil
				},
			},
			wantErr: "failed to configure proxy URL: parse \"http://admin:password@%%\": invalid URL escape \"%%\"",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got, err := NewClient(context.Background(), test.cfg, nil, gsm.Client{SM: test.sm})
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				if test.cfg == nil {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}

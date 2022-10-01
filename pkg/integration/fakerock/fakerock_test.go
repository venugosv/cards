package fakerock

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/test/util/bufconn"

	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/metadata"

	frpb "github.com/anzx/fabricapis/pkg/fabric/service/fakerock/v1alpha1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
)

const (
	token        = "token"
	auth         = "auth"
	clientID     = "FAKEROCK_CLIENT_ID"
	clientSecret = clientID
)

type mockFakerockClient struct {
	jWKSFunc        func(ctx context.Context, in *frpb.JWKSRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error)
	loginFunc       func(ctx context.Context, in *frpb.LoginRequest, opts ...grpc.CallOption) (*frpb.LoginResponse, error)
	systemLoginFunc func(ctx context.Context, in *frpb.SystemLoginRequest, opts ...grpc.CallOption) (*frpb.SystemLoginResponse, error)
}

func (m mockFakerockClient) JWKS(ctx context.Context, in *frpb.JWKSRequest, _ ...grpc.CallOption) (*httpbody.HttpBody, error) {
	return m.jWKSFunc(ctx, in)
}

func (m mockFakerockClient) Login(ctx context.Context, in *frpb.LoginRequest, _ ...grpc.CallOption) (*frpb.LoginResponse, error) {
	return m.loginFunc(ctx, in)
}

func (m mockFakerockClient) SystemLogin(ctx context.Context, in *frpb.SystemLoginRequest, _ ...grpc.CallOption) (*frpb.SystemLoginResponse, error) {
	return m.systemLoginFunc(ctx, in)
}

type mockFakerockServer struct {
	frpb.UnimplementedFakerockAPIServer
	jwksFunc        func(context.Context, *frpb.JWKSRequest) (*httpbody.HttpBody, error)
	loginFunc       func(context.Context, *frpb.LoginRequest) (*frpb.LoginResponse, error)
	systemLoginFunc func(context.Context, *frpb.SystemLoginRequest) (*frpb.SystemLoginResponse, error)
}

func (m mockFakerockServer) JWKS(ctx context.Context, in *frpb.JWKSRequest) (*httpbody.HttpBody, error) {
	return m.jwksFunc(ctx, in)
}

func (m mockFakerockServer) Login(ctx context.Context, in *frpb.LoginRequest) (*frpb.LoginResponse, error) {
	return m.loginFunc(ctx, in)
}

func (m mockFakerockServer) SystemLogin(ctx context.Context, in *frpb.SystemLoginRequest) (*frpb.SystemLoginResponse, error) {
	return m.systemLoginFunc(ctx, in)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		setEnv      bool
		errContents string
	}{
		{
			name:        "invalid config",
			config:      &Config{BaseURL: "%%"},
			errContents: "fabric error: status_code=Internal, error_code=1, message=failed to create fakerock adapter, reason=unable to parse configured url",
		},
		{
			name: "nil config",
		},
		{
			name: "Valid config",
			config: &Config{
				BaseURL:         "http://bufconn",
				ClientSecretKey: clientID,
			},
			setEnv: true,
		},
		{
			name: "Valid config without clientID set",
			config: &Config{
				BaseURL:         "http://bufconn",
				ClientSecretKey: clientID,
			},
			setEnv:      false,
			errContents: "fabric error: status_code=Internal, error_code=1, message=failed to create fakerock adapter, reason=unable to find clientSecret",
		},
		{
			name: "unable to make successful connection",
			config: &Config{
				BaseURL:         "https://bufconn",
				ClientSecretKey: clientID,
			},
			setEnv:      true,
			errContents: "fabric error: status_code=Unavailable, error_code=2, message=failed to create fakerock adapter, reason=unable to make successful connection",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			l := bufconn.GetListener(func(server *grpc.Server) {
				frpb.RegisterFakerockAPIServer(server, mockFakerockServer{})
			})

			if test.setEnv {
				err := os.Setenv(clientID, clientSecret)
				require.NoError(t, err)
				defer os.Unsetenv(clientID)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			got, err := NewClient(ctx, test.config, grpc.WithContextDialer(l.BufDialer))
			if test.errContents != "" {
				require.Nil(t, got)
				require.Error(t, err)
				assert.EqualError(t, err, test.errContents)
			} else {
				require.NoError(t, err)
				if test.config == nil {
					assert.Nil(t, got)
				} else {
					require.NotNil(t, got)
				}
			}
		})
	}
}

func TestClient_ElevateContext(t *testing.T) {
	tests := []struct {
		name     string
		fakerock mockFakerockClient
		want     string
	}{
		{
			name: "happy path",
			fakerock: mockFakerockClient{
				systemLoginFunc: func(ctx context.Context, in *frpb.SystemLoginRequest, opts ...grpc.CallOption) (*frpb.SystemLoginResponse, error) {
					return &frpb.SystemLoginResponse{Token: token}, nil
				},
			},
			want: token,
		}, {
			name: "unhappy path",
			fakerock: mockFakerockClient{
				systemLoginFunc: func(ctx context.Context, in *frpb.SystemLoginRequest, opts ...grpc.CallOption) (*frpb.SystemLoginResponse, error) {
					return nil, fmt.Errorf("oh no!")
				},
			},
			want: "fabric error: status_code=Internal, error_code=2, message=failed to get token, reason=system login failed",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c := &Client{
				FakerockAPIClient: test.fakerock,
				basicAuth:         auth,
			}
			got, err := c.ElevateContext(context.Background())
			if err != nil {
				assert.EqualError(t, err, test.want)
			} else {
				md, _ := metadata.FromIncomingContext(got)
				gotToken := md.Get("authorization")
				want := fmt.Sprintf("Bearer %s", test.want)
				assert.Equal(t, want, gotToken[0])
			}
		})
	}
}

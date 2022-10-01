package entitlements

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/test/stubs/grpc/entitlements"

	"github.com/anzx/fabric-cards/test/util/bufconn"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
)

type mockEntitlementsClient struct {
	listCall func(ctx context.Context, in *entpb.ListEntitledCardsRequest, opts ...grpc.CallOption) (*entpb.ListEntitledCardsResponse, error)
	getCall  func(ctx context.Context, in *entpb.GetEntitledCardRequest, opts ...grpc.CallOption) (*entpb.EntitledCard, error)
}

func (m *mockEntitlementsClient) GetEntitledCard(ctx context.Context, in *entpb.GetEntitledCardRequest, opts ...grpc.CallOption) (*entpb.EntitledCard, error) {
	return m.getCall(ctx, in, opts...)
}

func (m *mockEntitlementsClient) ListEntitledCards(ctx context.Context, in *entpb.ListEntitledCardsRequest, opts ...grpc.CallOption) (*entpb.ListEntitledCardsResponse, error) {
	return m.listCall(ctx, in, opts...)
}

func (m *mockEntitlementsClient) ListPersonaForCardToken(ctx context.Context, _ *entpb.ListPersonaForCardTokenRequest, opts ...grpc.CallOption) (*entpb.ListPersonaForCardTokenResponse, error) {
	return nil, nil
}

func TestNewEntitlementsClient(t *testing.T) {
	tests := []struct {
		name          string
		input         *Config
		wantErr       string
		listenerClose bool
		serverClose   bool
	}{
		{
			name:    "invalid config",
			input:   &Config{BaseURL: "%%"},
			wantErr: "fabric error: status_code=Internal, error_code=1, message=failed to create entitlements adapter, reason=unable to parse configured url",
		},
		{
			name: "nil config",
		},
		{
			name: "Valid config",
			input: &Config{
				BaseURL: "localhost:9090",
			},
		},
		{
			name: "listener closed",
			input: &Config{
				BaseURL: "localhost:9090",
			},
			listenerClose: true,
			serverClose:   false,
			wantErr:       "fabric error: status_code=Unavailable, error_code=2, message=failed to create entitlements adapter, reason=unable to make successful connection",
		},
		{
			name: "server closed",
			input: &Config{
				BaseURL: "localhost:9090",
			},
			listenerClose: false,
			serverClose:   true,
			wantErr:       "fabric error: status_code=Unavailable, error_code=2, message=failed to create entitlements adapter, reason=unable to make successful connection",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			register := func(server *grpc.Server) {
				entpb.RegisterCardEntitlementsAPIServer(server, entitlements.StubCardEntitlementsAPIServer{})
			}

			listener := bufconn.GetListener(register)
			defer listener.Close()

			if test.listenerClose || test.serverClose {
				listener.Close()
			}

			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					return listener.Dial()
				}),
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			got, err := NewClient(ctx, tt.input, opts...)
			if test.wantErr != "" {
				require.Nil(t, got)
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				if test.input == nil {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}

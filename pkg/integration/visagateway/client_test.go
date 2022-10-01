package visagateway

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/cardonfile"
	cofpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/cardonfile"

	"google.golang.org/grpc/credentials/insecure"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"

	dcvv2pb "github.com/anzx/fabricapis/pkg/gateway/visa/service/dcvv2"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/dcvv2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/test/util/bufconn"

	"google.golang.org/grpc"
)

type mockDCVV2Server struct {
	GenerateFunc func(context.Context, *dcvv2pb.Request) (*dcvv2pb.Response, error)
}

func (m mockDCVV2Server) Generate(ctx context.Context, in *dcvv2pb.Request) (*dcvv2pb.Response, error) {
	return m.GenerateFunc(ctx, in)
}

type mockCustomerRulesServer struct {
	crpb.UnimplementedCustomerRulesAPIServer
}

type mockCardOnFileServer struct {
	cofpb.UnimplementedCardOnFileAPIServer
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		input         *Config
		wantErr       string
		listenerClose bool
		serverClose   bool
	}{
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
			wantErr:       "fabric error: status_code=Unavailable, error_code=2, message=failed to create visa gateway adapter, reason=unable to make successful connection",
		},
		{
			name: "server closed",
			input: &Config{
				BaseURL: "localhost:9090",
			},
			serverClose: true,
			wantErr:     "fabric error: status_code=Unavailable, error_code=2, message=failed to create visa gateway adapter, reason=unable to make successful connection",
		},
		{
			name:    "invalid config",
			input:   &Config{BaseURL: "%%"},
			wantErr: "fabric error: status_code=Internal, error_code=1, message=failed to create visa gateway adapter, reason=unable to parse configured url",
		},
		{
			name: "nil config",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			register := func(server *grpc.Server) {
				dcvv2pb.RegisterDCVV2APIServer(server, mockDCVV2Server{})
				crpb.RegisterCustomerRulesAPIServer(server, mockCustomerRulesServer{})
				cofpb.RegisterCardOnFileAPIServer(server, mockCardOnFileServer{})
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
				require.Error(t, err)

				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)

				if test.input == nil {
					require.Nil(t, got)
				} else {
					require.NotNil(t, got)
					assert.IsType(t, &Client{}, got)

					require.NotNil(t, got.DCVV2)
					assert.IsType(t, &dcvv2.Client{}, got.DCVV2)

					require.NotNil(t, got.CustomerRules)
					assert.IsType(t, &customerrules.Client{}, got.CustomerRules)

					require.NotNil(t, got.CardOnFile)
					assert.IsType(t, &cardonfile.Client{}, got.CardOnFile)
				}
			}
		})
	}
}

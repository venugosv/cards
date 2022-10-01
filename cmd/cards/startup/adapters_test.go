package startup

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"

	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/pkg/gsm"

	"github.com/googleapis/gax-go/v2"

	"github.com/anzx/fabric-cards/pkg/ratelimit"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/cmd/cards/config/app"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/echidna"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/ocv"
	"github.com/anzx/fabric-cards/pkg/integration/selfservice"
	"github.com/anzx/fabric-cards/pkg/integration/vault_external"
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-cards/pkg/middleware/logging"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/pkg/auditlog"

	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

const (
	baseURL    = "http://localhost:8080"
	validURL   = "localhost:8080"
	invalidURL = "%%"
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

func TestNewAdapters(t *testing.T) {
	local := sdk.EnvironmentLocal
	tests := []struct {
		name    string
		sm      gsm.SecretManager
		config  app.Spec
		wantErr error
	}{
		{
			name: "create payload interceptor",
			config: app.Spec{
				Log: logging.Config{
					PayloadLoggingDecider: grpclogging.PayloadLoggingDecider{
						Server: map[string]bool{
							"this": true,
						},
					},
				},
			},
		},
		{
			name: "create command centre adapter",
			config: app.Spec{
				CommandCentre: &commandcentre.Config{
					Env: &local,
				},
			},
		},
		{
			name: "create Eligibility adapter",
			config: app.Spec{
				Eligibility: &eligibility.Config{BaseURL: ""},
			},
		},
		{
			name: "create ctm adapter",
			config: app.Spec{
				CTM: &ctm.Config{
					ClientIDEnvKey: "CTMClientIDEnvKey",
				},
			},
			sm: mockSecretManager{
				name:    "testName",
				payload: "returned secret",
			},
		},
		{
			name: "create rateLimit adapter, return error",
			sm: mockSecretManager{
				name:    "testName",
				payload: "returned secret",
			},
			config: app.Spec{
				RateLimit: &ratelimit.Config{
					Redis: ratelimit.RedisConfig{
						Password: "redisPassword",
					},
				},
			},
			wantErr: errors.New("could not configure Rate Limit client"),
		},
		{
			name: "create ctm adapter",
			config: app.Spec{
				CTM: &ctm.Config{},
			},
			wantErr: errors.New("could not configure CTM Client with config &{BaseURL: ClientIDEnvKey: MaxRetries:0}"),
		},
		{
			name: "create echidna adapter",
			config: app.Spec{
				Echidna: &echidna.Config{
					ClientIDEnvKey: "EchidnaClientIDEnvKey",
				},
			},
			sm: mockSecretManager{
				name:    "testName",
				payload: "returned secret",
			},
		},
		{
			name: "create echidna adapter",
			config: app.Spec{
				Echidna: &echidna.Config{},
			},
			wantErr: errors.New("could not configure Echidna Client with config &{BaseURL: ClientIDEnvKey: ClientID: MaxRetries:0}"),
		},
		{
			name: "create vault adapter",
			config: app.Spec{
				Vault: &vault_external.Config{
					NoGoogleCredentialsClient: true,
					OverrideServiceEmail:      "foo@local",
				},
			},
			wantErr: errors.New("status_code=Internal, error_code=2, message=unable to create vault adapter, reason=error making HTTP request to vault API"),
		},
		{
			name: "create auditlog adapter",
			config: app.Spec{
				AuditLog: &auditlog.Config{},
			},
		},
		{
			name: "create ocv adapter",
			config: app.Spec{
				OCV: &ocv.Config{
					ClientIDEnvKey: "OCVClientIDEnvKey",
				},
			},
			sm: mockSecretManager{
				name:    "testName",
				payload: "returned secret",
			},
		},
		{
			name: "create ocv adapter",
			config: app.Spec{
				OCV: &ocv.Config{},
			},
			wantErr: errors.New("could not configure OCV Client with config &{BaseURL: ClientIDEnvKey: MaxRetries:0 EnableLogging:false}"),
		},
		{
			name: "successfully create adapters with only forgerock config supplied",
			config: app.Spec{
				Forgerock: &forgerock.Config{
					ClientSecretKey: "testName",
				},
			},
			sm: mockSecretManager{
				name:    "testName",
				payload: "returned secret",
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			gsmClient := &gsm.Client{
				SM: test.sm,
			}
			got, err := NewAdapters(context.Background(), test.config, gsmClient)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestNewAdaptersErrors(t *testing.T) {
	sit := sdk.EnvironmentSit
	tests := []struct {
		name   string
		config app.Spec
		errMsg string
	}{
		{
			name: "could not configure Eligibility Client",
			config: app.Spec{
				CommandCentre: &commandcentre.Config{
					Env: &sit,
				},
				Eligibility: &eligibility.Config{BaseURL: invalidURL},
			},
			errMsg: "could not configure Eligibility Client with config",
		},
		{
			name: "could not configure Entitlements client",
			config: app.Spec{
				CommandCentre: &commandcentre.Config{
					Env: &sit,
				},
				Eligibility:  &eligibility.Config{BaseURL: validURL},
				Entitlements: &entitlements.Config{BaseURL: invalidURL},
			},
			errMsg: "could not configure Entitlements client with config",
		},
		{
			name: "could not configure Self Service client",
			config: app.Spec{
				CommandCentre: &commandcentre.Config{
					Env: &sit,
				},
				Eligibility: &eligibility.Config{BaseURL: validURL},
				SelfService: &selfservice.Config{BaseURL: invalidURL},
			},
			errMsg: "could not configure Self Service client with config",
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAdapters(context.Background(), tt.config, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestNewAdaptersWithBlock(t *testing.T) {
	tests := []struct {
		name   string
		config app.Spec
	}{
		{
			name: "create entitlements adapter",
			config: app.Spec{
				Entitlements: &entitlements.Config{BaseURL: baseURL},
			},
		},
		{
			name: "create self service adapter",
			config: app.Spec{
				SelfService: &selfservice.Config{BaseURL: baseURL},
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			// Setup the grpc server
			svr := grpc.NewServer()
			defer svr.GracefulStop()
			// Listen for incoming connections
			listener, err := net.Listen("tcp", validURL)
			require.NoError(t, err)
			defer listener.Close()
			go func() {
				_ = svr.Serve(listener)
				svr.Stop()
			}()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			got, err := NewAdapters(ctx, test.config, nil)
			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

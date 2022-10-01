package startup

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/vault_external"
	"github.com/anzx/fabric-cards/pkg/integration/visagateway"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"

	"github.com/anzx/fabric-cards/pkg/integration/visa"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/pkg/auditlog"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/middleware/logging"

	"github.com/anzx/fabric-cards/cmd/cardcontrols/config/app"

	"github.com/anzx/fabric-cards/pkg/integration/eligibility"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/stretchr/testify/assert"
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

// Test to create adapters when config is supplied only, adapter creation is not blocking if config it not supplied.
func TestNewAdapters(t *testing.T) {
	local := sdk.EnvironmentLocal
	tests := []struct {
		name    string
		config  app.Spec
		wantErr error
	}{
		{
			name: "successfully create interceptors with only payload config supplied",
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
			name: "successfully create adapters with only command centre config supplied",
			config: app.Spec{
				CommandCentre: &commandcentre.Config{
					Env: &local,
				},
			},
		},
		{
			name: "successfully create adapters with only Eligibility config supplied",
			config: app.Spec{
				Eligibility: &eligibility.Config{BaseURL: ""},
			},
		},
		{
			name: "fail to create adapters with invalid ctm config supplied",
			config: app.Spec{
				CTM: &ctm.Config{},
			},
			wantErr: errors.New("could not configure CTM Client with config &{BaseURL: ClientIDEnvKey: MaxRetries:0}"),
		},
		{
			name: "successfully create adapters with only ctm config supplied",
			config: app.Spec{
				CTM: &ctm.Config{
					ClientIDEnvKey: "CTMClientIDEnvKey",
				},
			},
		},
		{
			name: "fail to create adapters with invalid visa config supplied",
			config: app.Spec{
				Visa: &visa.Config{},
			},
			wantErr: errors.New("could not configure Visa Client with config &{BaseURL: ClientIDEnvKey: ClientID: MaxRetries:0}"),
		},
		{
			name: "successfully create adapters with only visa config supplied",
			config: app.Spec{
				Visa: &visa.Config{
					ClientIDEnvKey: "VisaClientIDEnvKey",
				},
			},
		},
		{
			name: "successfully create adapters with only vault config supplied",
			config: app.Spec{
				Vault: &vault_external.Config{
					OverrideServiceEmail:      "foo@local",
					NoGoogleCredentialsClient: true,
				},
			},
			wantErr: errors.New("foo"),
		},
		{
			name: "successfully create adapters with only auditlog config supplied",
			config: app.Spec{
				AuditLog: &auditlog.Config{},
			},
		},
		{
			name: "successfully create adapters with only auditlog config supplied",
			config: app.Spec{
				AuditLog: &auditlog.Config{},
			},
		},
		{
			name: "successfully create adapters with only forgerock config supplied",
			config: app.Spec{
				Forgerock: &forgerock.Config{
					ClientSecretKey: "testName",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gsmClient := &gsm.Client{
				SM: mockSecretManager{
					name:    "testName",
					payload: "password",
				},
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
	tests := []struct {
		name   string
		config app.Spec
		errMsg string
	}{
		{
			name: "could not configure Eligibility Client",
			config: app.Spec{
				Eligibility: &eligibility.Config{BaseURL: invalidURL},
			},
			errMsg: "could not configure Eligibility Client with config",
		},
		{
			name: "could not configure Entitlements client",
			config: app.Spec{
				Entitlements: &entitlements.Config{BaseURL: invalidURL},
			},
			errMsg: "could not configure Entitlements client with config",
		},
		{
			name: "could not configure visa-gateway client",
			config: app.Spec{
				VisaGateway: &visagateway.Config{BaseURL: invalidURL},
			},
			errMsg: "unable to parse configured url",
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
			name: "create visa gateway adapter",
			config: app.Spec{
				VisaGateway: &visagateway.Config{BaseURL: baseURL},
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

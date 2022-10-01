package startup

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"

	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"

	"github.com/anzx/fabric-cards/pkg/middleware/logging"

	"github.com/anzx/fabric-cards/cmd/callback/config/app"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

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
		}, {
			name: "create command centre adapter",
			config: app.Spec{
				CommandCentre: &commandcentre.Config{
					Env: &local,
				},
			},
		}, {
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
		}, {
			name: "create ctm adapter",
			config: app.Spec{
				CTM: &ctm.Config{},
			},
			sm:      mockSecretManager{},
			wantErr: errors.New("could not configure CTM Client with config &{BaseURL: ClientIDEnvKey: MaxRetries:0}"),
		}, {
			name: "successfully create adapters with only forgerock config supplied",
			config: app.Spec{
				Forgerock: &forgerock.Config{
					ClientSecretKey: "ForgerockClientIDEnvKey",
				},
			},
			sm: mockSecretManager{
				name:    "ForgerockClientIDEnvKey",
				payload: "returned secret",
			},
		},
	}
	for _, test := range tests {
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

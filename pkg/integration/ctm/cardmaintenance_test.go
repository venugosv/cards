package ctm

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"github.com/anzx/fabric-cards/test/util"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/util/testutil"
	"github.com/stretchr/testify/assert"
)

const (
	firstName   = "Jane"
	lastName    = "Doe"
	clientIDKey = "ClientIDKey"
)

func TestClient_ReplaceCard(t *testing.T) {
	tests := []struct {
		name                string
		req                 *ReplaceCardRequest
		tokenizedCardNumber string
		want                string
		mockAPIc            apic.Clienter
		wantErr             error
	}{
		{
			name: "true returned on downstream 200",
			req: &ReplaceCardRequest{
				PlasticType:      NewNumber,
				FirstName:        firstName,
				LastName:         lastName,
				DispatchedMethod: DispatchedMethodMail,
			},
			tokenizedCardNumber: cardNumber,
			want:                "2222222222224466",
			mockAPIc:            util.MockAPIcer{Response: []byte(`{ "cardNumber": { "token": "2222222222224466", "last4digits": "4466" }}`)},
		},
		{
			name: "false returned on downstream 500",
			req: &ReplaceCardRequest{
				PlasticType:      NewNumber,
				FirstName:        firstName,
				LastName:         lastName,
				DispatchedMethod: DispatchedMethodMail,
			},
			tokenizedCardNumber: cardNumber,
			mockAPIc:            util.MockAPIcer{ResponseErr: errors.New("failed request")},
			wantErr:             errors.New("failed request"),
		},
		{
			name:                "handle unexpected body from downstream",
			tokenizedCardNumber: tokenizedCardNumber,
			wantErr:             errors.New("fabric error: status_code=Internal, error_code=2, message=failed request, reason=unexpected response from downstream"),
			mockAPIc:            util.MockAPIcer{Response: []byte(`%%`)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &client{
				apicClient: test.mockAPIc,
			}

			got, err := c.ReplaceCard(testutil.GetContext(true), test.req, test.tokenizedCardNumber)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func TestClient_ReplaceCardError(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("unable to access server", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: clientIDKey,
		}

		c, err := ClientFromConfig(context.Background(), server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.ReplaceCard(testutil.GetContext(true), &ReplaceCardRequest{}, cardNumber)
		require.Error(t, err)
		assert.Empty(t, got)
	})
	t.Run("no auth in header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		ctx := context.Background()
		config := &Config{
			BaseURL:        server.URL,
			ClientIDEnvKey: clientIDKey,
		}

		c, err := ClientFromConfig(ctx, server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.ReplaceCard(ctx, &ReplaceCardRequest{}, cardNumber)
		require.Error(t, err)
		assert.Empty(t, got)
	})
}

func TestClient_SetPreferences(t *testing.T) {
	tests := []struct {
		name     string
		want     bool
		mockAPIc apic.Clienter
		wantErr  error
	}{
		{
			name:     "true returned on downstream 200",
			want:     true,
			mockAPIc: util.MockAPIcer{},
		},
		{
			name:     "false returned on downstream 500",
			want:     false,
			wantErr:  errors.New("failed request"),
			mockAPIc: util.MockAPIcer{ResponseErr: errors.New("failed request")},
		},
		{
			name: "NO CHANGES MADE OR DETECTED",
			want: true,
			mockAPIc: util.MockAPIcer{
				Response:    []byte(`{"errors": [{"type": "Provider Error","message": "NO CHANGES MADE OR DETECTED","code": "909","location": "CTM","severity": "Fatal"}],"httpCode": "400","moreInformation": "CONDITION_CODE 1506004"}`),
				ResponseErr: anzerrors.New(codes.Internal, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from downstream")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := c.UpdatePreferences(testutil.GetContext(true), &UpdatePreferencesRequest{}, tokenizedCardNumber)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_SetPreferencesError(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("unable to access server", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: clientIDKey,
		}

		c, err := ClientFromConfig(context.Background(), server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.UpdatePreferences(testutil.GetContext(true), &UpdatePreferencesRequest{}, tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
	t.Run("no endpoint provided in config", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: clientIDKey,
		}
		c, err := ClientFromConfig(context.Background(), server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.UpdatePreferences(testutil.GetContext(true), &UpdatePreferencesRequest{}, tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
	t.Run("no auth in header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: clientIDKey,
		}
		ctx := context.Background()
		c, err := ClientFromConfig(ctx, server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.UpdatePreferences(ctx, &UpdatePreferencesRequest{}, tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
}

func TestClient_UpdateDetails(t *testing.T) {
	tests := []struct {
		name     string
		req      *UpdateDetailsRequest
		want     bool
		mockAPIc util.MockAPIcer
		wantErr  error
	}{
		{
			name:     "true returned on downstream 200",
			want:     true,
			mockAPIc: util.MockAPIcer{},
		},
		{
			name:     "false returned on downstream 500",
			want:     false,
			wantErr:  errors.New("failed request"),
			mockAPIc: util.MockAPIcer{ResponseErr: errors.New("failed request")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := c.UpdateDetails(testutil.GetContext(true), tt.req, tokenizedCardNumber)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_UpdateDetailsError(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("unable to access server", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: clientIDKey,
		}
		c, err := ClientFromConfig(context.Background(), server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.UpdateDetails(testutil.GetContext(true), &UpdateDetailsRequest{}, tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
	t.Run("no auth in header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{BaseURL: "localhost:8000", ClientIDEnvKey: clientIDKey}
		ctx := context.Background()
		c, err := ClientFromConfig(ctx, server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.UpdateDetails(ctx, &UpdateDetailsRequest{}, tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
}

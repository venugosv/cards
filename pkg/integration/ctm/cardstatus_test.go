package ctm

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	testUtil "github.com/anzx/fabric-cards/test/util"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/util/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClient_Activate(t *testing.T) {
	tests := []struct {
		name     string
		req      string
		want     bool
		wantErr  bool
		mockAPIc apic.Clienter
	}{
		{
			name:     "true returned on downstream 200",
			req:      tokenizedCardNumber,
			want:     true,
			mockAPIc: testUtil.MockAPIcer{},
		},
		{
			name:     "false returned on downstream 500",
			req:      tokenizedCardNumber,
			want:     false,
			wantErr:  true,
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("failed request")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := c.Activate(testutil.GetContext(true), tt.req)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_ActivateError(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("unable to access server", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: key,
		}
		c, err := ClientFromConfig(context.Background(), server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.Activate(testutil.GetContext(true), tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
	t.Run("no auth in header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			ClientIDEnvKey: key,
		}
		ctx := context.Background()

		c, err := ClientFromConfig(ctx, server.Client(), config, gsmClient)
		require.NoError(t, err)

		got, err := c.Activate(ctx, tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		want     bool
		wantErr  bool
		mockAPIc apic.Clienter
	}{
		{
			name:     "true returned on downstream 200",
			status:   StatusTemporaryBlock,
			want:     true,
			mockAPIc: testUtil.MockAPIcer{},
		},
		{
			name:     "false returned on downstream 500",
			status:   StatusTemporaryBlock,
			want:     false,
			wantErr:  true,
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("failed request")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := c.UpdateStatus(testutil.GetContext(true), tokenizedCardNumber, tt.status)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_UpdateError(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("unable to access server", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: key,
		}
		c, err := ClientFromConfig(context.Background(), server.Client(), config, gsmClient)
		require.NoError(t, err)
		got, err := c.UpdateStatus(testutil.GetContext(true), tokenizedCardNumber, StatusTemporaryBlock)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
	t.Run("no auth in header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        "http://apisit03.service.anz",
			ClientIDEnvKey: key,
		}
		ctx := context.Background()

		c, err := ClientFromConfig(ctx, server.Client(), config, gsmClient)
		require.NoError(t, err)

		got, err := c.UpdateStatus(ctx, tokenizedCardNumber, StatusTemporaryBlock)
		assert.NotNil(t, err)
		assert.False(t, got)
	})
}

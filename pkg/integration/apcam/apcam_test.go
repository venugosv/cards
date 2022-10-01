package apcam

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"github.com/anzx/fabric-cards/pkg/util/testutil"
	testUtil "github.com/anzx/fabric-cards/test/util"
	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getResponse() *Response {
	return &Response{
		TraceInfo: TraceInfo{
			MessageID:      "2790503e-5208-11ea-9d06-cd011c010000",
			ConversationID: "2790503e-5208-11ea-9d06-cd011c010000",
		},
		Apple: AppleResponseData{
			EncryptedPassData:  "TUJQQUQtMS1GSy0xMjM0NTYuMS0tVERFQS03QUYyOTFDOTFGM0VENEVGOTJDMUQ0NUVGRjEyN0MxRjlBQkMxMjM0N0U=",
			ActivationData:     "QUJDREVGLTEtRkstMTIzNDU2LjEtLVRERUEtN0FGMjkxQzkxRjNFRDRFRjkyQzFENDVFRkYxMjdDMUY5QUJDMTIzNDdF",
			EphemeralPublicKey: "UVVKRFJFVkdMVEV0UmtzdE1USXpORFUyTGpFdExWUkVSVUV0TjBGR01qa3hRemt4UmpORlJEUkZSamt5UXpGRU5EVk",
		},
	}
}

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

func TestClientFromConfig(t *testing.T) {
	key := "APCamKey"
	tests := []struct {
		name       string
		httpClient *http.Client
		sm         mockSecretManager
		config     *Config
		wantErr    bool
		wantClient bool
	}{
		{
			name:       "New APCam with httpClient supplied",
			httpClient: httptest.NewServer(nil).Client(),
			sm: mockSecretManager{
				name:    "testName",
				payload: "secret",
			},
			config:     &Config{ClientIDEnvKey: key},
			wantClient: true,
		},
		{
			name:       "New APCam without httpClient supplied",
			httpClient: nil,
			sm: mockSecretManager{
				name:    "testName",
				payload: "secret",
			},
			config:     &Config{ClientIDEnvKey: key},
			wantClient: true,
		},
		{
			name:       "New APCam with bad baseURL",
			httpClient: nil,
			config: &Config{
				ClientIDEnvKey: key,
				BaseURL:        "%%",
			},
			wantErr: true,
		},
		{
			name:       "New APCam with no config",
			httpClient: nil,
			config:     nil,
			wantErr:    false,
			wantClient: false,
		},
		{
			name:       "New APCam with no client ID found",
			httpClient: nil,
			config:     &Config{ClientIDEnvKey: key},
			sm: mockSecretManager{
				name:    "testName",
				payload: "",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gsmClient := &gsm.Client{SM: test.sm}

			got, err := ClientFromConfig(context.Background(), test.httpClient, test.config, gsmClient)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if test.wantClient {
					assert.NotNil(t, got)
				} else {
					assert.Nil(t, got)
				}
			}
		})
	}
}

func TestClient_PushProvision(t *testing.T) {
	successfulResponseData, _ := json.Marshal(getResponse())

	tests := []struct {
		name           string
		request        Request
		want           *Response
		wantErr        string
		mockAPIc       apic.Clienter
		requestHandler http.HandlerFunc
	}{
		{
			name: "successfully push provision",
			request: Request{
				TraceInfo: TraceInfo{},
				CardInfo:  CardInfo{},
				Apple:     Apple{},
				UUID:      "",
			},
			want:     getResponse(),
			mockAPIc: testUtil.MockAPIcer{Response: successfulResponseData},
		},
		{
			name: "fail to unmarshall body",
			request: Request{
				TraceInfo: TraceInfo{},
				CardInfo:  CardInfo{},
				Apple:     Apple{},
				UUID:      "",
			},
			mockAPIc: testUtil.MockAPIcer{Response: []byte{32}},
			wantErr:  "unexpected response from downstream",
		},
		{
			name:     "handle 401",
			request:  Request{},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed push provision request, reason=unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{Response: []byte(`{Incorrect data}`)},
		},
		{
			name:     "handle 400",
			request:  Request{},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed push provision request, reason=unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("bad")},
		},
	}
	for _, test := range tests {
		apcam := &apcam{
			apicClient: test.mockAPIc,
		}

		t.Run(test.name, func(t *testing.T) {
			got, err := apcam.PushProvision(testutil.GetContext(true), test.request)

			if test.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

package visa

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/util/testutil"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	"github.com/pkg/errors"

	testUtil "github.com/anzx/fabric-cards/test/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/stretchr/testify/assert"
)

const (
	testDocumentID = "ctc-vd-857a8766-160b-498d-820f-cfd333399c22"
	key            = "ClientIDKey"
)

var updateRequest = &Request{
	TransactionControls: []*TransactionControl{
		{
			ShouldDeclineAll:     false,
			DeclineThreshold:     util.ToFloat64Ptr(200),
			ControlEnabled:       true,
			ControlType:          ccpb.ControlType_TCT_E_COMMERCE.String(),
			UserIdentifier:       "abhi-539d-4f93-ba00-77ef9ff873a2",
			AlertThreshold:       util.ToFloat64Ptr(10),
			ShouldAlertOnDecline: false,
		},
		{
			ShouldDeclineAll:     false,
			DeclineThreshold:     util.ToFloat64Ptr(200),
			ControlEnabled:       true,
			ControlType:          ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
			UserIdentifier:       "abhi-539d-4f93-ba00-77ef9ff873a2",
			AlertThreshold:       util.ToFloat64Ptr(10),
			ShouldAlertOnDecline: false,
		},
	},
}

type mockSecretManager struct {
	accessSecretVersionFunc func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

func (m mockSecretManager) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return m.accessSecretVersionFunc(ctx, req, opts...)
}

func gsmClient() *gsm.Client {
	return &gsm.Client{
		SM: mockSecretManager{
			accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
				return &secretmanagerpb.AccessSecretVersionResponse{
					Name:    "testName",
					Payload: &secretmanagerpb.SecretPayload{Data: []byte("secret")},
				}, nil
			},
		},
	}
}

func TestClient_UpdateControls(t *testing.T) {
	tests := []struct {
		name           string
		documentID     string
		request        *Request
		want           *Resource
		wantErr        string
		requestHandler http.HandlerFunc
		mockAPIc       apic.Clienter
	}{
		{
			name:       "successfully updated control doc",
			documentID: "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			request:    updateRequest,
			want: &Resource{
				TransactionControls: []*TransactionControl{
					{
						ShouldDeclineAll:     false,
						DeclineThreshold:     util.ToFloat64Ptr(200),
						ControlEnabled:       true,
						ControlType:          ccpb.ControlType_TCT_E_COMMERCE.String(),
						UserIdentifier:       "abhi-539d-4f93-ba00-77ef9ff873a2",
						AlertThreshold:       util.ToFloat64Ptr(10),
						ShouldAlertOnDecline: false,
					},
					{
						ShouldDeclineAll:     false,
						DeclineThreshold:     util.ToFloat64Ptr(200),
						ControlEnabled:       true,
						ControlType:          ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
						UserIdentifier:       "abhi-539d-4f93-ba00-77ef9ff873a2",
						AlertThreshold:       util.ToFloat64Ptr(10),
						ShouldAlertOnDecline: false,
					},
				},
				LastUpdateTimeStamp: "2017-03-21 05:03:53",
				DocumentID:          "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			},
			mockAPIc: testUtil.MockAPIcer{Response: updateControlDocumentResponse},
		},
		{
			name:       "failed to update control doc due to server error",
			documentID: "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			request:    request,
			wantErr:    "unexpected response from downstream",
			mockAPIc:   testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name:       "cannot parse nil control document request",
			documentID: "",
			request:    request,
			wantErr:    "invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := api.UpdateControls(testutil.GetContext(true), tt.documentID, tt.request)
			if tt.wantErr != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClient_UpdateControls_EdgeCases(t *testing.T) {
	gsmClient := gsmClient()
	key := "ClientIDKey"

	t.Run("fails when no auth header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()

		wantErr := errors.New("fabric error: status_code=Unauthenticated, error_code=5, message=service unauthenticated, reason=failed to extract auth token")

		api, err := NewClient(context.Background(), server.URL, key, server.Client(), 0, gsmClient)
		require.NoError(t, err)

		_, err = api.UpdateControls(context.Background(), "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b", request)

		assert.NotNil(t, err)
		assert.Equal(t, wantErr.Error(), err.Error())
	})
	t.Run("unable to find downstream", func(t *testing.T) {
		api, err := NewClient(context.Background(), "", key, &http.Client{Transport: &http.Transport{}}, 0, gsmClient)
		require.NoError(t, err)

		_, err = api.UpdateControls(testutil.GetContext(true), "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b", request)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "service unavailable")
	})
}

var updateControlDocumentResponse = []byte(`
{
"receivedTimestamp": "2017-03-21 05:03:53.501",
"resource": {
"transactionControls": [
{
"shouldDeclineAll": false,
"declineThreshold": 200,
"isControlEnabled": true,
"controlType": "TCT_E_COMMERCE",
"userIdentifier": "abhi-539d-4f93-ba00-77ef9ff873a2",
"alertThreshold": 10,
"shouldAlertOnDecline": false
},
{
"shouldDeclineAll": false,
"declineThreshold": 200,
"isControlEnabled": true,
"controlType": "TCT_ATM_WITHDRAW",
"userIdentifier": "abhi-539d-4f93-ba00-77ef9ff873a2",
"alertThreshold": 10,
"shouldAlertOnDecline": false
}
],
"lastUpdateTimeStamp": "2017-03-21 05:03:53",
"documentID": "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"
},
"processingTimeinMs": 37
}`)

var request = &Request{
	GlobalControls: []*GlobalControl{
		{
			ShouldDeclineAll:     false,
			ControlEnabled:       true,
			UserIdentifier:       "abhi-539d-4f93-ba00-77ef9ff873a2",
			AlertThreshold:       util.ToFloat64Ptr(15),
			ShouldAlertOnDecline: false,
		},
	},
}

func TestClient_CreateControls(t *testing.T) {
	tests := []struct {
		name       string
		documentID string
		request    *Request
		want       *Resource
		wantErr    string
		mockAPIc   apic.Clienter
	}{
		{
			name:       "successfully created a new control doc",
			documentID: "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			request:    request,
			want: &Resource{
				GlobalControls: []*GlobalControl{
					{
						ShouldDeclineAll:     false,
						ControlEnabled:       true,
						UserIdentifier:       "abhi-539d-4f93-ba00-77ef9ff873a2",
						AlertThreshold:       util.ToFloat64Ptr(15),
						ShouldAlertOnDecline: false,
					},
				},
				LastUpdateTimeStamp: "2018-06-06 21:28:42",
				DocumentID:          "ctc-vd-00d53da6-4590-4802-a034-bbc81dafb072",
			},
			mockAPIc: testUtil.MockAPIcer{Response: createDocumentResponse},
		},
		{
			name:       "failed to create a new control doc due to server error",
			documentID: "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			request:    request,
			wantErr:    "unexpected response from downstream",
			mockAPIc:   testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name:    "unable to parse nil control document",
			wantErr: "invalid argument",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testutil.GetContext(true)

			api := &client{
				apicClient: test.mockAPIc,
			}

			got, err := api.CreateControls(ctx, test.documentID, test.request)
			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestClient_CreateControls_EdgeCases(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("fails when no auth header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()
		documentId := testDocumentID

		wantErr := fmt.Errorf("fabric error: status_code=Unauthenticated, error_code=5, message=service unauthenticated, reason=failed to extract auth token")

		api, err := NewClient(testutil.GetContext(false), server.URL, key, server.Client(), 0, gsmClient)
		require.NoError(t, err)
		_, err = api.CreateControls(context.Background(), documentId, request)

		assert.NotNil(t, err)
		assert.Equal(t, wantErr.Error(), err.Error())
	})
	t.Run("unable to find downstream", func(t *testing.T) {
		documentId := testDocumentID

		api, err := NewClient(context.Background(), "", key, &http.Client{Transport: &http.Transport{}}, 0, gsmClient)
		require.NoError(t, err)
		_, err = api.CreateControls(testutil.GetContext(true), documentId, request)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "service unavailable")
	})
	t.Run("invalid argument", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()

		api, err := NewClient(context.Background(), server.URL, key, server.Client(), 0, gsmClient)
		require.NoError(t, err)
		_, err = api.CreateControls(context.Background(), "", nil)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid argument")
	})
}

var createDocumentResponse = []byte(`
{
"receivedTimestamp": "2018-06-06 21:28:42.817",
"resource": {
"globalControls": [
{
"shouldDeclineAll": false,
"isControlEnabled": true,
"userIdentifier": "abhi-539d-4f93-ba00-77ef9ff873a2",
"alertThreshold": 15,
"shouldAlertOnDecline": false
}
],
"lastUpdateTimeStamp": "2018-06-06 21:28:42",
"documentID": "ctc-vd-00d53da6-4590-4802-a034-bbc81dafb072"
},
"processingTimeinMs": 359
}`)

func TestClient_DeleteControls(t *testing.T) {
	tests := []struct {
		name       string
		documentID string
		request    *Request
		want       *Resource
		wantErr    string
		mockAPIc   apic.Clienter
	}{
		{
			name:       "successfully delete a control",
			documentID: "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			request:    request,
			want: &Resource{
				LastUpdateTimeStamp: "2017-03-21 05:03:53",
				DocumentID:          "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			},
			mockAPIc: testUtil.MockAPIcer{Response: deleteControlDocumentResponse},
		},
		{
			name:       "failed to delete a control due to server error",
			documentID: "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			request:    request,
			wantErr:    "unexpected response from downstream",
			mockAPIc:   testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name:    "unable to parse nil control document",
			wantErr: "invalid argument",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			api := &client{
				apicClient: test.mockAPIc,
			}

			got, err := api.DeleteControls(testutil.GetContext(true), test.documentID, test.request)
			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestClient_DeleteControls_EdgeCases(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("fails when no auth header", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()
		documentId := testDocumentID

		api, err := NewClient(context.Background(), server.URL, key, server.Client(), 0, gsmClient)
		require.NoError(t, err)

		_, err = api.DeleteControls(context.Background(), documentId, request)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "auth token")
	})
	t.Run("unable to find downstream", func(t *testing.T) {
		documentId := testDocumentID

		api, err := NewClient(context.Background(), "", key, &http.Client{Transport: &http.Transport{}}, 0, gsmClient)
		require.NoError(t, err)
		_, err = api.DeleteControls(testutil.GetContext(true), documentId, request)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "service unavailable")
	})
	t.Run("invalid argument", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()

		api, err := NewClient(context.Background(), server.URL, key, server.Client(), 0, gsmClient)
		require.NoError(t, err)
		_, err = api.DeleteControls(context.Background(), "", nil)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid argument")
	})
}

var deleteControlDocumentResponse = []byte(`
{
"receivedTimestamp": "2017-03-21 05:03:53.501",
"resource": {
"lastUpdateTimeStamp": "2017-03-21 05:03:53",
"documentID": "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"
},
"processingTimeinMs": 37
}`)

func Test_checkPrimaryAccountNumberErrors(t *testing.T) {
	tests := []string{
		key, "aaaaaaaaaaaaaaaa", "a", "", "%%",
	}
	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			assert.False(t, checkPrimaryAccountNumber(test))
		})
	}
}

func TestClientFromConfig(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("New Client with httpClient supplied", func(t *testing.T) {
		server := httptest.NewServer(nil)
		config := &Config{
			BaseURL:        server.URL,
			ClientIDEnvKey: key,
		}
		got, err := ClientFromConfig(context.Background(), server.Client(), config, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("New Client without httpClient supplied", func(t *testing.T) {
		config := &Config{
			ClientIDEnvKey: key,
		}
		got, err := ClientFromConfig(context.Background(), nil, config, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("unable to parse destination url", func(t *testing.T) {
		config := &Config{
			BaseURL: "%%",
		}
		_, err := ClientFromConfig(context.Background(), nil, config, gsmClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse configured url")
	})
	t.Run("config not provided, continues", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, nil, gsmClient)
		require.NoError(t, err)
		require.Nil(t, got)
	})
	t.Run("APIc client ID not provided", func(t *testing.T) {
		gsmClientNoSecret := &gsm.Client{
			SM: mockSecretManager{
				accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
					return &secretmanagerpb.AccessSecretVersionResponse{
						Name:    "testName",
						Payload: &secretmanagerpb.SecretPayload{Data: []byte("")},
					}, nil
				},
			},
		}

		config := &Config{
			ClientIDEnvKey: key,
		}
		_, err := ClientFromConfig(context.Background(), nil, config, gsmClientNoSecret)
		require.Error(t, err)
		require.Equal(t, err.Error(), "fabric error: status_code=Internal, error_code=1, message=failed to create APIc adapter, reason=unable to find clientID")
	})
}

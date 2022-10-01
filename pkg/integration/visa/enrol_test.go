package visa

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

func TestClient_EnrollByPan(t *testing.T) {
	tests := []struct {
		name                 string
		primaryAccountNumber string
		want                 string
		mockAPIc             apic.Clienter
		wantErr              string
	}{
		{
			name:                 "successfully enrolled",
			primaryAccountNumber: "4514170000000001",
			want:                 "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			mockAPIc:             testUtil.MockAPIcer{Response: registrationResponse},
		},
		{
			name:                 "successfully enrolled",
			primaryAccountNumber: "%%",
			wantErr:              "fabric error: status_code=InvalidArgument, error_code=4, message=failed to register card number, reason=cannot parse requested card number",
		},
		{
			name:                 "failed to enroll due to server error",
			primaryAccountNumber: "4514170000000002",
			wantErr:              "unexpected response from downstream",
			mockAPIc:             testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name:                 "unexpected response from visa",
			primaryAccountNumber: "4514170000000002",
			wantErr:              "unexpected response from downstream",
			mockAPIc:             testUtil.MockAPIcer{Response: []byte("%%")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := api.Register(testutil.GetContext(true), tt.primaryAccountNumber)
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

func TestClient_EnrollByPan_EdgeCases(t *testing.T) {
	gsmClient := gsmClient()

	t.Run("unable to build request", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()

		api, err := NewClient(context.Background(), server.URL, "ClientIDKey", server.Client(), 0, gsmClient)
		require.NoError(t, err)
		_, err = api.Register(context.Background(), "4514170000000002")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "fabric error: status_code=Unauthenticated, error_code=5, message=service unauthenticated, reason=failed to extract auth token")
	})
}

var registrationResponse = []byte(`
{
"receivedTimestamp": "2017-03-21 04:56:12.551",
"resource": {
"documentID": "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"
},
"processingTimeinMs": 10
}`)

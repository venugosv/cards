package visa

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	testUtil "github.com/anzx/fabric-cards/test/util"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
)

func TestClient_CardReplacement(t *testing.T) {
	type args struct {
		currentAccountId string
		newAccountId     string
	}
	tests := []struct {
		name           string
		args           args
		want           bool
		mockAPIc       apic.Clienter
		wantErr        string
		requestHandler http.HandlerFunc
	}{
		{
			name: "successfully replaced card number",
			args: args{
				currentAccountId: "4514170000000001",
				newAccountId:     "4514170000000002",
			},
			want:     true,
			mockAPIc: testUtil.MockAPIcer{Response: replacementResponse},
		},
		{
			name: "failed to parse response",
			args: args{
				currentAccountId: "4514170000000001",
				newAccountId:     "4514170000000002",
			},
			want:     false,
			wantErr:  "unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{Response: []byte(`%%`)},
		},
		{
			name: "failed to replaced card number",
			args: args{
				currentAccountId: "4514170000000001",
				newAccountId:     "4514170000000002",
			},
			want:     false,
			wantErr:  "failed to replace card number",
			mockAPIc: testUtil.MockAPIcer{Response: replacementFailedResponse},
		},
		{
			name: "failed to replace card number due to server error",
			args: args{
				currentAccountId: "4514170000000001",
				newAccountId:     "4514170000000002",
			},
			want:     false,
			wantErr:  "unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name: "cannot parse current card number",
			args: args{
				currentAccountId: "",
				newAccountId:     "4514170000000002",
			},
			want:    false,
			wantErr: "cannot parse requested card number",
		},
		{
			name: "cannot parse new card number",
			args: args{
				currentAccountId: "4514170000000002",
				newAccountId:     "",
			},
			want:    false,
			wantErr: "cannot parse requested card number",
		},
		{
			name: "cannot parse both card numbers",
			args: args{
				currentAccountId: "",
				newAccountId:     "",
			},
			want:    false,
			wantErr: "cannot parse requested card number",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testutil.GetContext(true)
			api := &client{
				apicClient: test.mockAPIc,
			}

			got, err := api.ReplaceCard(ctx, test.args.currentAccountId, test.args.newAccountId)
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

func TestClient_CardReplacement_EdgeCases(t *testing.T) {
	gsmClient := gsmClient()
	key := "ClientIDKey"

	t.Run("unable to build request", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()

		api, err := NewClient(context.Background(), server.URL, key, server.Client(), 0, gsmClient)
		require.NoError(t, err)
		_, err = api.ReplaceCard(context.Background(), "4514170000000001", "4514170000000002")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "fabric error: status_code=Unauthenticated, error_code=5, message=service unauthenticated, reason=failed to extract auth token")
	})
	t.Run("unable to find downstream", func(t *testing.T) {
		api, err := NewClient(context.Background(), "", key, &http.Client{Transport: &http.Transport{}}, 0, gsmClient)
		require.NoError(t, err)
		_, err = api.ReplaceCard(testutil.GetContext(true), "4514170000000001", "4514170000000002")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "service unavailable")
	})
}

var replacementResponse = []byte(`{
"receivedTimestamp": "2017-03-21 05:54:07.729",
"resource": {
"status": "SUCCESS"
},
"processingTimeinMs": 105
}`)

var replacementFailedResponse = []byte(`{
"receivedTimestamp": "2017-03-21 05:54:07.729",
"resource": {
"status": "FAILED"
},
"processingTimeinMs": 105
}`)

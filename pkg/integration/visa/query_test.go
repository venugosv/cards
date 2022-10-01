package visa

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"github.com/anzx/fabric-cards/pkg/util/testutil"
	testUtil "github.com/anzx/fabric-cards/test/util"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_QueryControls(t *testing.T) {
	tests := []struct {
		name                 string
		primaryAccountNumber string
		want                 *Resource
		wantErr              string
		mockAPIc             apic.Clienter
	}{
		{
			name:                 "successfully queried controls",
			primaryAccountNumber: "4514170000000001",
			want: &Resource{
				MerchantControls: []*MerchantControl{
					{
						ShouldDeclineAll:      true,
						ControlEnabled:        true,
						ControlType:           ccpb.ControlType_MCT_GAMBLING.String(),
						ImpulseDelayPeriod:    util.ToStringPtr("24:00"),
						ImpulseDelayStart:     util.ToStringPtr("2020-10-26 02:52:13"),
						ImpulseDelayEnd:       util.ToStringPtr("2020-10-27 02:52:13"),
						ImpulseDelayRemaining: util.ToStringPtr("00:00:00"),
						UserIdentifier:        "abhishek492-aghjba3-aa4bb",
						AlertThreshold:        util.ToFloat64Ptr(15),
						ShouldAlertOnDecline:  false,
					},
				},
				LastUpdateTimeStamp: "2020-10-26 02:52:13",
				DocumentID:          "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b",
			},
			mockAPIc: testUtil.MockAPIcer{Response: queryListResponse},
		},
		{
			name:                 "failed to parse response",
			primaryAccountNumber: "4514170000000001",
			wantErr:              "unexpected response from downstream",
			mockAPIc:             testUtil.MockAPIcer{Response: []byte("%%")},
		},
		{
			name:                 "failed to query controls due to server error",
			primaryAccountNumber: "4514170000000002",
			wantErr:              "unexpected response from downstream",
			mockAPIc:             testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name:                 "cannot parse requested card number",
			primaryAccountNumber: "",
			wantErr:              "cannot parse requested card number",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testutil.GetContext(true)
			api := &client{
				apicClient: test.mockAPIc,
			}

			got, err := api.QueryControls(ctx, test.primaryAccountNumber)
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

func TestClient_QueryControls_EdgeCases(t *testing.T) {
	gsmClient := gsmClient()
	t.Run("unable to build request", func(t *testing.T) {
		server := httptest.NewServer(nil)
		defer server.Close()

		api, err := NewClient(context.Background(), server.URL, key, server.Client(), 0, gsmClient)
		require.NoError(t, err)
		_, err = api.QueryControls(context.Background(), "4514170000000002")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "fabric error: status_code=Unauthenticated, error_code=5, message=service unauthenticated, reason=failed to extract auth token")
	})
	t.Run("unable to find downstream", func(t *testing.T) {
		api, err := NewClient(context.Background(), "", key, &http.Client{Transport: &http.Transport{}}, 0, gsmClient)
		require.NoError(t, err)
		_, err = api.QueryControls(testutil.GetContext(true), "4514170000000002")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "service unavailable")
	})
}

var queryListResponse = []byte(`{"receivedTimestamp":"2020-10-27 03:34:19.209","processingTimeinMs":46,"resource":{"controlDocuments":[{"merchantControls":[{"shouldDeclineAll":true,"alertThreshold":15,"userIdentifier":"abhishek492-aghjba3-aa4bb","isControlEnabled":true,"controlType":"MCT_GAMBLING","impulseDelayPeriod":"24:00","impulseDelayStart":"2020-10-26 02:52:13","impulseDelayEnd":"2020-10-27 02:52:13","impulseDelayRemaining":"00:00:00"}],"lastUpdateTimeStamp":"2020-10-26 02:52:13","documentID":"ctc-vd-857a8766-160b-498d-820f-bf4339949c1b","createdOnDate":"2020-10-06 06:10:52.993"}]}}`)

func Test_temp(t *testing.T) {
	queryListResponseWithMultipleDocs := []byte(`{
  "receivedTimestamp": "2020-10-27 03:34:19.209",
  "processingTimeinMs": 46,
  "resource": {
    "controlDocuments": [
      {
        "merchantControls": [
          {
           "shouldDeclineAll": true,
            "alertThreshold": 15,
            "userIdentifier": "abhishek492-aghjba3-aa4bb",
            "isControlEnabled": true,
            "controlType": "MCT_GAMBLING",
            "impulseDelayPeriod": "24:00",
            "impulseDelayStart": "2020-10-26 02:52:13",
            "impulseDelayEnd": "2020-10-27 02:52:13",
            "impulseDelayRemaining": "00:00:00"
          }
        ],
        "lastUpdateTimeStamp": "2020-10-26 02:52:13",
        "documentID": "DocID1",
        "createdOnDate": "2020-10-06 06:10:52.993"
      },
      {
        "merchantControls": [
          {
            "shouldDeclineAll": true,
            "alertThreshold": 15,
            "userIdentifier": "abhishek492-aghjba3-aa4bb",
            "isControlEnabled": true,
            "controlType": "MCT_GAMBLING",
            "impulseDelayPeriod": "24:00",
            "impulseDelayStart": "2020-10-26 02:52:13",
            "impulseDelayEnd": "2020-10-27 02:52:13",
            "impulseDelayRemaining": "00:00:00"
          }
        ],
        "lastUpdateTimeStamp": "2020-10-26 02:52:13",
        "documentID": "DocID2",
        "createdOnDate": "2020-10-06 06:10:52.993"
      }
    ]
  }
}`)

	mockAPIc := testUtil.MockAPIcer{Response: queryListResponseWithMultipleDocs}
	want := &Resource{
		LastUpdateTimeStamp: "2020-10-26 02:52:13",
		DocumentID:          "DocID1",
		GlobalControls:      nil,
		MerchantControls: []*MerchantControl{
			{
				ShouldDeclineAll:      true,
				AlertThreshold:        util.ToFloat64Ptr(15),
				UserIdentifier:        "abhishek492-aghjba3-aa4bb",
				ControlEnabled:        true,
				ControlType:           "MCT_GAMBLING",
				ImpulseDelayPeriod:    util.ToStringPtr("24:00"),
				ImpulseDelayStart:     util.ToStringPtr("2020-10-26 02:52:13"),
				ImpulseDelayEnd:       util.ToStringPtr("2020-10-27 02:52:13"),
				ImpulseDelayRemaining: util.ToStringPtr("00:00:00"),
			},
		},
		TransactionControls: nil,
	}
	t.Run("temp", func(t *testing.T) {
		api := &client{
			apicClient: mockAPIc,
		}
		got, err := api.QueryControls(testutil.GetContext(true), "1234567890123456")
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
}

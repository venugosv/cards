package visa

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/stretchr/testify/assert"
)

func Test_handleAccountUpdateResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *StatusResource
		wantErr bool
	}{
		{
			name:  "successfully unmarshal bytes",
			input: replacementResponse,
			want: &StatusResource{
				Status: "SUCCESS",
			},
		},
		{
			name:    "failed to unmarshal bytes",
			input:   []byte(`%%`),
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := handleAccountUpdateResponse(context.Background(), test.input)
			if test.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_handleTransactionControlDocument(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *Resource
		wantErr bool
	}{
		{
			name:  "successfully unmarshal bytes",
			input: createDocumentResponse,
			want: &Resource{
				ShouldDecouple:      false,
				LastUpdateTimeStamp: "2018-06-06 21:28:42",
				DocumentID:          "ctc-vd-00d53da6-4590-4802-a034-bbc81dafb072",
				GlobalControls: []*GlobalControl{
					{
						ShouldDeclineAll:     false,
						ControlEnabled:       true,
						UserIdentifier:       "abhi-539d-4f93-ba00-77ef9ff873a2",
						AlertThreshold:       util.ToFloat64Ptr(15),
						ShouldAlertOnDecline: false,
					},
				},
			},
		},
		{
			name:    "failed to unmarshal bytes",
			input:   []byte(`%%`),
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := handleTransactionControlDocument(context.Background(), test.input)
			if test.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_handleTransactionControlListResponses(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []Resource
		wantErr bool
	}{
		{
			name:  "successfully unmarshal bytes",
			input: queryListResponse,
			want: []Resource{
				{
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
			},
		},
		{
			name:    "successfully unmarshal bytes, no control documents",
			input:   []byte(`{"receivedTimestamp":"2020-10-27 03:34:19.209","processingTimeinMs":46,"resource":{"controlDocuments":[]}}`),
			wantErr: true,
		},
		{
			name:    "failed to unmarshal bytes",
			input:   []byte(`%%`),
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := handleTransactionControlListResponses(context.Background(), test.input)
			if test.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

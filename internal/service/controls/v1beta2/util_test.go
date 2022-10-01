package v1beta2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
)

const (
	documentID = "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"
)

var (
	myPersonaId    = "foobar"
	otherPersonaId = "barfoo"
)

var sampleControlDoc = []byte(" {\n    \"createdOnDate\": \"2021-12-09 22:58:40.302\",\n    \"documentID\": \"ctc-vn-b5820ae0-bf1c-4dd4-a441-e3b695cf767a\",\n    \"lastUpdateTimeStamp\": \"2021-12-10 02:53:44\",\n    \"merchantControls\": [\n      {\n        \"controlType\": \"MCT_GAMBLING\",\n        \"declineThreshold\": 0,\n        \"impulseDelayEnd\": \"2021-12-10 02:54:44\",\n        \"impulseDelayPeriod\": \"00:01\",\n        \"impulseDelayRemaining\": \"00:00:59\",\n        \"impulseDelayStart\": \"2021-12-10 02:53:44\",\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true,\n        \"userIdentifier\": \"b345d387-8943-4c3a-9019-22dd80fe8716\"\n      },\n      {\n        \"controlType\": \"MCT_GAMBLING\",\n        \"declineThreshold\": 0,\n        \"impulseDelayEnd\": \"2021-12-10 02:54:44\",\n        \"impulseDelayPeriod\": \"00:01\",\n        \"impulseDelayRemaining\": \"00:00:59\",\n        \"impulseDelayStart\": \"2021-12-10 02:53:44\",\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true,\n        \"userIdentifier\": \"fb4aa2cb-d744-4b46-9dae-d717d060cb8c\"\n      },\n      {\n        \"controlType\": \"MCT_GAMBLING\",\n        \"declineThreshold\": 0,\n        \"impulseDelayEnd\": \"2021-12-10 02:54:44\",\n        \"impulseDelayPeriod\": \"00:01\",\n        \"impulseDelayRemaining\": \"00:00:59\",\n        \"impulseDelayStart\": \"2021-12-10 02:53:44\",\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true\n      }\n    ],\n    \"transactionControls\": [\n      {\n        \"controlType\": \"TCT_ATM_WITHDRAW\",\n        \"declineThreshold\": 0,\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true,\n        \"userIdentifier\": \"b345d387-8943-4c3a-9019-22dd80fe8716\"\n      },\n      {\n        \"controlType\": \"TCT_ATM_WITHDRAW\",\n        \"declineThreshold\": 0,\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true,\n        \"userIdentifier\": \"fb4aa2cb-d744-4b46-9dae-d717d060cb8c\"\n      },\n      {\n        \"controlType\": \"TCT_ATM_WITHDRAW\",\n        \"declineThreshold\": 0,\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true\n      },\n      {\n        \"alertThreshold\": 15,\n        \"controlType\": \"TCT_E_COMMERCE\",\n        \"declineThreshold\": 0,\n        \"isControlEnabled\": true,\n        \"shouldAlertOnDecline\": false,\n        \"shouldDeclineAll\": true,\n        \"userIdentifier\": \"fb4aa2cb-d744-4b46-9dae-d717d060cb8c\"\n      },\n      {\n        \"controlType\": \"TCT_E_COMMERCE\",\n        \"declineThreshold\": 0,\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true,\n        \"userIdentifier\": \"b345d387-8943-4c3a-9019-22dd80fe8716\"\n      },\n      {\n        \"controlType\": \"TCT_E_COMMERCE\",\n        \"declineThreshold\": 0,\n        \"isControlEnabled\": true,\n        \"shouldDeclineAll\": true\n      }]}\n")

func Test_getCardControlResponse(t *testing.T) {
	tests := []struct {
		name      string
		personaID string
		input     *crpb.Resource
		want      *ccpb.CardControlResponse
		count     int
	}{
		{
			name:      "Successful call with full document",
			personaID: myPersonaId,
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
					},
				},
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:      "MCT_ADULT_ENTERTAINMENT",
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						ControlType:      "TCT_ATM_WITHDRAW",
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
					{
						ControlType: ccpb.ControlType_TCT_ATM_WITHDRAW,
					},
					{
						ControlType: ccpb.ControlType_MCT_ADULT_ENTERTAINMENT,
					},
				},
			},
			count: 3,
		},
		{
			name: "Successful call with global document",
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
				},
			},
			count: 1,
		},
		{
			name: "Successful call with merchant document",
			input: &crpb.Resource{
				MerchantControls: []*crpb.MerchantControl{
					{
						IsControlEnabled: true,
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
						UserIdentifier:   &myPersonaId,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_MCT_GAMBLING,
					},
				},
			},
			count: 1,
		},
		{
			name: "Successful call with transaction document",
			input: &crpb.Resource{
				TransactionControls: []*crpb.TransactionControl{
					{
						IsControlEnabled: true,
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
						UserIdentifier:   &myPersonaId,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_TCT_ATM_WITHDRAW,
					},
				},
			},
			count: 1,
		},
		{
			name: "Successful call with no document",
			input: &crpb.Resource{
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
			},
			count: 0,
		},
		{
			name: "Successful call with empty document",
			input: &crpb.Resource{
				GlobalControls:      nil,
				MerchantControls:    nil,
				TransactionControls: nil,
				LastUpdateTimeStamp: "",
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
			},
			count: 0,
		},
		{
			name: "Successful call with all controls not enable",
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: false,
						UserIdentifier:   &myPersonaId,
					},
				},
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:      "MCT_ADULT_ENTERTAINMENT",
						IsControlEnabled: false,
						UserIdentifier:   &myPersonaId,
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						ControlType:      "TCT_ATM_WITHDRAW",
						IsControlEnabled: false,
						UserIdentifier:   &myPersonaId,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
			},
			count: 0,
		},
		{
			name: "Successful call with global control not enable, no merchant document",
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: false,
						UserIdentifier:   &myPersonaId,
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_TCT_ATM_WITHDRAW,
					},
				},
			},
			count: 1,
		},
		{
			name: "Successful call with merchant control not enable, no transaction document",
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
					},
				},
				MerchantControls: []*crpb.MerchantControl{
					{
						IsControlEnabled: false,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
					},
					{
						IsControlEnabled: false,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
				},
			},
			count: 1,
		},
		{
			name: "Successful call with transaction control not enable, no global document",
			input: &crpb.Resource{
				MerchantControls: []*crpb.MerchantControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						IsControlEnabled: false,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_MCT_GAMBLING,
					},
				},
			},
			count: 1,
		},
		{
			name:  "Successful call with nil document",
			input: nil,
			want:  &ccpb.CardControlResponse{},
			count: 0,
		},
		{
			name: "global control correctly filtered on PersonaID",
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
				},
			},
			count: 1,
		},
		{
			name: "transaction controls are correctly filtered on PersonaID",
			input: &crpb.Resource{
				TransactionControls: []*crpb.TransactionControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_TCT_CONTACTLESS.String(),
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_TCT_ATM_WITHDRAW,
					}, {
						ControlType: ccpb.ControlType_TCT_CONTACTLESS,
					},
				},
			},
			count: 2,
		},
		{
			name: "merchant controls are correctly filtered on PersonaID",
			input: &crpb.Resource{
				MerchantControls: []*crpb.MerchantControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_MCT_ALCOHOL.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_MCT_ALCOHOL,
					},
				},
			},
			count: 2,
		},
		{
			name: "many controls are correctly filtered on PersonaID",
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
					},
				},
				MerchantControls: []*crpb.MerchantControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_MCT_ALCOHOL.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_MCT_ALCOHOL.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_MCT_GROCERY.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_MCT_GROCERY.String(),
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						IsControlEnabled: true,
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					},
					{
						IsControlEnabled: true,
						ControlType:      ccpb.ControlType_TCT_CONTACTLESS.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_TCT_CONTACTLESS.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_TCT_CONTACTLESS.String(),
					},
					{
						IsControlEnabled: true,
						ControlType:      ccpb.ControlType_TCT_CROSS_BORDER.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &myPersonaId,
						ControlType:      ccpb.ControlType_TCT_CROSS_BORDER.String(),
					},
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_TCT_CROSS_BORDER.String(),
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
					{
						ControlType: ccpb.ControlType_MCT_ALCOHOL,
					},
					{
						ControlType: ccpb.ControlType_MCT_GAMBLING,
					},
					{
						ControlType: ccpb.ControlType_MCT_GROCERY,
					},
					{
						ControlType: ccpb.ControlType_TCT_ATM_WITHDRAW,
					},
					{
						ControlType: ccpb.ControlType_TCT_CONTACTLESS,
					},
					{
						ControlType: ccpb.ControlType_TCT_CROSS_BORDER,
					},
				},
			},
			count: 7,
		},
		{
			name: "no controls for this user",
			input: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
					},
				},
				MerchantControls: []*crpb.MerchantControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_MCT_AIRFARE.String(),
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						IsControlEnabled: true,
						UserIdentifier:   &otherPersonaId,
						ControlType:      ccpb.ControlType_TCT_CONTACTLESS.String(),
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: token1,
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
					{
						ControlType: ccpb.ControlType_TCT_CONTACTLESS,
					},
					{
						ControlType: ccpb.ControlType_MCT_AIRFARE,
					},
				},
			},
			count: 3,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got := getCardControlResponse(test.input, token1)
			assert.Equal(t, test.want.GetTokenizedCardNumber(), got.GetTokenizedCardNumber())
			assert.Len(t, got.GetCardControls(), test.count)
		})
	}
}

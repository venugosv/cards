package v1beta1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

const (
	documentID = "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"
)

func Test_getCardControlResponse(t *testing.T) {
	tests := []struct {
		name  string
		input *visa.Resource
		want  *ccpb.CardControlResponse
	}{
		{
			name: "Successful call with full document",
			input: &visa.Resource{
				GlobalControls: []*visa.GlobalControl{
					{
						ControlEnabled: true,
					},
				},
				MerchantControls: []*visa.MerchantControl{
					{
						ControlType:    "MCT_ADULT_ENTERTAINMENT",
						ControlEnabled: true,
					},
				},
				TransactionControls: []*visa.TransactionControl{
					{
						ControlType:    "TCT_ATM_WITHDRAW",
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    ccpb.ControlType_GCT_GLOBAL,
						ControlEnabled: true,
					},
					{
						ControlType:    ccpb.ControlType_TCT_ATM_WITHDRAW,
						ControlEnabled: true,
					},
					{
						ControlType:    ccpb.ControlType_MCT_ADULT_ENTERTAINMENT,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name: "Successful call with global document",
			input: &visa.Resource{
				GlobalControls: []*visa.GlobalControl{
					{
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    ccpb.ControlType_GCT_GLOBAL,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name: "Successful call with merchant document",
			input: &visa.Resource{
				GlobalControls: []*visa.GlobalControl{
					{
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    ccpb.ControlType_GCT_GLOBAL,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name: "Successful call with transaction document",
			input: &visa.Resource{
				GlobalControls: []*visa.GlobalControl{
					{
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    ccpb.ControlType_GCT_GLOBAL,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name: "Successful call with no document",
			input: &visa.Resource{
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Now().String(),
			},
			want: &ccpb.CardControlResponse{},
		},
		{
			name: "Successful call with empty document",
			input: &visa.Resource{
				GlobalControls:      nil,
				MerchantControls:    nil,
				TransactionControls: nil,
				LastUpdateTimeStamp: "",
			},
			want: &ccpb.CardControlResponse{},
		},
		{
			name:  "Successful call with nil document",
			input: nil,
			want:  &ccpb.CardControlResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCardControlResponse(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

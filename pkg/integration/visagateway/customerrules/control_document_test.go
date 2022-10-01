package customerrules

import (
	"context"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/pkg/feature"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/anzx/fabric-cards/pkg/integration/util"

	"github.com/stretchr/testify/require"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
	"github.com/stretchr/testify/assert"
)

const documentID = "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"

func Test_GetDeleteRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		document    *crpb.Resource
		controlType []ccpb.ControlType
		want        *crpb.ControlRequest
		wantBool    bool
	}{
		{
			name: "successfully add global control to remove list",
			document: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
			want: &crpb.ControlRequest{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
					},
				},
			},
			wantBool: true,
		},
		{
			name: "successfully add transaction control to remove list",
			document: &crpb.Resource{
				TransactionControls: []*crpb.TransactionControl{
					{
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
						IsControlEnabled: true,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			want: &crpb.ControlRequest{
				TransactionControls: []*crpb.TransactionControl{
					{
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
						IsControlEnabled: true,
					},
				},
			},
			wantBool: true,
		},
		{
			name: "successfully add merchant control to remove list",
			document: &crpb.Resource{
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:      ccpb.ControlType_MCT_ALCOHOL.String(),
						IsControlEnabled: true,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: []ccpb.ControlType{ccpb.ControlType_MCT_ALCOHOL},
			want: &crpb.ControlRequest{
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:      ccpb.ControlType_MCT_ALCOHOL.String(),
						IsControlEnabled: true,
					},
				},
			},
			wantBool: true,
		},
		{
			name: "successfully add merchant gambling control to remove list when impulse has expired",
			document: &crpb.Resource{
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:      ccpb.ControlType_MCT_ALCOHOL.String(),
						IsControlEnabled: true,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: []ccpb.ControlType{ccpb.ControlType_MCT_ALCOHOL},
			want: &crpb.ControlRequest{
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:      ccpb.ControlType_MCT_ALCOHOL.String(),
						IsControlEnabled: true,
					},
				},
			},
			wantBool: true,
		},
		{
			name: "successfully add many controls with gambling block to remove request",
			document: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
						IsControlEnabled: true,
					},
					{
						ControlType:      ccpb.ControlType_TCT_CONTACTLESS.String(),
						IsControlEnabled: true,
					},
				},
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:           ccpb.ControlType_MCT_GAMBLING.String(),
						IsControlEnabled:      true,
						ImpulseDelayPeriod:    util.ToStringPtr("48:00"),
						ImpulseDelayRemaining: util.ToStringPtr(noTimeRemaining),
						ImpulseDelayStart:     util.ToStringPtr("2020-05-18 23:34:50"),
						ImpulseDelayEnd:       util.ToStringPtr("2020-05-20 23:34:50"),
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: []ccpb.ControlType{
				ccpb.ControlType_GCT_GLOBAL,
				ccpb.ControlType_TCT_ATM_WITHDRAW,
				ccpb.ControlType_MCT_GAMBLING,
			},
			want: &crpb.ControlRequest{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
					},
				},
				TransactionControls: []*crpb.TransactionControl{
					{
						ControlType:      ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
						IsControlEnabled: true,
					},
				},
				MerchantControls: []*crpb.MerchantControl{
					{
						ControlType:           ccpb.ControlType_MCT_GAMBLING.String(),
						IsControlEnabled:      true,
						ImpulseDelayPeriod:    util.ToStringPtr("48:00"),
						ImpulseDelayRemaining: util.ToStringPtr(noTimeRemaining),
						ImpulseDelayStart:     util.ToStringPtr("2020-05-18 23:34:50"),
						ImpulseDelayEnd:       util.ToStringPtr("2020-05-20 23:34:50"),
					},
				},
			},
			wantBool: true,
		},
		{
			name: "false return on control not in the document",
			document: &crpb.Resource{
				GlobalControls: []*crpb.GlobalControl{
					{
						IsControlEnabled: true,
					},
				},
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: []ccpb.ControlType{ccpb.ControlType_TCT_CONTACTLESS},
			want:        nil,
			wantBool:    false,
		},
		{
			name: "false return on unknown control",
			document: &crpb.Resource{
				DocumentId:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: []ccpb.ControlType{ccpb.ControlType_UNKNOWN_UNSPECIFIED},
			want:        nil,
			wantBool:    false,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			request, ok := GetDeleteRequest(test.document, test.controlType)
			assert.Equal(t, test.want, request)
			assert.Equal(t, test.wantBool, ok)
		})
	}
}

func TestControlRequest(t *testing.T) {
	err := feature.FeatureGate.Set(map[feature.Feature]bool{
		feature.GCT_GLOBAL:       true,
		feature.MCT_GAMBLING:     true,
		feature.TCT_ATM_WITHDRAW: true,
	})
	require.NoError(t, err)
	t.Run("testWithControls", func(t *testing.T) {
		cr := []*ccpb.ControlRequest{
			{
				ControlType: ccpb.ControlType_GCT_GLOBAL,
			}, {
				ControlType: ccpb.ControlType_MCT_GAMBLING,
			}, {
				ControlType: ccpb.ControlType_TCT_ATM_WITHDRAW,
			},
		}
		controls, err := WithControls(context.Background(), cr, "UID")
		require.NoError(t, err)
		got := ControlRequest(controls...)
		want := &crpb.ControlRequest{
			GlobalControls: []*crpb.GlobalControl{
				{
					IsControlEnabled:                  true,
					UserIdentifier:                    util.ToStringPtr("UID"),
					DeclineAllNonTokenizeTransactions: util.ToBoolPtr(true),
					ShouldAlertOnDecline:              util.ToBoolPtr(true),
				},
			},
			MerchantControls: []*crpb.MerchantControl{
				{
					ShouldDeclineAll:     util.ToBoolPtr(true),
					IsControlEnabled:     true,
					UserIdentifier:       util.ToStringPtr("UID"),
					ControlType:          ccpb.ControlType_MCT_GAMBLING.String(),
					ShouldAlertOnDecline: util.ToBoolPtr(true),
				},
			},
			TransactionControls: []*crpb.TransactionControl{
				{
					ShouldDeclineAll:     util.ToBoolPtr(true),
					IsControlEnabled:     true,
					ControlType:          ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					UserIdentifier:       util.ToStringPtr("UID"),
					ShouldAlertOnDecline: util.ToBoolPtr(true),
				},
			},
		}
		assert.Equal(t, want, got)
	})
}

func TestWithControlsErrors(t *testing.T) {
	err := feature.FeatureGate.Set(map[feature.Feature]bool{
		feature.GCT_GLOBAL: false,
	})
	require.NoError(t, err)
	t.Run("test", func(t *testing.T) {
		cr := []*ccpb.ControlRequest{
			{
				ControlType: ccpb.ControlType_GCT_GLOBAL,
			},
		}
		got, err := WithControls(context.Background(), cr, "UID")
		require.Error(t, err)
		require.Nil(t, got)
		assert.Contains(t, err.Error(), "control is disabled")
	})
}

func TestCategory_String(t *testing.T) {
	tests := map[Category]string{
		TRANSACTION: "TRANSACTION",
		MERCHANT:    "MERCHANT",
		GLOBAL:      "GLOBAL",
		4:           "",
	}
	for category, str := range tests {
		assert.Equal(t, category.String(), str)
	}
}

func TestGetCategory(t *testing.T) {
	tests := map[ccpb.ControlType]Category{
		ccpb.ControlType_TCT_ATM_WITHDRAW:            TRANSACTION,
		ccpb.ControlType_TCT_AUTO_PAY:                TRANSACTION,
		ccpb.ControlType_TCT_BRICK_AND_MORTAR:        TRANSACTION,
		ccpb.ControlType_TCT_CROSS_BORDER:            TRANSACTION,
		ccpb.ControlType_TCT_E_COMMERCE:              TRANSACTION,
		ccpb.ControlType_TCT_CONTACTLESS:             TRANSACTION,
		ccpb.ControlType_MCT_ADULT_ENTERTAINMENT:     MERCHANT,
		ccpb.ControlType_MCT_AIRFARE:                 MERCHANT,
		ccpb.ControlType_MCT_ALCOHOL:                 MERCHANT,
		ccpb.ControlType_MCT_APPAREL_AND_ACCESSORIES: MERCHANT,
		ccpb.ControlType_MCT_AUTOMOTIVE:              MERCHANT,
		ccpb.ControlType_MCT_CAR_RENTAL:              MERCHANT,
		ccpb.ControlType_MCT_ELECTRONICS:             MERCHANT,
		ccpb.ControlType_MCT_SPORT_AND_RECREATION:    MERCHANT,
		ccpb.ControlType_MCT_GAMBLING:                MERCHANT,
		ccpb.ControlType_MCT_GAS_AND_PETROLEUM:       MERCHANT,
		ccpb.ControlType_MCT_GROCERY:                 MERCHANT,
		ccpb.ControlType_MCT_HOTEL_AND_LODGING:       MERCHANT,
		ccpb.ControlType_MCT_HOUSEHOLD:               MERCHANT,
		ccpb.ControlType_MCT_PERSONAL_CARE:           MERCHANT,
		ccpb.ControlType_MCT_SMOKE_AND_TOBACCO:       MERCHANT,
		ccpb.ControlType_GCT_GLOBAL:                  GLOBAL,
	}
	for controlType, category := range tests {
		assert.Equal(t, GetCategory(controlType), category)
	}
}

func TestEnrolled(t *testing.T) {
	t.Run("enrolled", func(t *testing.T) {
		assert.True(t, Enrolled(&crpb.Resource{DocumentId: documentID}))
	})
	t.Run("not enrolled", func(t *testing.T) {
		assert.False(t, Enrolled(&crpb.Resource{DocumentId: "NOT_ENROLLED"}))
	})
}

func TestGetImpulseDelayStartTimestamp(t *testing.T) {
	tests := []struct {
		name         string
		merchControl *crpb.MerchantControl
		want         *timestamppb.Timestamp
	}{
		{
			name: "happy path",
			merchControl: &crpb.MerchantControl{
				ImpulseDelayStart: util.ToStringPtr("2006/01/02 15:04:05"),
			},
			want: &timestamppb.Timestamp{
				Seconds: 1136214245,
				Nanos:   0,
			},
		},
		{
			name: "nil delay start",
			merchControl: &crpb.MerchantControl{
				ImpulseDelayStart: util.ToStringPtr(""),
			},
		},
		{
			name: "nil delay start",
			merchControl: &crpb.MerchantControl{
				ImpulseDelayStart: util.ToStringPtr("qwerty"),
			},
		},
	}
	for _, test := range tests {
		got := GetImpulseDelayStartTimestamp(test.merchControl)
		assert.Equal(t, test.want, got)
	}
}

func TestGetImpulseDelayPeriodProto(t *testing.T) {
	tests := []struct {
		name         string
		merchControl *crpb.MerchantControl
		want         *durationpb.Duration
	}{
		{
			name: "happy path",
			merchControl: &crpb.MerchantControl{
				ImpulseDelayPeriod: util.ToStringPtr("48:00"),
			},
			want: &durationpb.Duration{
				Seconds: 172800,
				Nanos:   0,
			},
		},
		{
			name: "nil delay period",
			merchControl: &crpb.MerchantControl{
				ImpulseDelayPeriod: util.ToStringPtr(""),
			},
		},
		{
			name: "invalid delay period",
			merchControl: &crpb.MerchantControl{
				ImpulseDelayPeriod: util.ToStringPtr("qwerty"),
			},
		},
	}
	for _, test := range tests {
		got := GetImpulseDelayPeriodProto(test.merchControl)
		assert.Equal(t, test.want, got)
	}
}

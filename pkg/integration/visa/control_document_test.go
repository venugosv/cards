package visa

import (
	"context"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/stretchr/testify/assert"
)

const documentID = "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"

func Test_removeControl(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		document    *Resource
		controlType ccpb.ControlType
		want        bool
	}{
		{
			name: "successfully remove global control",
			document: &Resource{
				GlobalControls: []*GlobalControl{
					{
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: ccpb.ControlType_GCT_GLOBAL,
			want:        true,
		},
		{
			name: "successfully remove transaction control",
			document: &Resource{
				TransactionControls: []*TransactionControl{
					{
						ControlType:    ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: ccpb.ControlType_TCT_ATM_WITHDRAW,
			want:        true,
		},
		{
			name: "successfully remove merchant control",
			document: &Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:    ccpb.ControlType_MCT_ADULT_ENTERTAINMENT.String(),
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: ccpb.ControlType_MCT_ADULT_ENTERTAINMENT,
			want:        true,
		},
		{
			name: "false return on unknown control",
			document: &Resource{
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: ccpb.ControlType_UNKNOWN_UNSPECIFIED,
			want:        false,
		},
		{
			name: "successfully remove merchant gambling control",
			document: &Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:    ccpb.ControlType_MCT_GAMBLING.String(),
						ControlEnabled: true,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			controlType: ccpb.ControlType_MCT_GAMBLING,
			want:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.document.RemoveControlByType(tt.controlType))
		})
	}
}

func Test_removeGamblingControl(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		document             *Resource
		index                int
		want                 bool
		wantMerchantControls []*MerchantControl
	}{
		{
			name: "return false when controlType does not equal MCT_GAMBLING",
			document: &Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:      ccpb.ControlType_MCT_ADULT_ENTERTAINMENT.String(),
						ControlEnabled:   true,
						ShouldDeclineAll: false,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			index: 0,
			want:  false,
			wantMerchantControls: []*MerchantControl{
				{
					ControlType:      ccpb.ControlType_MCT_ADULT_ENTERTAINMENT.String(),
					ControlEnabled:   true,
					ShouldDeclineAll: false,
				},
			},
		},
		{
			name: "add impulse delay on control where impulseDelay doesnt exist",
			document: &Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:      ccpb.ControlType_MCT_GAMBLING.String(),
						ControlEnabled:   true,
						ShouldDeclineAll: false,
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			index: 0,
			want:  true,
			wantMerchantControls: []*MerchantControl{
				{
					ControlType:        ccpb.ControlType_MCT_GAMBLING.String(),
					ControlEnabled:     true,
					ShouldDeclineAll:   false,
					ImpulseDelayPeriod: util.ToStringPtr("48:00"),
				},
			},
		},
		{
			name: "return false when attempting to remove control which has an existing impulse delay",
			document: &Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:           ccpb.ControlType_MCT_GAMBLING.String(),
						ControlEnabled:        true,
						ShouldDeclineAll:      false,
						ImpulseDelayPeriod:    util.ToStringPtr("48:00"),
						ImpulseDelayRemaining: util.ToStringPtr("48:00:00"),
						ImpulseDelayStart:     util.ToStringPtr("2020-05-18 23:34:50"),
						ImpulseDelayEnd:       util.ToStringPtr("2020-07-18 23:34:50"),
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			index: 0,
			want:  false,
			wantMerchantControls: []*MerchantControl{
				{
					ControlType:           ccpb.ControlType_MCT_GAMBLING.String(),
					ControlEnabled:        true,
					ShouldDeclineAll:      false,
					ImpulseDelayPeriod:    util.ToStringPtr("48:00"),
					ImpulseDelayRemaining: util.ToStringPtr("48:00:00"),
					ImpulseDelayStart:     util.ToStringPtr("2020-05-18 23:34:50"),
					ImpulseDelayEnd:       util.ToStringPtr("2020-07-18 23:34:50"),
				},
			},
		},
		{
			name: "successfully remove control which has an expired impulse delay",
			document: &Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:           ccpb.ControlType_MCT_GAMBLING.String(),
						ControlEnabled:        true,
						ShouldDeclineAll:      false,
						ImpulseDelayPeriod:    util.ToStringPtr("48:00"),
						ImpulseDelayRemaining: util.ToStringPtr("00:00:00"),
						ImpulseDelayStart:     util.ToStringPtr("2020-05-18 23:34:50"),
						ImpulseDelayEnd:       util.ToStringPtr("2020-07-18 23:34:50"),
					},
				},
				DocumentID:          documentID,
				LastUpdateTimeStamp: time.Time{}.String(),
			},
			index:                0,
			want:                 true,
			wantMerchantControls: []*MerchantControl{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.document.removeGamblingControl(tt.index)
			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantMerchantControls, tt.document.MerchantControls)
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
	t.Run("test", func(t *testing.T) {
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
		want := &Request{
			GlobalControls: []*GlobalControl{
				{
					ShouldDeclineAll:                  true,
					ControlEnabled:                    true,
					UserIdentifier:                    "UID",
					AlertThreshold:                    util.ToFloat64Ptr(15),
					DeclineThreshold:                  util.ToFloat64Ptr(0),
					DeclineAllNonTokenizeTransactions: true,
					ShouldAlertOnDecline:              false,
				},
			},
			MerchantControls: []*MerchantControl{
				{
					ShouldDeclineAll:     true,
					ControlEnabled:       true,
					UserIdentifier:       "UID",
					ControlType:          ccpb.ControlType_MCT_GAMBLING.String(),
					AlertThreshold:       util.ToFloat64Ptr(15),
					DeclineThreshold:     util.ToFloat64Ptr(0),
					ShouldAlertOnDecline: false,
				},
			},
			TransactionControls: []*TransactionControl{
				{
					ShouldDeclineAll:     true,
					ControlEnabled:       true,
					ControlType:          ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
					UserIdentifier:       "UID",
					AlertThreshold:       util.ToFloat64Ptr(15),
					DeclineThreshold:     util.ToFloat64Ptr(0),
					ShouldAlertOnDecline: false,
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

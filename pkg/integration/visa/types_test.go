package visa

import (
	"testing"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/stretchr/testify/assert"
)

func Test_Resource_FindControlByType(t *testing.T) {
	type args struct {
		controlDocument Resource
		controlType     ccpb.ControlType
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "find global controls",
			args: args{
				controlDocument: Resource{
					GlobalControls: []*GlobalControl{
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlEnabled: true,
						},
					},
				},
				controlType: ccpb.ControlType_GCT_GLOBAL,
			},
			want: 0,
		},
		{
			name: "find merchant controls",
			args: args{
				controlDocument: Resource{
					MerchantControls: []*MerchantControl{
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_MCT_ALCOHOL.String(),
							ControlEnabled: true,
						},
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_MCT_AIRFARE.String(),
							ControlEnabled: true,
						},
					},
				},
				controlType: ccpb.ControlType_MCT_AIRFARE,
			},
			want: 1,
		},
		{
			name: "can not find merchant controls",
			args: args{
				controlDocument: Resource{
					MerchantControls: []*MerchantControl{
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_MCT_ALCOHOL.String(),
							ControlEnabled: true,
						}, {
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_MCT_AIRFARE.String(),
							ControlEnabled: true,
						},
					},
				},
				controlType: ccpb.ControlType_MCT_AUTOMOTIVE,
			},
			want: -1,
		},
		{
			name: "find transaction controls",
			args: args{
				controlDocument: Resource{
					TransactionControls: []*TransactionControl{
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
							ControlEnabled: true,
						},
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_TCT_AUTO_PAY.String(),
							ControlEnabled: true,
						},
					},
				},
				controlType: ccpb.ControlType_TCT_AUTO_PAY,
			},
			want: 1,
		},
		{
			name: "can not find transaction controls",
			args: args{
				controlDocument: Resource{
					MerchantControls: []*MerchantControl{
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_TCT_ATM_WITHDRAW.String(),
							ControlEnabled: true,
						},
						{
							AlertThreshold: util.ToFloat64Ptr(0),
							ControlType:    ccpb.ControlType_TCT_AUTO_PAY.String(),
							ControlEnabled: true,
						},
					},
				},
				controlType: ccpb.ControlType_TCT_CONTACTLESS,
			},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.args.controlDocument.FindControlByType(tt.args.controlType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Control_GetImpulseDelayStartTimestamp(t *testing.T) {
	tests := []struct {
		name            string
		controlDocument Resource
		want            *timestamppb.Timestamp
	}{
		{
			name: "successfully get Start Timestamp",
			controlDocument: Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:       ccpb.ControlType_MCT_GAMBLING.String(),
						ImpulseDelayStart: func(s string) *string { return &s }("2020-05-18 23:34:50"),
					},
				},
			},
			want: &timestamppb.Timestamp{
				Seconds: 1589844890,
			},
		},
		{
			name: "successfully handle nil start",
			controlDocument: Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:       ccpb.ControlType_MCT_GAMBLING.String(),
						ImpulseDelayStart: nil,
					},
				},
			},
			want: nil,
		},
		{
			name: "successfully handle invalid time",
			controlDocument: Resource{
				MerchantControls: []*MerchantControl{
					{
						ControlType:       ccpb.ControlType_MCT_GAMBLING.String(),
						ImpulseDelayStart: func(s string) *string { return &s }("1888-05-18 26:34:50"),
					},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.controlDocument.MerchantControls[0].GetImpulseDelayStartTimestamp()
			assert.Equal(t, expected, tt.want)
		})
	}
}

func Test_Control_GetImpulseDelayPeriodProto(t *testing.T) {
	tests := []struct {
		name            string
		controlDocument Resource
		want            *durationpb.Duration
	}{
		{
			name: "successfully get period duration",
			controlDocument: Resource{
				MerchantControls: []*MerchantControl{
					{
						ImpulseDelayPeriod: func(s string) *string { return &s }("24:00"),
					},
				},
			},
			want: &durationpb.Duration{
				Seconds: 86400,
			},
		},
		{
			name: "successfully handle nil period",
			controlDocument: Resource{
				MerchantControls: []*MerchantControl{
					{
						ImpulseDelayPeriod: nil,
					},
				},
			},
			want: nil,
		},
		{
			name: "successfully handle invalid period",
			controlDocument: Resource{
				MerchantControls: []*MerchantControl{
					{
						ImpulseDelayPeriod: func(s string) *string { return &s }("26-61"),
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.controlDocument.MerchantControls[0].GetImpulseDelayPeriodProto()
			assert.Equal(t, expected, tt.want)
		})
	}
}

func TestResource_Enrolled(t *testing.T) {
	tests := []struct {
		name     string
		resource *Resource
		want     bool
	}{
		{
			name: "not enrolled",
			resource: &Resource{
				DocumentID: "NOT_ENROLLED",
			},
			want: false,
		},
		{
			name: "enrolled",
			resource: &Resource{
				DocumentID: "anythingelse",
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.resource.Enrolled())
		})
	}
}

func TestCategory_String(t *testing.T) {
	tests := []struct {
		name  string
		given Category
		want  string
	}{
		{
			name:  "Transaction",
			given: TRANSACTION,
			want:  "TRANSACTION",
		}, {
			name:  "Merchant",
			given: MERCHANT,
			want:  "MERCHANT",
		}, {
			name:  "Global",
			given: GLOBAL,
			want:  "GLOBAL",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.given.String())
		})
	}
}

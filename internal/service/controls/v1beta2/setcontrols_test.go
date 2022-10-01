package v1beta2

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/fixtures"
	"github.com/anzx/fabric-cards/test/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/pkg/errors"
)

const controlType = ccpb.ControlType_TCT_CONTACTLESS

func TestSetAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.SetVisaControl{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}
		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogHook(hook)
		request := &ccpb.SetControlsRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			CardControls: []*ccpb.ControlRequest{
				{
					ControlType: ccpb.ControlType_GCT_GLOBAL,
				},
				{
					ControlType: controlType,
				},
			},
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		_, _ = s.SetControls(ctx, request)

		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

func TestSetControls(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.SetControlsRequest
		want    *ccpb.CardControlResponse
		wantErr error
	}{
		{
			name:    "successful set request",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
					{
						ControlType: ccpb.ControlType_MCT_GAMBLING,
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: controlType,
					},
					{
						ControlType: ccpb.ControlType_MCT_GAMBLING,
					},
				},
			},
		},
		{
			name:    "audit log failure does not affect response",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
					{
						ControlType: ccpb.ControlType_MCT_GAMBLING,
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: controlType,
					},
					{
						ControlType: ccpb.ControlType_MCT_GAMBLING,
					},
				},
			},
		},
		{
			name:    "unable to call Entitlements",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set control failed, reason=service unavailable"),
		},
		{
			name:    "unable to query controls",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayListError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set control failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "unable to create control query",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayCreateError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set control failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "not enrolled but successfully resolved",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetNotEnrolled))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: controlType,
					},
				},
			},
		},
		{
			name:    "not enrolled, unable to create controls",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayCreateError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set control failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "not enrolled, unable to get control document",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetCanNotBeEnrolled))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=NotFound, error_code=2, message=set control failed, reason=no control document found"),
		},
		{
			name:    "unable to enrol by pan",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetNotEnrolled))).WithVisaGatewayRegistrationError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set control failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "unable to verify ownership",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: token,
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=set control failed, reason=user not entitled"),
		},
		{
			name:    "unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=set control failed, reason=card not eligible"),
		},
		{
			name:    "unable to create duplicated control",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(controlType),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=AlreadyExists, error_code=25000, message=set control failed, reason=control already exists"),
		},
		{
			name:    "unable to add one of the control types",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.SetControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
					{
						ControlType: controlType,
					},
				},
			},
			wantErr: errors.New("control is disabled"),
		},
	}
	err := feature.FeatureGate.Set(map[feature.Feature]bool{
		feature.TCT_CONTACTLESS: true,
		feature.MCT_GAMBLING:    true,
		feature.GCT_GLOBAL:      false,
	})
	require.NoError(t, err)
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardControlsServer(test.builder)
			got, err := s.SetControls(ctx, test.req)
			if got != nil {
				sort.Slice(got.CardControls, func(i, j int) bool {
					return got.CardControls[i].GetControlType() < got.CardControls[j].GetControlType()
				})
			}
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

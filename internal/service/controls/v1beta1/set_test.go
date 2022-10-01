package v1beta1

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/feature"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/util"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	"github.com/pkg/errors"
)

const controlType = ccpb.ControlType_TCT_CONTACTLESS

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.SetRequest
		want    *ccpb.CardControlResponse
		wantErr error
	}{
		{
			name:    "successful set request",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.SetRequest{
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
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    controlType,
						ControlEnabled: true,
					},
					{
						ControlType:    ccpb.ControlType_MCT_GAMBLING,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name:    "audit log failure does not affect response",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetRequest{
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
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    controlType,
						ControlEnabled: true,
					},
					{
						ControlType:    ccpb.ControlType_MCT_GAMBLING,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name:    "unable to call Entitlements",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetRequest{
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
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcQueryError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetRequest{
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
			name:    "unable to create control query",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcCreateError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetRequest{
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
			name:    "not enrolled but successfully resolved",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetNotEnrolled))),
			req: &ccpb.SetRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			want: &ccpb.CardControlResponse{
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    controlType,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name:    "not enrolled, unable to create controls",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcCreateError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.ControlRequest{
					{
						ControlType: controlType,
					},
				},
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set control failed, reason=service unavailable"),
		},
		{
			name:    "not enrolled, unable to get control document",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetCanNotBeEnrolled))),
			req: &ccpb.SetRequest{
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
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetNotEnrolled))).WithVtcEnrolError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.SetRequest{
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
			name:    "unable to verify ownership",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.SetRequest{
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
			req: &ccpb.SetRequest{
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
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcControls(controlType),
			req: &ccpb.SetRequest{
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
			req: &ccpb.SetRequest{
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardControlsServer(test.builder)
			got, err := s.Set(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func TestSetAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.SetVisaControl{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}
		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogHook(hook)
		request := &ccpb.SetRequest{
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

		_, _ = s.Set(ctx, request)
		require.NoError(t, sd.Validate())
		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

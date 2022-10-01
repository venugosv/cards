package v1beta1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/anzx/fabric-cards/test/util"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/test/data"
	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

func TestRemove(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.RemoveRequest
		want    *ccpb.CardControlResponse
		wantErr error
	}{
		{
			name:    "Successfully remove single global control",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
			},
			want: &ccpb.CardControlResponse{
				CardControls: nil,
			},
		},
		{
			name:    "Successfully remove single global control",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithAuditLogError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
			},
			want: &ccpb.CardControlResponse{
				CardControls: nil,
			},
		},
		{
			name:    "Successfully remove single merchant control",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcControls(ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_ALCOHOL},
			},
			want: &ccpb.CardControlResponse{
				CardControls: nil,
			},
		},
		{
			name:    "Successfully remove single gambling control, control remains with impulse delay",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcControls(ccpb.ControlType_MCT_GAMBLING),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_GAMBLING},
			},
			want: &ccpb.CardControlResponse{
				CardControls: []*ccpb.CardControl{
					{
						ControlType:        ccpb.ControlType_MCT_GAMBLING,
						ControlEnabled:     true,
						ImpulseDelayStart:  &timestamppb.Timestamp{Seconds: 1589844890},
						ImpulseDelayPeriod: &durationpb.Duration{Seconds: 172800},
					},
				},
			},
		},
		{
			name:    "Successfully remove single transaction control",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcControls(ccpb.ControlType_TCT_ATM_WITHDRAW),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want: &ccpb.CardControlResponse{
				CardControls: nil,
			},
		},
		{
			name:    "Fail on Entitlements check",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=service unavailable"),
		},
		{
			name:    "successful on query response not enrolled",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetNotEnrolled))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want: &ccpb.CardControlResponse{},
		},
		{
			name:    "failed on query error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcQueryError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=service unavailable"),
		},
		{
			name:    "failed on invalid query response",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetCanNotBeEnrolled))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=NotFound, error_code=2, message=remove failed, reason=no control document found"),
		},
		{
			name:    "failed on invalid remove response",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetAllControls))).WithVtcUpdateError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    &ccpb.CardControlResponse{},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=service unavailable"),
		},
		{
			name:    "document not changed",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetAllControls))).WithVtcControls(ccpb.ControlType_GCT_GLOBAL),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
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
			name:    "unable to verify ownership",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVtcControls(ccpb.ControlType_GCT_GLOBAL),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.RandomUser().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=remove failed, reason=user not entitled"),
		},
		{
			name:    "unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))).WithVtcControls(ccpb.ControlType_GCT_GLOBAL),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			wantErr: errors.New("not eligible"),
		},
		{
			name:    "unable to tokenize card",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=service unavailable"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardControlsServer(test.builder)
			got, err := s.Remove(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func TestRemoveAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.RemoveVisaControl{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}

		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogHook(hook)
		request := &ccpb.RemoveRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			ControlTypes: []ccpb.ControlType{
				ccpb.ControlType_MCT_AIRFARE,
			},
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		_, _ = s.Remove(ctx, request)
		require.NoError(t, sd.Validate())
		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

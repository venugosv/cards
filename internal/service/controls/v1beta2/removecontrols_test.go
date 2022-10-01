package v1beta2

import (
	"context"
	"testing"

	pkgutil "github.com/anzx/fabric-cards/pkg/integration/util"
	"github.com/anzx/fabric-cards/test/util"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/test/data"
	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

func TestRemoveAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.RemoveVisaControl{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}

		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogHook(hook)
		request := &ccpb.RemoveControlsRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			ControlTypes: []ccpb.ControlType{
				ccpb.ControlType_MCT_AIRFARE,
			},
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		_, _ = s.RemoveControls(ctx, request)

		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

func TestRemoveControls(t *testing.T) {
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.RemoveControlsRequest
		want    *ccpb.CardControlResponse
		wantErr error
	}{
		{
			name:    "Successfully remove single global control",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls:        nil,
			},
		},
		{
			name:    "Successfully remove single global control with audit log",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithAuditLogError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls:        nil,
			},
		},
		{
			name:    "Successfully remove single transaction control",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_TCT_ATM_WITHDRAW),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls:        nil,
			},
		},
		{
			name:    "Successfully remove single merchant control not MCT_GAMBLING",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_ALCOHOL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls:        nil,
			},
		},
		{
			name:    "Gambling control with the same impulse remains after delete request",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_MCT_GAMBLING).WithVisaGatewayGamblingImpulse("2020/05/18 23:34:50", "12:00:00"),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_GAMBLING},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType:        ccpb.ControlType_MCT_GAMBLING,
						ImpulseDelayStart:  &timestamppb.Timestamp{Seconds: 1589844890},
						ImpulseDelayPeriod: &durationpb.Duration{Seconds: 172800},
					},
				},
			},
		},
		{
			name:    "Successfully remove gambling control with impulse delay expired",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_MCT_GAMBLING).WithVisaGatewayGamblingImpulse("2020/05/18 23:34:50", "00:00:00"),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_GAMBLING},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls:        nil,
			},
		},
		{
			name:    "Successfully remove single global control, while transaction control remains the same",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_TCT_CONTACTLESS),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_TCT_CONTACTLESS,
					},
				},
			},
		},
		{
			name:    "Successfully remove single transaction control, while merchant control remains the same",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_TCT_CONTACTLESS, ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_CONTACTLESS},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_MCT_ALCOHOL,
					},
				},
			},
		},
		{
			name:    "Successfully remove single merchant control, while global control the same",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_ALCOHOL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
				},
			},
		},
		{
			name:    "Successfully remove single transaction control, while gambling control remains with impulse delay",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_TCT_ATM_WITHDRAW, ccpb.ControlType_MCT_GAMBLING).WithVisaGatewayGamblingImpulse("2020-05-18 23:34:50", "12:00:00"),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType:        ccpb.ControlType_MCT_GAMBLING,
						ImpulseDelayStart:  &timestamppb.Timestamp{Seconds: 1589844890},
						ImpulseDelayPeriod: &durationpb.Duration{Seconds: 172800},
					},
				},
			},
		},
		{
			name:    "Successfully remove both global control and transaction control, while merchant control remains the same",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_TCT_CONTACTLESS, ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_TCT_CONTACTLESS},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_MCT_ALCOHOL,
					},
				},
			},
		},
		{
			name:    "Successfully remove both global control and merchant control, while transaction control remains the same",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_TCT_CONTACTLESS, ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_MCT_ALCOHOL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_TCT_CONTACTLESS,
					},
				},
			},
		},
		{
			name:    "Successfully remove both transaction control and merchant control, while global controls remains the same",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_TCT_CONTACTLESS, ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_CONTACTLESS, ccpb.ControlType_MCT_ALCOHOL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
				},
			},
		},
		{
			name:    "Successfully remove all controls",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_TCT_CONTACTLESS, ccpb.ControlType_MCT_ALCOHOL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL, ccpb.ControlType_TCT_CONTACTLESS, ccpb.ControlType_MCT_ALCOHOL},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls:        nil,
			},
		},
		{
			name:    "Fail on Entitlements check",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=service unavailable"),
		},
		{
			name:    "failed on query error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayListError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "failed on invalid query response",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetCanNotBeEnrolled))),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    nil,
			wantErr: errors.New("fabric error: status_code=NotFound, error_code=2, message=remove failed, reason=no control document found"),
		},
		{
			name:    "successful on query response not enrolled",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetNotEnrolled))),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want: &ccpb.CardControlResponse{},
		},
		{
			name:    "failed on invalid remove response",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_TCT_ATM_WITHDRAW).WithVisaGatewayDeleteError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want:    &ccpb.CardControlResponse{},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "document not changed",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
				},
			},
		},
		{
			name:    "unable to verify ownership",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.RandomUser().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=remove failed, reason=user not entitled"),
		},
		{
			name:    "unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			wantErr: errors.New("not eligible"),
		},
		{
			name:    "unable to tokenize card",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.RemoveControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				ControlTypes:        []ccpb.ControlType{ccpb.ControlType_TCT_ATM_WITHDRAW},
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=remove failed, reason=service unavailable"),
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardControlsServer(test.builder)
			got, err := s.RemoveControls(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func Test_RemoveGamblingControl(t *testing.T) {
	t.Run("Gambling control remains with impulse just set after delete request", func(t *testing.T) {
		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_MCT_GAMBLING)
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		req := &ccpb.RemoveControlsRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_GAMBLING},
		}
		got, err := s.RemoveControls(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, got.CardControls)
		require.NotNil(t, got.CardControls[0].GetImpulseDelayPeriod())
	})
	t.Run("Gambling control requested with no gambling control active", func(t *testing.T) {
		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(ccpb.ControlType_GCT_GLOBAL)
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		req := &ccpb.RemoveControlsRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_GAMBLING},
		}
		got, err := s.RemoveControls(ctx, req)
		require.NoError(t, err)
		require.Equal(t, got.CardControls[0].ControlType, ccpb.ControlType_GCT_GLOBAL)
	})
	t.Run("Gambling control removal requested along with other active controls", func(t *testing.T) {
		controls := []ccpb.ControlType{ccpb.ControlType_MCT_GAMBLING, ccpb.ControlType_GCT_GLOBAL}
		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayControls(controls...)
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		req := &ccpb.RemoveControlsRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			ControlTypes:        []ccpb.ControlType{ccpb.ControlType_MCT_GAMBLING, ccpb.ControlType_GCT_GLOBAL},
		}
		got, err := s.RemoveControls(ctx, req)
		require.NoError(t, err)
		assert.Len(t, got.CardControls, 1)
		for _, control := range got.CardControls {
			if control.GetControlType() == ccpb.ControlType_MCT_GAMBLING {
				assert.NotNil(t, control.ImpulseDelayPeriod)
			}
		}
	})
}

func Test_isGamblingControlInRequest(t *testing.T) {
	t.Run("with gambling control type, return true", func(t *testing.T) {
		req := &ccpb.RemoveControlsRequest{
			ControlTypes: []ccpb.ControlType{
				ccpb.ControlType_GCT_GLOBAL,
				ccpb.ControlType_TCT_CONTACTLESS,
				ccpb.ControlType_MCT_GAMBLING,
			},
		}
		got := gamblingBlockRequested(req.ControlTypes)
		assert.True(t, got)
	})
	t.Run("without gambling control type, return false", func(t *testing.T) {
		req := &ccpb.RemoveControlsRequest{
			ControlTypes: []ccpb.ControlType{
				ccpb.ControlType_GCT_GLOBAL,
				ccpb.ControlType_TCT_CONTACTLESS,
			},
		}
		got := gamblingBlockRequested(req.ControlTypes)
		assert.False(t, got)
	})
}

func Test_getGamblingControlFromDocument(t *testing.T) {
	t.Run("get gambling control from document", func(t *testing.T) {
		req := []*crpb.MerchantControl{
			{
				ControlType:      "MCT_GAMBLING",
				IsControlEnabled: true,
			},
			{
				ControlType:      "MCT_ELECTRONICS",
				IsControlEnabled: true,
			},
		}
		want := &crpb.MerchantControl{
			ControlType:      "MCT_GAMBLING",
			IsControlEnabled: true,
		}

		got, ok := getGamblingControlFromDocument(req)
		assert.True(t, ok)
		assert.Equal(t, want, got)
	})
	t.Run("return false when gambling control not enabled", func(t *testing.T) {
		req := []*crpb.MerchantControl{
			{
				ControlType:      "MCT_GAMBLING",
				IsControlEnabled: false,
			},
			{
				ControlType:      "MCT_ELECTRONICS",
				IsControlEnabled: true,
			},
		}
		got, ok := getGamblingControlFromDocument(req)
		assert.False(t, ok)
		require.Nil(t, got)
	})
	t.Run("return false when no gambling control", func(t *testing.T) {
		req := []*crpb.MerchantControl{
			{
				ControlType:      "MCT_ELECTRONICS",
				IsControlEnabled: true,
			},
		}
		got, ok := getGamblingControlFromDocument(req)
		assert.False(t, ok)
		require.Nil(t, got)
	})
}

func Test_canRemoveGamblingControl(t *testing.T) {
	t.Run("active impulse delay, return false", func(t *testing.T) {
		c := &crpb.MerchantControl{
			ControlType:           "MCT_GAMBLING",
			ImpulseDelayPeriod:    pkgutil.ToStringPtr(fortyEightHours),
			ImpulseDelayStart:     pkgutil.ToStringPtr("2022-18-03 00:00"),
			ImpulseDelayEnd:       pkgutil.ToStringPtr("2022-20-03 00:00"),
			ImpulseDelayRemaining: pkgutil.ToStringPtr(fortyEightHours),
		}
		got := canRemoveGamblingControl(c)
		assert.False(t, got)
	})
	t.Run("expired impulse delay, return true", func(t *testing.T) {
		c := &crpb.MerchantControl{
			ControlType:           "MCT_GAMBLING",
			ImpulseDelayPeriod:    pkgutil.ToStringPtr(fortyEightHours),
			ImpulseDelayStart:     pkgutil.ToStringPtr("2022-18-03 00:00"),
			ImpulseDelayEnd:       pkgutil.ToStringPtr("2022-20-03 00:00"),
			ImpulseDelayRemaining: pkgutil.ToStringPtr(noTimeRemaining),
		}
		got := canRemoveGamblingControl(c)
		assert.True(t, got)
	})
	t.Run("impulse delay doesn't exit, return false", func(t *testing.T) {
		c := &crpb.MerchantControl{
			ControlType: "MCT_GAMBLING",
		}
		got := canRemoveGamblingControl(c)
		assert.False(t, got)
	})
	t.Run("", func(t *testing.T) {
		c := &crpb.MerchantControl{
			UserIdentifier:        pkgutil.ToStringPtr("33f4a76b-b661-4b13-a4fd-646763cc9594"),
			ControlType:           "MCT_GAMBLING",
			ImpulseDelayEnd:       pkgutil.ToStringPtr("2022-03-20 04:13:33"),
			ImpulseDelayRemaining: pkgutil.ToStringPtr("00:00:00"),
			IsControlEnabled:      true,
			ImpulseDelayPeriod:    pkgutil.ToStringPtr("00:01"),
			ImpulseDelayStart:     pkgutil.ToStringPtr("2022-03-20 04:12:33"),
			ShouldDeclineAll:      pkgutil.ToBoolPtr(true),
		}
		assert.True(t, canRemoveGamblingControl(c))
	})
}

func Test_impulseDelayExists(t *testing.T) {
	t.Run("impulse delay exists, return true", func(t *testing.T) {
		c := &crpb.MerchantControl{
			ControlType:           "MCT_GAMBLING",
			ImpulseDelayPeriod:    pkgutil.ToStringPtr(fortyEightHours),
			ImpulseDelayStart:     pkgutil.ToStringPtr("2022-18-03 00:00"),
			ImpulseDelayEnd:       pkgutil.ToStringPtr("2022-20-03 00:00"),
			ImpulseDelayRemaining: pkgutil.ToStringPtr(fortyEightHours),
		}
		got := impulseDelayExists(c)
		assert.True(t, got)
	})
	t.Run("impulsed delay doesn't exist, return false", func(t *testing.T) {
		c := &crpb.MerchantControl{
			ControlType: "MCT_GAMBLING",
		}
		got := impulseDelayExists(c)
		assert.False(t, got)
	})
}

func Test_checkImpulseDelayActive(t *testing.T) {
	t.Run("active impulse delay, return true", func(t *testing.T) {
		c := &crpb.MerchantControl{
			ControlType:           "MCT_GAMBLING",
			ImpulseDelayPeriod:    pkgutil.ToStringPtr(fortyEightHours),
			ImpulseDelayStart:     pkgutil.ToStringPtr("2022-18-03 00:00"),
			ImpulseDelayEnd:       pkgutil.ToStringPtr("2022-20-03 00:00"),
			ImpulseDelayRemaining: pkgutil.ToStringPtr(fortyEightHours),
		}
		got := impulseDelayActive(c)
		assert.True(t, got)
	})
	t.Run("expired impulse delay, return false", func(t *testing.T) {
		c := &crpb.MerchantControl{
			ControlType:           "MCT_GAMBLING",
			ImpulseDelayPeriod:    pkgutil.ToStringPtr(fortyEightHours),
			ImpulseDelayStart:     pkgutil.ToStringPtr("2022-18-03 00:00"),
			ImpulseDelayEnd:       pkgutil.ToStringPtr("2022-20-03 00:00"),
			ImpulseDelayRemaining: pkgutil.ToStringPtr(noTimeRemaining),
		}
		got := impulseDelayActive(c)
		assert.False(t, got)
	})
}

package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/test/data"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"

	"github.com/stretchr/testify/assert"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

func TestServer_List(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		want    *ccpb.ListResponse
		wantErr error
	}{
		{
			name:    "successful call to list controls with a single card number",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))),
			want: &ccpb.ListResponse{
				CardControls: map[string]*ccpb.CardControlResponse{
					data.AUserWithACard().Token(): {
						CardControls: []*ccpb.CardControl{
							{
								ControlType:    ccpb.ControlType_GCT_GLOBAL,
								ControlEnabled: true,
							},
						},
					},
				},
			},
		},
		{
			name: "successful call to list controls with a multiple card numbers",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithACard(
						data.WithAToken("token1"),
						data.WithACardNumber("cardNumber1"),
						data.WithControls(data.CardControlsPresetGlobalControls),
					),
					data.WithACard(
						data.WithAToken("token2"),
						data.WithACardNumber("cardNumber2"),
						data.WithControls(data.CardControlsPresetContactlessControl),
					),
				)),
			want: &ccpb.ListResponse{
				CardControls: map[string]*ccpb.CardControlResponse{
					"token1": {
						CardControls: []*ccpb.CardControl{
							{
								ControlType:    ccpb.ControlType_GCT_GLOBAL,
								ControlEnabled: true,
							},
						},
					},
					"token2": {
						CardControls: []*ccpb.CardControl{
							{
								ControlType:    ccpb.ControlType_TCT_CONTACTLESS,
								ControlEnabled: true,
							},
						},
					},
				},
			},
		},
		{
			name: "successful call to list controls with a multiple card numbers, one not enrolled",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithACard(
						data.WithAToken("token1"),
						data.WithACardNumber("cardNumber1"),
						data.WithControls(data.CardControlsPresetGlobalControls),
					),
					data.WithACard(
						data.WithAToken("token2"),
						data.WithACardNumber("cardNumber2"),
						data.WithControls(data.CardControlsPresetNotEnrolled),
					),
				)),
			want: &ccpb.ListResponse{
				CardControls: map[string]*ccpb.CardControlResponse{
					"token1": {
						CardControls: []*ccpb.CardControl{
							{
								ControlType:    ccpb.ControlType_GCT_GLOBAL,
								ControlEnabled: true,
							},
						},
					},
					"token2": {},
				},
			},
		},
		{
			name: "successful call to list controls with a multiple card numbers, one not eligible",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithACard(
						data.WithAToken("token1"),
						data.WithACardNumber("cardNumber1"),
						data.WithControls(data.CardControlsPresetGlobalControls),
					),
					data.WithACard(
						data.WithAToken("token2"),
						data.WithACardNumber("cardNumber2"),
						data.WithStatus(ctm.StatusStolen),
					),
				)),
			want: &ccpb.ListResponse{
				CardControls: map[string]*ccpb.CardControlResponse{
					"token1": {
						CardControls: []*ccpb.CardControl{
							{
								ControlType:    ccpb.ControlType_GCT_GLOBAL,
								ControlEnabled: true,
							},
						},
					},
				},
			},
		},
		{
			name:    "Invalid ENT call",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithEntListError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=list controls failed, reason=service unavailable"),
		},
		{
			name:    "Invalid VTC call",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVtcQueryError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			want: &ccpb.ListResponse{
				CardControls: map[string]*ccpb.CardControlResponse{
					"6688390512341000": {},
				},
			},
		},
		{
			name:    "Unable to verify entitlements",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			want: &ccpb.ListResponse{
				CardControls: map[string]*ccpb.CardControlResponse{},
			},
		},
		{
			name:    "Unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))),
			want: &ccpb.ListResponse{
				CardControls: map[string]*ccpb.CardControlResponse{},
			},
		},
		{
			name:    "Unable to detokenize cards",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=list controls failed, reason=service unavailable"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildCardControlsServer(tt.builder)
			got, err := s.List(fixtures.GetTestContext(), nil)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

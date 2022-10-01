package v1beta2

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"

	"github.com/anz-bank/equals"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
)

const (
	token1      = "3930000046220001"
	cardNumber1 = "4622393000000001"
	token2      = "3930000046220002"
	cardNumber2 = "4622393000000002"
)

func TestServer_ListControls(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		want    *ccpb.ListControlsResponse
		wantErr error
	}{
		{
			name:    "successful call to list controls with a single card number",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))),
			want: &ccpb.ListControlsResponse{
				CardControls: []*ccpb.CardControlResponse{
					{
						TokenizedCardNumber: data.AUserWithACard().Token(),
						CardControls: []*ccpb.CardControl{
							{
								ControlType: ccpb.ControlType_GCT_GLOBAL,
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
						data.WithAToken(token1),
						data.WithACardNumber(cardNumber1),
						data.WithControls(data.CardControlsPresetGlobalControls),
					),
					data.WithACard(
						data.WithAToken(token2),
						data.WithACardNumber(cardNumber2),
						data.WithControls(data.CardControlsPresetContactlessControl),
					),
				),
			),
			want: &ccpb.ListControlsResponse{
				CardControls: []*ccpb.CardControlResponse{
					{
						TokenizedCardNumber: token1,
						CardControls: []*ccpb.CardControl{
							{
								ControlType: ccpb.ControlType_GCT_GLOBAL,
							},
						},
					},
					{
						TokenizedCardNumber: token2,
						CardControls: []*ccpb.CardControl{
							{
								ControlType: ccpb.ControlType_TCT_CONTACTLESS,
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
						data.WithAToken(token1),
						data.WithACardNumber(cardNumber1),
						data.WithControls(data.CardControlsPresetGlobalControls),
					),
					data.WithACard(
						data.WithAToken(token2),
						data.WithACardNumber(cardNumber2),
						data.WithControls(data.CardControlsPresetNotEnrolled),
					),
				),
			),
			want: &ccpb.ListControlsResponse{
				CardControls: []*ccpb.CardControlResponse{
					{
						TokenizedCardNumber: token1,
						CardControls: []*ccpb.CardControl{
							{
								ControlType: ccpb.ControlType_GCT_GLOBAL,
							},
						},
					},
					{
						TokenizedCardNumber: token2,
					},
				},
			},
		},
		{
			name: "successful call to list controls with a multiple card numbers, one not eligible",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithACard(
						data.WithAToken(token1),
						data.WithACardNumber(cardNumber1),
						data.WithControls(data.CardControlsPresetGlobalControls),
					),
					data.WithACard(
						data.WithAToken(token2),
						data.WithACardNumber(cardNumber2),
						data.WithStatus(ctm.StatusStolen),
					),
				),
			),
			want: &ccpb.ListControlsResponse{
				CardControls: []*ccpb.CardControlResponse{
					{
						TokenizedCardNumber: token1,
						CardControls: []*ccpb.CardControl{
							{
								ControlType: ccpb.ControlType_GCT_GLOBAL,
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
			name:    "Invalid VisaGateway call",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVisaGatewayListError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			want: &ccpb.ListControlsResponse{
				CardControls: []*ccpb.CardControlResponse{
					{
						TokenizedCardNumber: "6688390512341000",
					},
				},
			},
		},
		{
			name:    "Unable to verify entitlements",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			want: &ccpb.ListControlsResponse{
				CardControls: []*ccpb.CardControlResponse{},
			},
		},
		{
			name:    "Unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))),
			want: &ccpb.ListControlsResponse{
				CardControls: []*ccpb.CardControlResponse{},
			},
		},
		{
			name:    "Unable to detokenize cards",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=list controls failed, reason=service unavailable"),
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardControlsServer(test.builder)
			fmt.Println(test.want)
			got, err := s.ListControls(fixtures.GetTestContext(), nil)
			fmt.Println(got)
			if test.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.wantErr.Error())
			} else {
				assert.Nil(t, err)
				sort.Slice(got.CardControls, func(i, j int) bool {
					return got.CardControls[i].GetTokenizedCardNumber() < got.CardControls[j].GetTokenizedCardNumber()
				})
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

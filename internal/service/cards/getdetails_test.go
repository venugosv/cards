package cards

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/feature"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	pbtype "github.com/anzx/fabricapis/pkg/fabric/type"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestGetCardDetails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		builder     *fixtures.ServerBuilder
		cardToken   string
		want        *cpb.GetDetailsResponse
		wantErr     error
		enableDCVV2 bool
	}{
		{
			name: "Entitlements fails, get card details return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			cardToken: data.AUserWithACard().Token(),
			wantErr:   errors.New("fabric error: status_code=Unavailable, error_code=2, message=get details failed, reason=service unavailable"),
		},
		{
			name: "Entitlements false, get card details return error",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithAPersonaID("personaID"),
					data.WithACard(
						data.WithAToken("token"),
						data.WithACardNumber("1234567890123456")))),
			cardToken: data.RandomUser().Token(),
			wantErr:   errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=get details failed, reason=user not entitled"),
		},
		{
			name:      "Vault fails get card details return error",
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(errors.New("oh no!")),
			cardToken: data.AUserWithACard().Token(),
			wantErr:   errors.New("fabric error: status_code=Internal, error_code=20004, message=get details failed, reason=service unavailable"),
		},
		{
			name: "CTM fails get card details return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmInquiryError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			cardToken: data.AUserWithACard().Token(),
			wantErr:   errors.New("fabric error: status_code=Unavailable, error_code=2, message=get details failed, reason=service unavailable"),
		},
		{
			name:      "Card inactive and therefore ineligible",
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.Inactive)),
			cardToken: data.AUserWithACard().Token(),
			wantErr:   errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=get details failed, reason=card not eligible"),
		},
		{
			name:      "Everything works return expected result",
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.Active)),
			cardToken: data.AUserWithACard().Token(),
			want: &cpb.GetDetailsResponse{
				Name:       "NATHAN FUKUSHIMA",
				CardNumber: data.AUserWithACard().CardNumber(),
				Cvc:        "",
				ExpiryDate: &pbtype.Date{
					Year:  &pbtype.OptionalInt32{Value: 2017},
					Month: &pbtype.OptionalInt32{Value: 5},
				},
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
					epb.Eligibility_ELIGIBILITY_BLOCK,
					epb.Eligibility_ELIGIBILITY_GET_DETAILS,
				},
			},
		},
		{
			name:      "card stolen and no longer eligible",
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))),
			cardToken: data.AUserWithACard().Token(),
			want:      &cpb.GetDetailsResponse{},
			wantErr:   fmt.Errorf("fabric error: status_code=PermissionDenied, error_code=20002, message=get details failed, reason=card not eligible"),
		},
		{
			name:      "dcvv2 fails",
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.Active)).WithDCVV2GenerateError(errors.New("oh no")),
			cardToken: data.AUserWithACard().Token(),
			want: &cpb.GetDetailsResponse{
				Name:       "NATHAN FUKUSHIMA",
				CardNumber: data.AUserWithACard().CardNumber(),
				ExpiryDate: &pbtype.Date{
					Year:  &pbtype.OptionalInt32{Value: 2017},
					Month: &pbtype.OptionalInt32{Value: 5},
				},
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
					epb.Eligibility_ELIGIBILITY_BLOCK,
					epb.Eligibility_ELIGIBILITY_GET_DETAILS,
				},
			},
			enableDCVV2: true,
		},
		{
			name:      "Everything works return expected result including dcvv2",
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.Active)),
			cardToken: data.AUserWithACard().Token(),
			want: &cpb.GetDetailsResponse{
				Name:       "NATHAN FUKUSHIMA",
				CardNumber: data.AUserWithACard().CardNumber(),
				ExpiryDate: &pbtype.Date{
					Year:  &pbtype.OptionalInt32{Value: 2017},
					Month: &pbtype.OptionalInt32{Value: 5},
				},
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
					epb.Eligibility_ELIGIBILITY_BLOCK,
					epb.Eligibility_ELIGIBILITY_GET_DETAILS,
				},
			},
			enableDCVV2: true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.enableDCVV2 {
				require.NoError(t, feature.FeatureGate.Set(map[feature.Feature]bool{
					feature.DCVV2: true,
				}))
			}

			s := buildCardServer(tt.builder)

			ctx := fixtures.GetTestContext()

			got, err := s.GetDetails(ctx, &cpb.GetDetailsRequest{TokenizedCardNumber: tt.cardToken})
			if test.wantErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, test.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.ExpiryDate, got.ExpiryDate)
				for _, expected := range tt.want.Eligibilities {
					assert.Contains(t, got.Eligibilities, expected)
				}
			}
		})
	}
}

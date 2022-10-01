package cards

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/date"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/test/fixtures"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestAuditTrail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		cardToken string
		want      *cpb.AuditTrailResponse
		wantErr   error
	}{
		{
			name: "Entitlements fails audit trail return error",
			builder: fixtures.AServer().WithData().WithEntMayError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			cardToken: data.AUserWithACard().Token(),
			wantErr:   errors.New("fabric error: status_code=Unavailable, error_code=2, message=audit trail failed, reason=service unavailable"),
		},
		{
			name: "Entitlements false audit trail return not entitled",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithAPersonaID("personaID"),
					data.WithACard(
						data.WithAToken("token"),
						data.WithACardNumber("1234567890123456")))),
			cardToken: data.RandomUser().Token(),
			wantErr:   errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=audit trail failed, reason=user not entitled"),
		},
		{
			name: "CTM fails audit trail return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmInquiryError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			cardToken: data.AUserWithACard().Token(),
			want:      nil,
			wantErr:   errors.New("fabric error: status_code=Unavailable, error_code=2, message=audit trail failed, reason=service unavailable"),
		},
		{
			name:      "Everything works return expected result",
			builder:   fixtures.AServer().WithData(data.AUserWithACard()),
			cardToken: data.AUserWithACard().Token(),
			want: &cpb.AuditTrailResponse{
				AccountsLinked:        2,
				TotalCards:            1,
				Activated:             true,
				CardControlEnabled:    false,
				MerchantUpdateEnabled: true,
				ReplacedDate:          nil,
				ReplacementCount:      0,
				IssueDate:             date.NewDate(2015, 8, 5).ToProto(),
				ReissueDate:           nil,
				ExpiryDate:            date.NewDate(2017, 5, 0).ToProto(),
				PreviousExpiryDate:    nil,
				DetailsChangedDate:    date.NewDate(2015, 8, 5).ToProto(),
				ClosedDate:            nil,
				Limits: []*cpb.Limit{
					{
						DailyLimit:          "1000",
						DailyLimitAvailable: "1000",
						LastTransaction:     nil,
						Type:                "APO",
					},
					{
						DailyLimit:          "2500",
						DailyLimitAvailable: "2347",
						LastTransaction:     date.NewDate(2015, 8, 5).ToProto(),
						Type:                "ATMEFTPOS",
					},
				},
				PinChangeDate:     date.NewDate(2015, 8, 5).ToProto(),
				PinChangeCount:    1,
				LastPinFailed:     nil,
				PinFailedCount:    0,
				Status:            "Issued",
				StatusChangedDate: date.NewDate(2015, 8, 5).ToProto(),
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardServer(test.builder)
			got, err := s.AuditTrail(fixtures.GetTestContext(), &cpb.AuditTrailRequest{TokenizedCardNumber: test.cardToken})
			if test.wantErr != nil {
				assert.NotNil(t, err)
				if err != nil {
					assert.Equal(t, test.wantErr.Error(), err.Error())
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestGetMaskedCardNumber(t *testing.T) {
	t.Parallel()
	t.Run("successfully return masked card", func(t *testing.T) {
		cardNumber := data.AUserWithACard().CardNumber()
		card := &ctm.Card{
			Token:       data.AUserWithACard().Token(),
			Last4Digits: cardNumber[len(cardNumber)-4:],
		}
		want := &cpb.MaskedCard{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			Last_4Digits:        cardNumber[len(cardNumber)-4:],
		}
		got := getMaskedCard(card)
		assert.Equal(t, want, got)
	})
	t.Run("nil card number", func(t *testing.T) {
		got := getMaskedCard(nil)
		assert.Nil(t, got)
	})
}

package cards

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/pkg/date"

	"github.com/anzx/fabric-cards/pkg/integration/cardcontrols"

	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"

	"github.com/anzx/fabric-cards/pkg/integration/selfservice"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/dcvv2"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"google.golang.org/grpc/codes"

	pbtype "github.com/anzx/fabricapis/pkg/fabric/type"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/test/fixtures"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		want    *cpb.ListResponse
		wantErr error
	}{
		{
			name: "Entitlements fails get cards return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntListError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			want:    &cpb.ListResponse{},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=list cards failed, reason=service unavailable"),
		},
		{
			name: "Entitlements false, get cards return error",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithAPersonaID("personaID"),
					data.WithACard(
						data.WithAToken("token"),
						data.WithACardNumber("1234567890123456")))),
			want:    &cpb.ListResponse{},
			wantErr: errors.New("fabric error: status_code=NotFound, error_code=2, message=list cards failed, reason=cannot find user data"),
		},
		{
			name: "CTM fails get cards return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithCtmInquiryError(anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=list cards failed, reason=service unavailable"),
		},
		{
			name:    "CTM succeeds get cards return true",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			want: &cpb.ListResponse{
				Cards: []*cpb.Card{
					{
						Name:                "MR NATHAN FUKUSHIMA",
						TokenizedCardNumber: data.AUserWithACard().Token(),
						Last_4Digits:        data.AUserWithACard().CardNumber()[len(data.AUserWithACard().CardNumber())-4:],
						Status:              "Issued",
						ExpiryDate: &pbtype.Date{
							Year:  &pbtype.OptionalInt32{Value: 2017},
							Month: &pbtype.OptionalInt32{Value: 5},
						},
						AccountNumbers: []string{"1234567890"},
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
							epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
						},
						Wallets: &cpb.Wallets{
							Other:      0,
							Fitness:    0,
							ApplePay:   2,
							ECommerce:  0,
							SamsungPay: 0,
							GooglePay:  1,
						},
					},
				},
			},
		},
		{
			name: "return visible card only, no error",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithACard(
						data.WithACardNumber(data.AUserWithACard().CardNumber()),
						data.WithAToken(data.AUserWithACard().Token()),
						data.WithStatus(ctm.StatusIssued),
						data.Active),
					data.WithACard(
						data.WithACardNumber(data.RandomCardNumber()),
						data.WithAToken("stolenToken"),
						data.WithStatus(ctm.StatusStolen),
						data.WithNewCard(data.AUserWithACard().CardNumber(), data.AUserWithACard().Token()),
						data.Active),
				),
			),
			want: &cpb.ListResponse{
				Cards: []*cpb.Card{
					{
						Name:                "MR NATHAN FUKUSHIMA",
						TokenizedCardNumber: data.AUserWithACard().Token(),
						Last_4Digits:        data.AUserWithACard().CardNumber()[len(data.AUserWithACard().CardNumber())-4:],
						Status:              "Issued",
						ExpiryDate: &pbtype.Date{
							Year:  &pbtype.OptionalInt32{Value: 2017},
							Month: &pbtype.OptionalInt32{Value: 5},
						},
						AccountNumbers: []string{"1234567890"},
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
							epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
						},
						Wallets: &cpb.Wallets{
							Other:      0,
							Fitness:    0,
							ApplePay:   2,
							ECommerce:  0,
							SamsungPay: 0,
							GooglePay:  1,
						},
					},
				},
			},
		},
		{
			name:    "User contains only 1 card, marked as stolen, new card number not returned in entitlements list response, stolen card is returned",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen), data.WithNewCardNumber(data.RandomCardNumber()))),
			want: &cpb.ListResponse{
				Cards: []*cpb.Card{
					{
						Name:                "MR NATHAN FUKUSHIMA",
						TokenizedCardNumber: data.AUserWithACard().Token(),
						Last_4Digits:        data.AUserWithACard().CardNumber()[len(data.AUserWithACard().CardNumber())-4:],
						Status:              "Stolen",
						ExpiryDate: &pbtype.Date{
							Year:  &pbtype.OptionalInt32{Value: 2017},
							Month: &pbtype.OptionalInt32{Value: 5},
						},
						AccountNumbers: []string{"1234567890"},
						Eligibilities: []epb.Eligibility{
							epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
						},
						Wallets: &cpb.Wallets{
							Other:      0,
							Fitness:    0,
							ApplePay:   2,
							ECommerce:  0,
							SamsungPay: 0,
							GooglePay:  1,
						},
					},
				},
			},
		},
		{
			name: "list returns 2 cards when a user has replaced a card as stolen and is yet to activate the new card",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithACard(
						data.WithStatus(ctm.StatusStolen),
						data.WithStatusReason(ctm.StatusReasonWithPinOrAccountRelated),
						data.WithAToken("stolenToken"),
						data.WithACardNumber("stolenCardNumber"),
						data.Active,
						data.WithNewCard("newCardNumber", "newToken")),
					data.WithACard(
						data.WithStatus(ctm.StatusIssued),
						data.WithAToken("newToken"),
						data.WithACardNumber("newCardNumber"),
						data.Inactive),
				),
			),
			want: &cpb.ListResponse{
				Cards: []*cpb.Card{
					{
						Name:                "MR NATHAN FUKUSHIMA",
						TokenizedCardNumber: "stolenToken",
						Last_4Digits:        "mber",
						Status:              "Stolen",
						ExpiryDate: &pbtype.Date{
							Year:  &pbtype.OptionalInt32{Value: 2017},
							Month: &pbtype.OptionalInt32{Value: 5},
						},
						AccountNumbers: []string{"1234567890"},
						Eligibilities: []epb.Eligibility{
							epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
						},
						NewTokenizedCardNumber: "newToken",
						Wallets: &cpb.Wallets{
							Other:      0,
							Fitness:    0,
							ApplePay:   2,
							ECommerce:  0,
							SamsungPay: 0,
							GooglePay:  1,
						},
					},
					{
						Name:                "MR NATHAN FUKUSHIMA",
						TokenizedCardNumber: "newToken",
						Last_4Digits:        "mber",
						Status:              "Issued",
						ExpiryDate: &pbtype.Date{
							Year:  &pbtype.OptionalInt32{Value: 2017},
							Month: &pbtype.OptionalInt32{Value: 5},
						},
						AccountNumbers: []string{"1234567890"},
						Eligibilities: []epb.Eligibility{
							epb.Eligibility_ELIGIBILITY_APPLE_PAY,
							epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
							epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
							epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION,
							epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
							epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
							epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
							epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
							epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
							epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
						},
						Wallets: &cpb.Wallets{
							Other:      0,
							Fitness:    0,
							ApplePay:   2,
							ECommerce:  0,
							SamsungPay: 0,
							GooglePay:  1,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardServer(test.builder)
			got, err := s.List(fixtures.GetTestContext(), &cpb.ListRequest{})
			if test.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Len(t, got.GetCards(), len(test.want.GetCards()))
				assert.ElementsMatch(t, got.GetCards(), test.want.GetCards())
			}
		})
	}
}

func buildCardServer(c *fixtures.ServerBuilder) cpb.CardAPIServer {
	fabric := Fabric{
		CommandCentre: &commandcentre.Client{
			Publisher: c.CommandCentreEnv,
		},
		Eligibility: &eligibility.Client{
			CardEligibilityAPIClient: c.CardEligibilityAPIClient,
		},
		Entitlements: &entitlements.Client{
			CardEntitlementsAPIClient:    c.CardEntitlementsAPIClient,
			EntitlementsControlAPIClient: c.EntitlementsControlAPIClient,
		},
		SelfService: &selfservice.Client{
			PartyAPIClient: c.SelfServiceClient,
		},
		DCVV2: &dcvv2.Client{
			DCVV2APIClient: c.DCVV2Client,
			ClientID:       "foobar",
		},
		CardControls: &cardcontrols.Client{
			CardControlsAPIClient: c.CardControlsClient,
		},
	}
	internal := Internal{
		RateLimit: c.RateLimit,
	}
	external := External{
		CTM:     c.CTMClient,
		Echidna: c.EchidnaClient,
		Vault:   c.VaultClient,
		AuditLog: &auditlogger.Client{
			Publisher: c.AuditLogPublisher,
		},
		OCV:       c.OCVClient,
		Forgerock: c.ForgerockClient,
	}
	return NewServer(fabric, internal, external)
}

func TestGetTimestamp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, input string
		format      date.Format
		want        *pbtype.Date
	}{
		{
			name:   "2015-08-01",
			format: date.YYYYMMDD,
			input:  "2015-08-01",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2015},
				Month: &pbtype.OptionalInt32{Value: 8},
				Day:   &pbtype.OptionalInt32{Value: 1},
			},
		},
		{
			name:   "2017-05-01",
			format: date.YYYYMMDD,
			input:  "2017-05-01",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 5},
				Day:   &pbtype.OptionalInt32{Value: 1},
			},
		},
		{
			name:   "2015-08",
			format: "2006-01",
			input:  "2015-08",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2015},
				Month: &pbtype.OptionalInt32{Value: 8},
			},
		},
		{
			name:   "201508",
			format: date.YYYYMM,
			input:  "201508",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2015},
				Month: &pbtype.OptionalInt32{Value: 8},
			},
		},
		{
			name:   "1508",
			format: date.YYMM,
			input:  "1508",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2015},
				Month: &pbtype.OptionalInt32{Value: 8},
			},
		},
		{
			name:   "1701",
			format: date.YYMM,
			input:  "1701",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 1},
			},
		},
		{
			name:   "1702",
			format: date.YYMM,
			input:  "1702",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 2},
			},
		},
		{
			name:   "1703",
			format: date.YYMM,
			input:  "1703",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 3},
			},
		},
		{
			name:   "1704",
			format: date.YYMM,
			input:  "1704",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 4},
			},
		},
		{
			name:   "1705",
			format: date.YYMM,
			input:  "1705",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 5},
			},
		},
		{
			name:   "1706",
			format: date.YYMM,
			input:  "1706",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 6},
			},
		},
		{
			name:   "1707",
			format: date.YYMM,
			input:  "1707",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 7},
			},
		},
		{
			name:   "1708",
			format: date.YYMM,
			input:  "1708",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 8},
			},
		},
		{
			name:   "1709",
			format: date.YYMM,
			input:  "1709",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 9},
			},
		},
		{
			name:   "1710",
			format: date.YYMM,
			input:  "1710",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 10},
			},
		},
		{
			name:   "1711",
			format: date.YYMM,
			input:  "1711",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 11},
			},
		},
		{
			name:   "1712",
			format: date.YYMM,
			input:  "1712",
			want: &pbtype.Date{
				Year:  &pbtype.OptionalInt32{Value: 2017},
				Month: &pbtype.OptionalInt32{Value: 12},
			},
		},
		{
			name:   "no input, nil output",
			format: date.YYMM,
			input:  "",
			want:   nil,
		},
		{
			name:   "bad format",
			format: date.YYMM,
			input:  "1234567",
			want:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := date.GetDate(context.Background(), test.format, test.input)
			assert.Equal(t, test.want, got)
		})
	}
}

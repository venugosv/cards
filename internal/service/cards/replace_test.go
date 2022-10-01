package cards

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	matchers "github.com/anzx/fabric-cards/pkg/integration/commandcentre/matchers"
	mock "github.com/anzx/fabric-cards/pkg/integration/commandcentre/mocks"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/notification"
	"github.com/golang/mock/gomock"

	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"

	"github.com/anzx/fabric-cards/pkg/integration/ocv"

	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"github.com/anz-bank/equals"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/test/util"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/test/fixtures"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	servicedata "github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

func TestReplaceCard(t *testing.T) {
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *cpb.ReplaceRequest
		want    *cpb.ReplaceResponse
		wantErr error
	}{
		{
			name: "Entitlements fails replacement return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(
				anzerrors.New(codes.Unavailable, "Entitlements/GetEntitledCard failed",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replacement failed, reason=service unavailable"),
		},
		{
			name: "Entitlements false, get card details return error",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithAPersonaID("personaID"),
					data.WithACard(
						data.WithAToken("token"),
						data.WithACardNumber("1234567890123456")))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.RandomUser().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=replacement failed, reason=user not entitled"),
		},
		{
			name: "CTM fails replacement return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmReplaceError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replacement failed, reason=service unavailable"),
		},
		{
			name: "Mailing address not available, residential address available, replacement return true",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithSelfServiceResponse(&sspb.GetPartyResponse{
				LegalName: &sspb.Name{
					Name:       "Ms. Oprah Gail Winfrey",
					Prefix:     "Queen",
					FirstName:  "Oprah",
					MiddleName: "Gail",
					LastName:   "Winfrey",
				},
				ResidentialAddress: &sspb.Address{
					LineOne:    "Level 13",
					LineTwo:    "839 Collins Street",
					City:       "Docklands",
					PostalCode: "3008",
					State:      "VIC",
					Country:    "AUS",
				},
			}),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			want: &cpb.ReplaceResponse{
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
				NewTokenizedCardNumber: data.AUserWithACard().Token(),
			},
		},
		{
			name: "Mailing address and residential address empty, replacement return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithSelfServiceResponse(&sspb.GetPartyResponse{
				LegalName: &sspb.Name{
					Name:       "Ms. Oprah Gail Winfrey",
					Prefix:     "Queen",
					FirstName:  "Oprah",
					MiddleName: "Gail",
					LastName:   "Winfrey",
				},
			}),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			wantErr: errors.New("fabric error: status_code=NotFound, error_code=20003, message=replacement failed, reason=address not found"),
		},
		{
			name:    "CTM succeeds damaged replacement return true",
			builder: fixtures.AServer().WithData(data.AUserWithACard(), data.AUserWithACard()),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			want: &cpb.ReplaceResponse{
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
				NewTokenizedCardNumber: data.AUserWithACard().Token(),
			},
		},
		{
			name:    "audit log failure does not affect response",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			want: &cpb.ReplaceResponse{
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
				NewTokenizedCardNumber: data.AUserWithACard().Token(),
			},
		},
		{
			name:    "Eligibility false replacement return true",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEligibilityError(),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_STOLEN,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=replacement failed, reason=card not eligible"),
		},
		{
			name: "replace succeeds, card inquiry call failed return true without Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmInquiryError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_STOLEN,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replacement failed, reason=service unavailable"),
		},
		{
			name:    "unable to call self service",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithSelfServiceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_STOLEN,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replacement failed, reason=service unavailable"),
		},
		{
			name:    "CTM succeeds lost replacement return true",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_LOST,
			},
			want: &cpb.ReplaceResponse{
				NewTokenizedCardNumber: "Some new card number",
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
			},
		},
		{
			name: "succeed when the card is already replaced",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithACard(
						data.WithACardNumber(data.AUserWithACard().CardNumber()),
						data.WithAToken(data.AUserWithACard().Token()),
						data.WithStatus(ctm.StatusStolen),
						data.WithNewCard("4444333322221111", "1234123412341234"),
						data.Active),
					data.WithACard(
						data.WithACardNumber("4444333322221111"),
						data.WithAToken("1234123412341234"),
						data.WithStatus(ctm.StatusIssued),
						data.Active),
				),
			),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_STOLEN,
			},
			want: &cpb.ReplaceResponse{
				NewTokenizedCardNumber: "1234123412341234",
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
			},
		},
		{
			name:    "OCV retrieve party call fails",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithOCVRetrievePartyError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_LOST,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replacement failed, reason=service unavailable"),
		},
		{
			name:    "OCV maintain contract call fails",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithOCVMaintainContractError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_LOST,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replacement failed, reason=service unavailable"),
		},
		{
			name:    "Entitlements RegisterCardToPersona call fails",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntitlementsRegisterCardToPersonaErr(anzerrors.New(codes.Internal, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unable to register card to persona"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_LOST,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=replacement failed, reason=unable to register card to persona"),
		},
		{
			name:    "Entitlements RegisterCardToPersona call fails",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntitlementsForcePartyToLatestErr(anzerrors.New(codes.Internal, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unable to force party to latest"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_LOST,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=replacement failed, reason=unable to force party to latest"),
		},
		{
			name: "Same-day replacement fails before starting",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithOCVRetrievePartyResp(func(parties []*ocv.RetrievePartyRs) {
				for _, p := range parties {
					if a := p.GetAccount(fmt.Sprintf("enc(%s)", data.AUserWithACard().Token())); a != nil {
						a.AccountOpenedDate = time.Now().Format("2006-01-02")
						break
					}
				}
			}),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_LOST,
			},
			wantErr: errors.New("fabric error: status_code=InvalidArgument, error_code=20005, message=replacement failed, reason=cannot replace card number on the same day it was created"),
		},
		{
			name: "Same-day replacement allowed for damaged",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithOCVRetrievePartyResp(func(parties []*ocv.RetrievePartyRs) {
				for _, p := range parties {
					if a := p.GetAccount(fmt.Sprintf("enc(%s)", data.AUserWithACard().Token())); a != nil {
						a.AccountOpenedDate = time.Now().Format("2006-01-02")
						break
					}
				}
			}),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
			},
			want: &cpb.ReplaceResponse{
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
				NewTokenizedCardNumber: data.AUserWithACard().Token(),
			},
		},
		{
			name: "card controls fails",
			builder: fixtures.AServer().WithData(
				data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).
				WithCardControlsTransferControlsError(anzerrors.New(codes.Internal, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from visa gateway"))),
			req: &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              cpb.ReplaceRequest_REASON_LOST,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=replacement failed, reason=unexpected response from visa gateway"),
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, feature.FeatureGate.Set(map[feature.Feature]bool{
				feature.REASON_LOST:    true,
				feature.REASON_STOLEN:  true,
				feature.REASON_DAMAGED: true,
			}))
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardServer(tt.builder)
			got, err := s.Replace(ctx, test.req)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr.Error())
				assert.NotContains(t, b.String(), "no error sent to audit log")
			} else {
				require.NoError(t, err)
				if plasticType(test.req.Reason) == ctm.SameNumber {
					assert.Equal(t, test.req.TokenizedCardNumber, got.NewTokenizedCardNumber)
				} else {
					assert.NotNil(t, got.NewTokenizedCardNumber)
					assert.NotEqual(t, test.req.TokenizedCardNumber, got.NewTokenizedCardNumber)
				}
				equals.AssertJson(t, test.want.GetEligibilities(), got.GetEligibilities())
				assert.NotContains(t, b.String(), "no response data sent to audit log")
			}
		})
	}
}

func TestServer_ReplaceFeatureToggle(t *testing.T) {
	tests := []struct {
		name    string
		reason  cpb.ReplaceRequest_Reason
		want    *cpb.ReplaceResponse
		wantErr error
	}{
		{
			name:    "reason damaged enabled, stolen requested - fails",
			reason:  cpb.ReplaceRequest_REASON_STOLEN,
			wantErr: errors.New("reason not allowed"),
		},
		{
			name:    "reason damaged enabled, lost requested - fails",
			reason:  cpb.ReplaceRequest_REASON_LOST,
			wantErr: errors.New("reason not allowed"),
		},
		{
			name:   "reason damaged enabled, damaged requested - succeeds",
			reason: cpb.ReplaceRequest_REASON_DAMAGED,
			want: &cpb.ReplaceResponse{
				NewTokenizedCardNumber: data.AUserWithACard().Token(),
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
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, feature.FeatureGate.Set(map[feature.Feature]bool{
				feature.REASON_LOST:    false,
				feature.REASON_STOLEN:  false,
				feature.REASON_DAMAGED: true,
			}))
			builder := fixtures.AServer().WithData(data.AUserWithACard())
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardServer(builder)
			request := &cpb.ReplaceRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Reason:              test.reason,
			}
			got, err := s.Replace(ctx, request)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func TestReplaceCardAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		require.NoError(t, feature.FeatureGate.Set(map[feature.Feature]bool{
			feature.REASON_DAMAGED: true,
		}))

		sd := servicedata.ReplaceCard{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}
		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogHook(hook)
		request := &cpb.ReplaceRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			Reason:              cpb.ReplaceRequest_REASON_DAMAGED,
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardServer(builder)

		_, _ = s.Replace(ctx, request)

		require.NoError(t, sd.Validate())

		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, "N", sd.GetNewMediaType())
		assert.Equal(t, "N", sd.GetOldMediaType())
		assert.Equal(t, "20170501", sd.GetOldExpiryDate())
		assert.Equal(t, "20170501", sd.GetNewExpiryDate())
		assert.Equal(t, "20150805", sd.GetOldLastIssueDate())
		assert.Equal(t, "20150805", sd.GetNewLastIssueDate())
		assert.Equal(t, "MR NATHAN FUKUSHIMA MR NATHAN FUKUSHIMA", sd.GetOldNameOnInstrument())
		assert.Equal(t, "MR NATHAN FUKUSHIMA MR NATHAN FUKUSHIMA", sd.GetNewNameOnInstrument())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

func TestToEmbossedName(t *testing.T) {
	tests := []struct {
		scenario       string
		expectedResult string
		FirstName      string
		MiddleName     string
		LastName       string
	}{
		{
			scenario:       "Standard Case",
			expectedResult: "JOE BLOW",
			FirstName:      "Joe",
			LastName:       "Blow",
		},
		{
			scenario:       "Name Greater than 21 return firstname initial + lastName",
			expectedResult: "J BLOW",
			FirstName:      "JoeReallyLongFirstNameGreaterThan21Chars",
			LastName:       "Blow",
		},
		{
			scenario:       "Name Greater than 21 return firstname initial + lastName truncated to 21",
			expectedResult: "J BLOWREALLYLONGLASTN",
			FirstName:      "JoeReallyLongFirstNameGreaterThan21Chars",
			LastName:       "BlowReallyLongLastNameHopeThisWorks",
		},
		{
			scenario:       "Name With 22 chars",
			expectedResult: "T LASTNAME",
			FirstName:      "TestFirstName",
			LastName:       "LastName",
		},
		{
			scenario:       "Name With 21 chars",
			expectedResult: "FIRSTNAMEXXX LASTNAME",
			FirstName:      "FirstNameXXX",
			LastName:       "LastName",
		},
		{
			scenario:       "Name With 20 chars",
			expectedResult: "FIRSTNAMEXX LASTNAME",
			FirstName:      "FirstNameXX",
			LastName:       "LastName",
		},
		{
			scenario:       "Name With multiple spaces",
			expectedResult: "HUMAN 1234",
			FirstName:      "Human  ",
			LastName:       "1234  ",
		},
		{
			scenario:       "When lastname has whitespace in the right with max length",
			expectedResult: "J BLOWREALLYLONGLASTN",

			FirstName: "JoeReallyLongFirstNameGreaterThan21Chars",
			LastName:  "BlowReallyLongLastNameHopeThisWorks  ",
		},
		{
			scenario:       "When lastname has whitespace in the left with max length",
			expectedResult: "J BLOWREALLYLONGLASTN",

			FirstName: "JoeReallyLongFirstNameGreaterThan21Chars",
			LastName: "	BlowReallyLongLastNameHopeThisWorks",
		},
		{
			scenario:       "FirstName Only",
			expectedResult: "JOE",
			FirstName:      "Joe",
		},
		{
			scenario:       "FirstName Only > 21 chars",
			expectedResult: "JOEREALLYLONGFIRSTNAM",
			FirstName:      "JoeReallyLongFirstNameGreaterThan21Chars",
		},
		{
			scenario:       "When firstma,e has whitespace in the left with max length",
			expectedResult: "J BLOWREALLYLONGLASTN",

			FirstName: "	JoeReallyLongFirstNameGreaterThan21Chars",
			LastName: "BlowReallyLongLastNameHopeThisWorks",
		},
		{
			scenario:       "Multiple first names ",
			expectedResult: "JOE XAVIER BLOW",
			FirstName:      "Joe Xavier",
			LastName:       "Blow",
		},
		{
			scenario:       "Multiple Last names ",
			expectedResult: "JOE BLOW SMITH JONES",
			FirstName:      "Joe",
			LastName:       "Blow Smith Jones",
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			result := toEmbossedName(test.FirstName, test.LastName)
			assert.Equal(t, test.expectedResult, result)
			assert.LessOrEqual(t, len(result), MaxEmbossedName)
		})
	}
}

func TestToCtmName(t *testing.T) {
	tests := []struct {
		scenario       string
		expectedResult string
		FirstName      string
		LastName       string
	}{
		{
			scenario:       "Standard Case",
			expectedResult: "JOEBLOW",

			FirstName: "Joe",
			LastName:  "Blow",
		},
		{
			scenario:       "Case greater than 24 chars long last name",
			expectedResult: "JOEBLOWAWAYLASTNAMETHAT",

			FirstName: "Joe",
			LastName:  "BlowAwayLastNameThatIsReallyLong",
		},
		{
			scenario:       "Case greater than 24 chars long firstname",
			expectedResult: "JBLOWAWAY",

			FirstName: "JoeReallyLongFirstNameThatWonFit",
			LastName:  "BlowAway",
		},
		{
			scenario:       "OnlyFirstName ",
			expectedResult: "JOEREALLYLONGLASTNAM",

			FirstName: "JoeReallyLongLastNameThatWonFit",
		},
		{
			scenario:       "Only LastName ",
			expectedResult: "JOEREALLYLONGLASTNAM",

			LastName: "JoeReallyLongLastNameThatWonFit",
		},
		{
			scenario:       "Multiple FirstName ",
			expectedResult: "JOE DANIELBLOW",

			FirstName: "Joe Daniel",
			LastName:  "Blow",
		},
		{
			scenario:       "Multiple LastName ",
			expectedResult: "JOEBLOW SMITH",

			FirstName: "Joe",
			LastName:  "Blow Smith",
		},
		{
			scenario:       "Multiple LastName ",
			expectedResult: "FLAST NAME",

			FirstName: "First Name test 1234",
			LastName:  "Last Name",
		},
		{
			scenario:       "Multiple LastName ",
			expectedResult: "FLAST NAME IS MORE TH",

			FirstName: "First Name1",
			LastName:  "Last Name is more tha",
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			firstName, LastName := toCTMName(test.FirstName, test.LastName)
			ctmName := firstName + LastName
			assert.Equal(t, test.expectedResult, removeSpaces(ctmName))
			assert.LessOrEqual(t, len(ctmName), MaxCtmNameLength)
		})
	}
}

func TestToInitials(t *testing.T) {
	tests := []struct {
		scenario       string
		expectedResult string
		str            string
	}{
		{
			scenario:       "Case No Middle Name",
			expectedResult: "J",
			str:            "Joe Blow",
		},
		{
			scenario:       "Case One Middle Name",
			expectedResult: "J",
			str:            "Joe Xavier Blow",
		},
		{
			scenario:       "Case Multiple Middle Names",
			expectedResult: "J",
			str:            "Joe Xavier Dylan Frank Blow",
		},
		{
			scenario:       "Case Only One Name",
			expectedResult: "J",
			str:            "Joe",
		},
		{
			scenario:       "Case tab",
			expectedResult: "J",
			str: "Joe	",
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			assert.Equal(t, test.expectedResult, toFirstInitial(test.str))
		})
	}
}

func Test_server_publishNotification(t *testing.T) {
	tests := []struct {
		name      string
		personaID string
	}{
		{
			name:      "Happy flow",
			personaID: "1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cc := mock.NewMockPublisher(ctrl)
			notificationToMatch := &sdk.NotificationForPersona{
				PersonaID: tt.personaID,
				Notification: notification.Simple{
					ActionURL: "https://plus.anz/cards",
				},
				Preview: notification.Preview{
					Title: "Card Ordered",
					Body:  "We've cancelled your current card. Your new one should arrive in 5 to 10 days.",
				},
			}
			cc.EXPECT().Publish(gomock.Any(), &matchers.NotificationMatcher{Notification: notificationToMatch}).Times(1).Return(&sdk.PublishResponse{
				Status: "",
			}, nil)
			s := &server{
				Fabric: Fabric{CommandCentre: &commandcentre.Client{Publisher: cc}},
			}
			s.publishNotification(context.Background(), tt.personaID)
		})
	}
}

package cards

import (
	"context"
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/stretchr/testify/require"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/test/util"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/test/fixtures"
	"github.com/stretchr/testify/assert"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"

	"google.golang.org/protobuf/encoding/protojson"
)

func TestActivation(t *testing.T) {
	t.Parallel()
	errFuncCount := 0
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		request *cpb.ActivateRequest
		want    *cpb.ActivateResponse
		wantErr error
	}{
		{
			name:    "rate limit throw error",
			builder: fixtures.AServer().WithRateLimitError(errors.New("over rate limit")),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			wantErr: errors.New("fabric error: status_code=ResourceExhausted, error_code=7, message=activation failed, reason=over rate limit"),
		},
		{
			name:    "vault fails activation return false",
			builder: fixtures.AServer().WithVaultError(errors.New("oh no!")),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=20004, message=activation failed, reason=service unavailable"),
		},
		{
			name:    "last 6 digits do not match activation return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        "123456",
			},
			wantErr: errors.New("fabric error: status_code=InvalidArgument, error_code=20006, message=activation failed, reason=last 6 Digits do not match"),
		},
		{
			name: "Entitlements fails, activation return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=activation failed, reason=service unavailable"),
		},
		{
			name:    "CTM failed with debit card inquiry, return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.Inactive)).WithCtmInquiryError(errors.New("oh no")),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=20000, message=activation failed, reason=service unavailable"),
		},
		{
			name:    "CTM successful with debit card inquiry, card status lost, return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusLost))),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=20002, message=activation failed, reason=card ineligible"),
		},
		{
			name: "entitled false activation return false",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithAPersonaID("personaID"),
					data.WithACard(
						data.WithAToken("token"),
						data.WithACardNumber("1234567890123456")))),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: "token",
				Last_6Digits:        "123456",
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=activation failed, reason=user not entitled"),
		},
		{
			name: "CTM fails activation return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.Inactive)).WithCtmActivateError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=activation failed, reason=service unavailable"),
		},
		{
			name: "CTM succeeds activation, fails to return Eligibility return true with no Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.Inactive)).WithCtmInquiryErrorFunc(func() error {
				if errFuncCount == 0 {
					errFuncCount++
					return nil
				}
				return errors.New("oh no")
			}),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			want: &cpb.ActivateResponse{},
		},
		{
			name:    "CTM succeeds activation return true",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.Inactive)),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			want: &cpb.ActivateResponse{
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
			name:    "auditlogger error has no affect on response",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.Inactive)).WithAuditLogError(errors.New("oh no")),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			want: &cpb.ActivateResponse{
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
			name:    "ineligible for activation",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.Active)),
			request: &cpb.ActivateRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
			},
			want: &cpb.ActivateResponse{
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardServer(test.builder)

			got, err := s.Activate(ctx, test.request)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func TestActivationAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.ActivateCard{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}
		builder := fixtures.AServer().WithData(data.AUserWithACard(data.Inactive)).WithAuditLogHook(hook)
		request := &cpb.ActivateRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			Last_6Digits:        data.AUserWithACard().CardNumber()[10:],
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardServer(builder)

		_, _ = s.Activate(ctx, request)
		require.NoError(t, sd.Validate())
		assert.Equal(t, "CNE", sd.GetCustomerClass())
		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, "New", sd.GetIssueReason())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

func Test_last6DigitsMatch(t *testing.T) {
	t.Run("Successfully match last 6 digits", func(t *testing.T) {
		assert.True(t, last6DigitsMatch("1234567890123456", "123456"))
	})
	t.Run("last 6 digits do not match", func(t *testing.T) {
		assert.False(t, last6DigitsMatch("1234567890123456", "654321"))
	})
	t.Run("card number provided is an empty string", func(t *testing.T) {
		assert.False(t, last6DigitsMatch("", "654321"))
	})
	t.Run("token provided as card number", func(t *testing.T) {
		assert.False(t, last6DigitsMatch("6688390512341000", "654321"))
	})
	t.Run("16 alpha characters provided as card number", func(t *testing.T) {
		assert.False(t, last6DigitsMatch("aaaaaaaaaaaaaaaa", "654321"))
	})
	t.Run("short alpha string provided as card number", func(t *testing.T) {
		assert.False(t, last6DigitsMatch("a", "654321"))
	})
	t.Run("short numerical string provided as card number", func(t *testing.T) {
		assert.False(t, last6DigitsMatch("123456", "654321"))
	})
}

package cards

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/util"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/test/fixtures"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	servicedata "github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"

	"github.com/stretchr/testify/assert"
)

func TestServer_ResetPIN(t *testing.T) {
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		req       *cpb.ResetPINRequest
		want      *cpb.ResetPINResponse
		wantErr   error
	}{
		{
			name: "Entitlements fails, ResetPIN return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=reset PIN failed, reason=service unavailable"),
		},
		{
			name:      "User does not have operation manage card",
			personaID: data.RandomUser().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=reset PIN failed, reason=user not entitled"),
		},
		{
			name:      "Vault fail ResetPIN return error",
			personaID: data.AUserWithACard().PersonaID,
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithVaultError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=reset PIN failed, reason=service unavailable"),
		},
		{
			name:      "EchidnaClient Call fails, return error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1010),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=reset PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna 1011 Select PIN operation failed due to request-response error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1011),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=reset PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna 1012 RemotePIN service unavailable",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1012),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=reset PIN failed, reason=Service unavailable"),
		},
		{
			name:      "Echidna 1013 Select PIN operation has timed out",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1013),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=DeadlineExceeded, error_code=2, message=reset PIN failed, reason=Operation has timed out"),
		},
		{
			name:      "Echidna 1015 Select PIN operation failed due to RemotePIN service error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1015),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=reset PIN failed, reason=Operation failed due to service error"),
		},
		{
			name:      "EchidnaClient Call succeeds, return response",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))),
			req: &cpb.ResetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			want: &cpb.ResetPINResponse{
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_SET_PIN,
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
			ctx, b := fixtures.GetTestContextWithLogger(&test.personaID)
			s := buildCardServer(tt.builder)
			got, err := s.ResetPIN(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func TestResetPINAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.ChangePin{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}

		builder := fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithAuditLogHook(hook)
		request := &cpb.ResetPINRequest{
			TokenizedCardNumber: data.AUserWithACard(data.Inactive).Token(),
			EncryptedPinBlock:   "hahahaha",
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardServer(builder)

		_, _ = s.ResetPIN(ctx, request)
		require.NoError(t, sd.Validate())
		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

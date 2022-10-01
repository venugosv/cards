package cards

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/test/data"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	encryptedVerifyPINBlock = "Kjqob9DsvF2BTBTC3l5kB4aQCCk07kRPDroSJ8EuUfEY1Q9QxNJLu/ySzZN6/g4QoBeLXhrfTDSo0WwzY2mV1n82d4HOcG827Zw9r1+/wONQ8kU81jWF142qwqLskwapk3vH8Ol3SqRRm87gE4cBSv/ffD9q6rZYkNvIqtg82bPll/cFF85ave0zYYNNCwwf33hwkf4GlEzEzBwf2XqFuwSveXJ6Owmex4AaZaWRi2C45njcnn7F7X+EsWwzbiOnhExHJhJUW9J0z7gWuxmjduMLBy420qNPYNKytPtHNXGnxpba9TpI+I48A50hmDVRLAJUBgjLbIZfegVZztp3pQ=="
)

func TestServer_VerifyPIN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		req       *cpb.VerifyPINRequest
		want      *cpb.VerifyPINResponse
		wantErr   error
	}{
		{
			name:      "Entitlements fails replacement return false",
			personaID: data.AUserWithACard().PersonaID,
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=verify PIN failed, reason=service unavailable"),
		},
		{
			name:      "User does not have operation manage card",
			personaID: data.RandomUser().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=verify PIN failed, reason=user not entitled"),
		},
		{
			name:      "Eligibility false replacement return true",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))).WithEligibilityError(),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=verify PIN failed, reason=card not eligible"),
		},
		{
			name:      "Vault fail replacement return true",
			personaID: data.AUserWithACard().PersonaID,
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=verify PIN failed, reason=service unavailable"),
		},
		{
			name:      "EchidnaClient Call fails, return false",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1010),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=verify PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna code 1011: Verify PIN operation failed due to request-response error., return false",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1011),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=verify PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna code 1012: RemotePIN service unavailable., return false",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1012),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=verify PIN failed, reason=Service unavailable"),
		},
		{
			name:      "Echidna code 1013: Verify PIN operation has timed out, return false",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1013),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=DeadlineExceeded, error_code=2, message=verify PIN failed, reason=Operation has timed out"),
		},
		{
			name:      "Echidna code 1015: Verify PIN operation failed due to RemotePIN service error. with error 55 incorrect pin, return false",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1015),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=verify PIN failed, reason=Operation failed due to service error"),
		},
		{
			name:      "EchidnaClient Call succeeds return true",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()),
			req: &cpb.VerifyPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedVerifyPINBlock,
			},
			want: &cpb.VerifyPINResponse{},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardServer(tt.builder)
			got, err := s.VerifyPIN(fixtures.GetTestContextWithJWT(test.personaID), test.req)
			if test.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

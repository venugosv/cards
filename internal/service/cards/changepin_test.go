package cards

import (
	"context"
	"testing"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/test/util"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/test/fixtures"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	"github.com/pkg/errors"
)

const (
	encryptedOldPINBlock = "h6VXDX/10a4s4EF2h947f55/xCPiADzEcOju1ZnPIrtjEIwejhEKlIfURWLE7vajffhZ/YNFdQV/fG9aQRALJQpos8K4w772PyNYumlML1vA6wtZninHJwaYJOGueX+N2gd63lsyhnwt5dezTNBhPDWcccXeRc3BK4nNWNkbm40Ng+zHCxSjn/Oqcav00kAIXBtGxfijU/8s7sAf8Ss9TwoVq1FG35ceFiWqfTotH4wD8X7BwH2N7WN3KYJMOeA9myVffdqnv9QNk/ldq/Nvou3AaH2qUilqORyzErBugSfoAeCe9u2/BAzKZeh1q7USKhO5SzfNGb+WwY8Ui1+juQ=="
	encryptedNewPINBlock = "depZGnbxrHYgnFJYv78HAOMdetVjQxf6ZRLCfUkO5D7Fn8JGCtXvaZbYwTATTM6iFrhW9gk4Z+7Ow3aPqW+SwghKLpZaLdQHs8kwGGHYNysP43Gblgt3EagrigNcytinHTVBixj4cDVKy3UzlKDuoM42Fp81Qrw/YNuI2+HAokHuC1UP2bL5+nGmQ04dsjn0QEa4cTmhsaeaZxQNvNiFKibztjIfHVWvvKsztDc2rflme6KuVVvtXaS3Psls2GGg6fc3CfV+2rXQzvk/KnyVmgqvr0+NKONJf/PqrILyqwqEhREfC8bmtzz9x0GPffoozEdQl7mG6gTgRrYF7qXhqQ=="
)

func TestServer_ChangePIN(t *testing.T) {
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		req       *cpb.ChangePINRequest
		want      *cpb.ChangePINResponse
		wantErr   error
	}{
		{
			name: "Entitlements fails changepin return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=change PIN failed, reason=service unavailable"),
		},
		{
			name: "Entitlements false changepin return error",
			builder: fixtures.AServer().WithData(
				data.AUser(
					data.WithAPersonaID("personaID"),
					data.WithACard(
						data.WithAToken("token"),
						data.WithACardNumber("1234567890123456")))),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.RandomUser().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=change PIN failed, reason=user not entitled"),
		},
		{
			name:      "User does not have operation manage error",
			personaID: data.RandomUser().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().CardNumber(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=change PIN failed, reason=user not entitled"),
		},
		{
			name:      "Eligibility false replacement return error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEligibilityError(),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=change PIN failed, reason=card not eligible"),
		},
		{
			name:      "Vault fail replacement return error",
			personaID: data.AUserWithACard().PersonaID,
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=change PIN failed, reason=service unavailable"),
		},
		{
			name:      "Echidna 55 Incorrect PIN",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(55),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=InvalidArgument, error_code=4, message=change PIN failed, reason=Incorrect PIN"),
		},
		{
			name:      "Echidna 75 Maximum PIN tries exceeded",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(75),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=ResourceExhausted, error_code=7, message=change PIN failed, reason=Maximum PIN tries exceeded"),
		},
		{
			name:      "Echidna 1010 Change PIN operation failed due to Tandem error response",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1010),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=change PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna 1011 Change PIN operation failed due to request-response error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1011),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=change PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna 1012 RemotePIN service unavailable",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1012),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=change PIN failed, reason=Service unavailable"),
		},
		{
			name:      "Echidna 1013 Change PIN operation has timed out",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1013),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=DeadlineExceeded, error_code=2, message=change PIN failed, reason=Operation has timed out"),
		},
		{
			name:      "Echidna 1015 Change PIN operation failed due to RemotePIN service error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1015),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=change PIN failed, reason=Operation failed due to service error"),
		},
		{
			name:      "EchidnaClient Call succeeds, no error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			want: &cpb.ChangePINResponse{},
		},
		{
			name:      "audit log failure does not affect response",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogError(errors.New("oh no")),
			req: &cpb.ChangePINRequest{
				TokenizedCardNumber:  data.AUserWithACard().Token(),
				EncryptedPinBlockNew: encryptedOldPINBlock,
				EncryptedPinBlockOld: encryptedNewPINBlock,
			},
			want: &cpb.ChangePINResponse{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, b := fixtures.GetTestContextWithLogger(&test.personaID)
			s := buildCardServer(test.builder)
			got, err := s.ChangePIN(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

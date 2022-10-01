package cards

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/test/util"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"

	"github.com/pkg/errors"
)

const (
	encryptedPINBlock = "Kjqob9DsvF2BTBTC3l5kB4aQCCk07kRPDroSJ8EuUfEY1Q9QxNJLu/ySzZN6/g4QoBeLXhrfTDSo0WwzY2mV1n82d4HOcG827Zw9r1+/wONQ8kU81jWF142qwqLskwapk3vH8Ol3SqRRm87gE4cBSv/ffD9q6rZYkNvIqtg82bPll/cFF85ave0zYYNNCwwf33hwkf4GlEzEzBwf2XqFuwSveXJ6Owmex4AaZaWRi2C45njcnn7F7X+EsWwzbiOnhExHJhJUW9J0z7gWuxmjduMLBy420qNPYNKytPtHNXGnxpba9TpI+I48A50hmDVRLAJUBgjLbIZfegVZztp3pQ=="
)

func TestServer_SetPIN(t *testing.T) {
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		req       *cpb.SetPINRequest
		want      *cpb.SetPINResponse
		wantErr   error
	}{
		{
			name:      "Entitlements fails SetPIN return error",
			personaID: data.AUserWithACard().PersonaID,
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEntMayError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set PIN failed, reason=service unavailable"),
		},
		{
			name:      "User does not have operation manage card",
			personaID: data.RandomUser().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=set PIN failed, reason=user not entitled"),
		},
		{
			name:      "Eligibility false SetPIN return error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen), data.WithPINChangeCount(0))).WithEligibilityError(),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=set PIN failed, reason=card not eligible"),
		},
		{
			name:      "Vault fail SetPIN return error",
			personaID: data.AUserWithACard().PersonaID,
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithVaultError(
				anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set PIN failed, reason=service unavailable"),
		},
		{
			name:      "Echidna 1010 Select PIN operation failed due to Tandem error response",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1010),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=set PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna 1011 Select PIN operation failed due to request-response error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1011),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Internal, error_code=2, message=set PIN failed, reason=Operation failed due to internal error"),
		},
		{
			name:      "Echidna 1012 RemotePIN service unavailable",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1012),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set PIN failed, reason=Service unavailable"),
		},
		{
			name:      "Echidna 1013 Select PIN operation has timed out",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1013),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=DeadlineExceeded, error_code=2, message=set PIN failed, reason=Operation has timed out"),
		},
		{
			name:      "Echidna 1015 Select PIN operation failed due to RemotePIN service error",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithEchidnaErrorCode(1015),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=set PIN failed, reason=Operation failed due to service error"),
		},
		{
			name:      "Echidna succeeds, return response",
			personaID: data.AUserWithACard().PersonaID,
			builder:   fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))),
			req: &cpb.SetPINRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				EncryptedPinBlock:   encryptedPINBlock,
			},
			want: &cpb.SetPINResponse{
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
			got, err := s.SetPIN(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func TestSetPINAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.SetPin{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}

		builder := fixtures.AServer().WithData(data.AUserWithACard(data.WithPINChangeCount(0))).WithAuditLogHook(hook)
		request := &cpb.SetPINRequest{
			TokenizedCardNumber: data.AUserWithACard(data.Inactive).Token(),
			EncryptedPinBlock:   "hahahaha",
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardServer(builder)

		_, _ = s.SetPIN(ctx, request)
		require.NoError(t, sd.Validate())
		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

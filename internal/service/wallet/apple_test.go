package wallet

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/fixtures"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestServer_CreateApplePaymentToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		req       *cpb.CreateApplePaymentTokenRequest
		wantErr   string
	}{
		{
			name:    "happy path",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &cpb.CreateApplePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Nonce:               nonce,
				NonceSignature:      nonceSignature,
				Certificates:        []string{certificate},
			},
		},
		{
			name: "entitlements failed",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithEntMayError(anzerrors.New(codes.Internal, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from downstream"))),
			req: &cpb.CreateApplePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Nonce:               nonce,
				NonceSignature:      nonceSignature,
				Certificates:        []string{certificate},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name:    "eligibility failed",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEligibilityError(),
			req: &cpb.CreateApplePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Nonce:               nonce,
				NonceSignature:      nonceSignature,
				Certificates:        []string{certificate},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name: "ctm inquiry failed",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithCtmInquiryError(anzerrors.New(codes.Internal, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from downstream"))),
			req: &cpb.CreateApplePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Nonce:               nonce,
				NonceSignature:      nonceSignature,
				Certificates:        []string{certificate},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name: "vault failed decode",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithVaultError(anzerrors.New(codes.Internal, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from downstream"))),
			req: &cpb.CreateApplePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Nonce:               nonce,
				NonceSignature:      nonceSignature,
				Certificates:        []string{certificate},
			},
			wantErr: "fabric error: status_code=Internal, error_code=20004, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name: "apcam failed provisioning",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithPushProvisionError(anzerrors.New(codes.Internal, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from downstream"))),
			req: &cpb.CreateApplePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Nonce:               nonce,
				NonceSignature:      nonceSignature,
				Certificates:        []string{certificate},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardServer(test.builder)
			got, err := s.CreateApplePaymentToken(fixtures.GetTestContext(), test.req)
			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.Error(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got.GetActivationData())
				assert.NotNil(t, got.GetEncryptedPassData())
				assert.NotNil(t, got.GetEphemeralPublicKey())
			}
		})
	}
}

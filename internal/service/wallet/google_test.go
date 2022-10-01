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

const stableHardwareId = "stableHardwareId"

func TestServer_CreateGooglePaymentToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		req       *cpb.CreateGooglePaymentTokenRequest
		wantErr   string
	}{
		{
			name:    "happy path",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &cpb.CreateGooglePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				StableHardwareId:    stableHardwareId,
			},
		},
		{
			name: "entitlements failed",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithEntMayError(anzerrors.New(codes.Internal, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from downstream"))),
			req: &cpb.CreateGooglePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				StableHardwareId:    stableHardwareId,
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name:    "eligibility failed",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEligibilityError(),
			req: &cpb.CreateGooglePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				StableHardwareId:    stableHardwareId,
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name: "vault failed decode",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithVaultError(anzerrors.New(codes.Internal, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unexpected response from downstream"))),
			req: &cpb.CreateGooglePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				StableHardwareId:    stableHardwareId,
			},
			wantErr: "fabric error: status_code=Internal, error_code=20004, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name: "gpay failed provisioning",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithGPayError(anzerrors.New(codes.Internal, "failed to Create GPay JWE",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.ValidationFailure, "unable to create payload encryptor"))),
			req: &cpb.CreateGooglePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				StableHardwareId:    stableHardwareId,
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
		{
			name: "self service failed provisioning",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).
				WithSelfServiceError(anzerrors.New(codes.Unavailable, "failed to create SelfService adapter",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "unable to make successful connection"))),
			req: &cpb.CreateGooglePaymentTokenRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				StableHardwareId:    stableHardwareId,
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=Push Provisioning Failed, reason=service unavailable",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardServer(test.builder)
			got, err := s.CreateGooglePaymentToken(fixtures.GetTestContext(), test.req)
			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.Error(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got.OpaquePaymentCard)
				assert.Equal(t, cpb.TokenProvider_TOKEN_PROVIDER_VISA, got.TokenProvider)
				assert.Equal(t, cpb.CardNetwork_CARD_NETWORK_VISA, got.CardNetwork)
				assert.NotNil(t, got.UserAddress)
			}
		})
	}
}

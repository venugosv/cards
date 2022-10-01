package wallet

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	"github.com/anzx/pkg/auditlog"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"github.com/anzx/fabric-cards/pkg/date"
	"github.com/anzx/fabric-cards/pkg/integration/apcam"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"
)

func (s server) CreateApplePaymentToken(ctx context.Context, req *cpb.CreateApplePaymentTokenRequest) (retResponse *cpb.CreateApplePaymentTokenResponse, retError error) {
	tokenizedCardNumber := req.GetTokenizedCardNumber()
	serviceData := &servicedata.CreatePaymentToken{
		TokenizedCardNumber: req.TokenizedCardNumber,
		Provider:            apple,
	}

	defer func() {
		if err := serviceData.Validate(); err != nil {
			logf.Error(ctx, err, "invalid service data payload")
		}
		s.auditLog.Publish(ctx, auditlog.EventCreatePaymentToken, retResponse, retError, serviceData)
	}()

	entitledCard, err := s.entitlements.GetEntitledCard(ctx, tokenizedCardNumber, entitlements.OPERATION_VIEW_CARD)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.PermissionDenied, pushProvisioningFailed, anzerrors.GetErrorInfo(err))
	}

	serviceData.AccountNumbers = entitledCard.GetAccountNumbers()

	if err := s.eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_APPLE_PAY, tokenizedCardNumber); err != nil {
		return nil, anzerrors.Wrap(err, codes.PermissionDenied, pushProvisioningFailed, anzerrors.GetErrorInfo(err))
	}

	card, err := s.ctm.DebitCardInquiry(ctx, tokenizedCardNumber)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.NotFound, pushProvisioningFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, serviceUnavailable))
	}

	serviceData.Last_4Digits = card.CardNumber.Last4Digits
	expiry := date.GetDate(ctx, date.YYMM, card.ExpiryDate)

	cardNumber, err := s.vault.DecodeCardNumber(ctx, tokenizedCardNumber)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, pushProvisioningFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardTokenizationFailed, serviceUnavailable))
	}

	pushRequest := apcam.Request{
		CardInfo: apcam.CardInfo{
			Fpan:       cardNumber,
			ExpiryDate: date.ProtoToDate(expiry).String(),
		},
		Apple: apcam.Apple{
			Nonce:          req.GetNonce(),
			NonceSignature: req.GetNonceSignature(),
			Certificates:   req.GetCertificates(),
		},
	}

	pushResponse, err := s.apcam.PushProvision(ctx, pushRequest)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, pushProvisioningFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, serviceUnavailable))
	}

	push := pushResponse.Apple

	return &cpb.CreateApplePaymentTokenResponse{
		ActivationData:     push.ActivationData,
		EncryptedPassData:  push.EncryptedPassData,
		EphemeralPublicKey: push.EphemeralPublicKey,
	}, nil
}

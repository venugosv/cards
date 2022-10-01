package wallet

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/gpay"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	"github.com/anzx/pkg/auditlog"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

func (s server) CreateGooglePaymentToken(ctx context.Context, req *cpb.CreateGooglePaymentTokenRequest) (retResponse *cpb.CreateGooglePaymentTokenResponse, retError error) {
	tokenizedCardNumber := req.GetTokenizedCardNumber()
	serviceData := &servicedata.CreatePaymentToken{
		TokenizedCardNumber: req.TokenizedCardNumber,
		Provider:            google,
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

	if err := s.eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_GOOGLE_PAY, tokenizedCardNumber); err != nil {
		return nil, anzerrors.Wrap(err, codes.PermissionDenied, pushProvisioningFailed, anzerrors.GetErrorInfo(err))
	}

	card, err := s.ctm.DebitCardInquiry(ctx, tokenizedCardNumber)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.NotFound, pushProvisioningFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, serviceUnavailable))
	}

	cardNumber, err := s.vault.DecodeCardNumber(ctx, tokenizedCardNumber)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, pushProvisioningFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardTokenizationFailed, serviceUnavailable))
	}

	party, err := s.selfService.GetParty(ctx)
	if err != nil {
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), pushProvisioningFailed, anzerrors.GetErrorInfo(err))
	}

	address, err := party.GetAddress(ctx)
	if err != nil {
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), pushProvisioningFailed, anzerrors.GetErrorInfo(err))
	}

	name, err := party.GetName(ctx)
	if err != nil {
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), pushProvisioningFailed, anzerrors.GetErrorInfo(err))
	}

	payload, err := gpay.NewPayload(ctx, card, cardNumber, address, req.GetStableHardwareId(), req.GetActiveWalletId())
	if err != nil {
		return nil, anzerrors.New(codes.InvalidArgument, pushProvisioningFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, err.Error()))
	}

	opc, err := s.gPay.CreateJWE(ctx, string(payload))
	if err != nil {
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), pushProvisioningFailed, anzerrors.GetErrorInfo(err))
	}

	return &cpb.CreateGooglePaymentTokenResponse{
		OpaquePaymentCard: opc,
		TokenProvider:     cpb.TokenProvider_TOKEN_PROVIDER_VISA,
		CardNetwork:       cpb.CardNetwork_CARD_NETWORK_VISA,
		UserAddress: &cpb.Address{
			LineOne:            address.GetLineOne(),
			LineTwo:            address.GetLineTwo(),
			CountryCode:        address.GetCountry(),
			Locality:           address.GetCity(),
			AdministrativeArea: address.GetState(),
			Name:               name,
			PhoneNumber:        party.GetSecurityMobile(),
			PostalCode:         address.GetPostalCode(),
		},
	}, nil
}

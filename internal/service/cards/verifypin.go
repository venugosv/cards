package cards

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/integration/echidna"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
)

const verifyPINFailed = "verify PIN failed"

func (s server) VerifyPIN(ctx context.Context, req *cpb.VerifyPINRequest) (*cpb.VerifyPINResponse, error) {
	if _, err := s.Entitlements.GetEntitledCard(ctx, req.TokenizedCardNumber, entitlements.OPERATION_MANAGE_CARD); err != nil {
		return nil, serviceErr(err, verifyPINFailed)
	}

	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_CHANGE_PIN, req.TokenizedCardNumber); err != nil {
		return nil, serviceErr(err, verifyPINFailed)
	}

	cardNumber, err := s.Vault.DecodeCardNumber(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, verifyPINFailed)
	}

	request := echidna.IncomingRequest{
		PlainPAN:          cardNumber,
		EncryptedPINBlock: req.EncryptedPinBlock,
	}

	if err := s.Echidna.VerifyPIN(ctx, request); err != nil {
		return nil, serviceErr(err, verifyPINFailed)
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardStatusChange)

	return &cpb.VerifyPINResponse{}, nil
}

package cards

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/pkg/integration/echidna"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/pkg/auditlog"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

const changePINFailed = "change PIN failed"

func (s server) ChangePIN(ctx context.Context, req *cpb.ChangePINRequest) (retResponse *cpb.ChangePINResponse, retError error) {
	defer func() {
		s.AuditLog.Publish(ctx, auditlog.EventChangePin, retResponse, retError, nil)
	}()

	if _, err := s.Entitlements.GetEntitledCard(ctx, req.TokenizedCardNumber, entitlements.OPERATION_MANAGE_CARD); err != nil {
		return nil, serviceErr(err, changePINFailed)
	}

	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_CHANGE_PIN, req.TokenizedCardNumber); err != nil {
		return nil, serviceErr(err, changePINFailed)
	}

	cardNumber, err := s.Vault.DecodeCardNumber(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, changePINFailed)
	}

	request := echidna.IncomingChangePINRequest{
		PlainPAN:             cardNumber,
		EncryptedPINBlockNew: req.EncryptedPinBlockNew,
		EncryptedPINBlockOld: req.EncryptedPinBlockOld,
	}

	if err := s.Echidna.ChangePIN(ctx, request); err != nil {
		return nil, serviceErr(err, changePINFailed)
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardStatusChange)

	return &cpb.ChangePINResponse{}, nil
}

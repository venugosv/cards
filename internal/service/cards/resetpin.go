package cards

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/integration/echidna"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	servicedata "github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	"github.com/anzx/pkg/auditlog"
)

const resetPINFailed = "reset PIN failed"

func (s server) ResetPIN(ctx context.Context, req *cpb.ResetPINRequest) (retResponse *cpb.ResetPINResponse, retError error) {
	serviceData := &servicedata.ChangePin{
		TokenizedCardNumber: req.TokenizedCardNumber,
	}

	defer func() {
		if err := serviceData.Validate(); err != nil {
			logf.Error(ctx, err, "invalid service data payload")
		}
		s.AuditLog.Publish(ctx, auditlog.EventChangePin, retResponse, retError, serviceData)
	}()

	entitledCard, err := s.Entitlements.GetEntitledCard(ctx, req.TokenizedCardNumber, entitlements.OPERATION_MANAGE_CARD)
	if err != nil {
		return nil, serviceErr(err, resetPINFailed)
	}
	serviceData.AccountNumbers = entitledCard.GetAccountNumbers()

	cardNumber, err := s.Vault.DecodeCardNumber(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, resetPINFailed)
	}

	serviceData.Last_4Digits = cardNumber[12:]

	request := echidna.IncomingRequest{
		PlainPAN:          cardNumber,
		EncryptedPINBlock: req.EncryptedPinBlock,
	}

	if err := s.Echidna.SelectPIN(ctx, request); err != nil {
		return nil, serviceErr(err, resetPINFailed)
	}

	if feature.FeatureGate.Enabled(feature.PIN_CHANGE_COUNT) {
		if ok, err := s.CTM.UpdatePINInfo(ctx, req.TokenizedCardNumber); !ok {
			logf.Error(ctx, err, "failed to update PIN info")
		}
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardStatusChange)

	card, err := s.CTM.DebitCardInquiry(ctx, req.TokenizedCardNumber)
	if err != nil {
		return &cpb.ResetPINResponse{}, nil
	}

	return &cpb.ResetPINResponse{Eligibilities: card.Eligibility()}, nil
}

package v1beta1

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"
	"github.com/anzx/pkg/auditlog"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

func (s server) Block(ctx context.Context, req *ccpb.BlockRequest) (retResponse *ccpb.BlockResponse, retError error) {
	serviceData := &servicedata.BlockCard{
		TokenizedCardNumber: req.TokenizedCardNumber,
	}
	defer func() {
		if err := serviceData.Validate(); err != nil {
			logf.Error(ctx, err, "invalid service data payload")
		}
		s.AuditLog.Publish(ctx, actionAuditLog(req.Action), retResponse, retError, serviceData)
	}()

	failMessage := getFailMessage(req.Action)

	entitledCard, err := s.Entitlements.GetEntitledCard(ctx, req.TokenizedCardNumber, entitlements.OPERATION_CARDCONTROLS)
	if err != nil {
		return nil, serviceErr(err, failMessage)
	}
	serviceData.AccountNumbers = entitledCard.GetAccountNumbers()

	if err := s.Eligibility.Can(ctx, actionEligibility(req.Action), req.TokenizedCardNumber); err != nil {
		return nil, serviceErr(err, failMessage)
	}

	response, err := s.CTM.UpdateStatus(ctx, req.TokenizedCardNumber, actionCardStatus(req.Action))
	if err != nil {
		return nil, serviceErr(err, failMessage)
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardStatusChange)

	card, err := s.CTM.DebitCardInquiry(ctx, req.TokenizedCardNumber)
	if err != nil {
		return &ccpb.BlockResponse{
			Status: true,
		}, nil
	}

	serviceData.Last_4Digits = card.CardNumber.Last4Digits

	return &ccpb.BlockResponse{
		Status:        response,
		Eligibilities: card.Eligibility(),
	}, nil
}

func getFailMessage(action ccpb.BlockRequest_Action) string {
	failMessage := "block failed"
	if action == ccpb.BlockRequest_ACTION_UNBLOCK {
		failMessage = "unblock failed"
	}
	return failMessage
}

func actionEligibility(in ccpb.BlockRequest_Action) epb.Eligibility {
	switch in {
	case ccpb.BlockRequest_ACTION_UNKNOWN_UNSPECIFIED:
		return epb.Eligibility_ELIGIBILITY_INVALID_UNSPECIFIED
	case ccpb.BlockRequest_ACTION_BLOCK:
		return epb.Eligibility_ELIGIBILITY_BLOCK
	case ccpb.BlockRequest_ACTION_UNBLOCK:
		return epb.Eligibility_ELIGIBILITY_UNBLOCK
	default:
		return epb.Eligibility_ELIGIBILITY_INVALID_UNSPECIFIED
	}
}

func actionCardStatus(in ccpb.BlockRequest_Action) ctm.Status {
	switch in {
	case ccpb.BlockRequest_ACTION_BLOCK:
		return ctm.StatusTemporaryBlock
	case ccpb.BlockRequest_ACTION_UNBLOCK:
		return ctm.StatusIssued
	default:
		return ""
	}
}

func actionAuditLog(in ccpb.BlockRequest_Action) auditlog.Event {
	switch in {
	case ccpb.BlockRequest_ACTION_BLOCK:
		return auditlog.EventBlockCard
	case ccpb.BlockRequest_ACTION_UNBLOCK:
		return auditlog.EventUnblockCard
	default:
		return ""
	}
}

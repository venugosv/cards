package v1beta2

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/auditlog"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

func (s server) BlockCard(ctx context.Context, req *ccpb.BlockCardRequest) (retResponse *ccpb.BlockCardResponse, retError error) {
	failMessage := getFailMessage(req.GetAction())

	entitledCard, err := s.Entitlements.GetEntitledCard(ctx, req.GetTokenizedCardNumber(), entitlements.OPERATION_CARDCONTROLS)
	if err != nil {
		return nil, serviceErr(err, failMessage)
	}

	card, err := s.CTM.DebitCardInquiry(ctx, req.GetTokenizedCardNumber())
	if err != nil {
		return nil, serviceErr(err, failMessage)
	}

	if hasTempBlock(req.GetAction(), card.Status) {
		if err := s.Eligibility.Can(ctx, actionEligibility(req.Action), req.GetTokenizedCardNumber()); err != nil {
			return nil, serviceErr(err, failMessage)
		}

		defer func() {
			serviceData := &servicedata.BlockCard{
				TokenizedCardNumber: req.GetTokenizedCardNumber(),
				Last_4Digits:        card.CardNumber.Last4Digits,
				AccountNumbers:      entitledCard.GetAccountNumbers(),
			}
			if err := serviceData.Validate(); err != nil {
				logf.Error(ctx, err, "invalid service data payload")
			}
			s.AuditLog.Publish(ctx, actionAuditLog(req.GetAction()), retResponse, retError, serviceData)
		}()

		if _, err := s.CTM.UpdateStatus(ctx, req.GetTokenizedCardNumber(), actionCardStatus(req.GetAction())); err != nil {
			return nil, serviceErr(err, failMessage)
		}

		s.CommandCentre.PublishEventAsync(ctx, event.CardStatusChange)

		card, err = s.CTM.DebitCardInquiry(ctx, req.GetTokenizedCardNumber())
		if err != nil {
			return &ccpb.BlockCardResponse{}, nil
		}

		return &ccpb.BlockCardResponse{
			Eligibilities: card.Eligibility(),
		}, nil
	}

	switch req.GetAction() {
	case ccpb.BlockCardRequest_ACTION_BLOCK:
		request := &ccpb.SetControlsRequest{
			TokenizedCardNumber: req.GetTokenizedCardNumber(),
			CardControls: []*ccpb.ControlRequest{
				{
					ControlType: ccpb.ControlType_GCT_GLOBAL,
				},
			},
		}
		if _, err := s.SetControls(ctx, request); err != nil {
			return nil, serviceErr(err, failMessage)
		}
	case ccpb.BlockCardRequest_ACTION_UNBLOCK:
		request := &ccpb.RemoveControlsRequest{
			TokenizedCardNumber: req.GetTokenizedCardNumber(),
			ControlTypes: []ccpb.ControlType{
				ccpb.ControlType_GCT_GLOBAL,
			},
		}
		if _, err := s.RemoveControls(ctx, request); err != nil {
			return nil, serviceErr(err, failMessage)
		}
	}

	return &ccpb.BlockCardResponse{
		Eligibilities: card.Eligibility(),
	}, nil
}

func hasTempBlock(action ccpb.BlockCardRequest_Action, cardStatus ctm.Status) bool {
	return action == ccpb.BlockCardRequest_ACTION_UNBLOCK && cardStatus == ctm.StatusTemporaryBlock
}

func actionAuditLog(in ccpb.BlockCardRequest_Action) auditlog.Event {
	switch in {
	case ccpb.BlockCardRequest_ACTION_BLOCK:
		return auditlog.EventBlockCard
	case ccpb.BlockCardRequest_ACTION_UNBLOCK:
		return auditlog.EventUnblockCard
	default:
		return ""
	}
}

func actionCardStatus(in ccpb.BlockCardRequest_Action) ctm.Status {
	switch in {
	case ccpb.BlockCardRequest_ACTION_BLOCK:
		return ctm.StatusTemporaryBlock
	case ccpb.BlockCardRequest_ACTION_UNBLOCK:
		return ctm.StatusIssued
	default:
		return ""
	}
}

func actionEligibility(in ccpb.BlockCardRequest_Action) epb.Eligibility {
	switch in {
	case ccpb.BlockCardRequest_ACTION_UNKNOWN_UNSPECIFIED:
		return epb.Eligibility_ELIGIBILITY_INVALID_UNSPECIFIED
	case ccpb.BlockCardRequest_ACTION_BLOCK:
		return epb.Eligibility_ELIGIBILITY_BLOCK
	case ccpb.BlockCardRequest_ACTION_UNBLOCK:
		return epb.Eligibility_ELIGIBILITY_UNBLOCK
	default:
		return epb.Eligibility_ELIGIBILITY_INVALID_UNSPECIFIED
	}
}

func getFailMessage(action ccpb.BlockCardRequest_Action) string {
	switch action {
	case ccpb.BlockCardRequest_ACTION_BLOCK:
		return "block failed"
	case ccpb.BlockCardRequest_ACTION_UNBLOCK:
		return "unblock failed"
	default:
		return "block/unblock failed"
	}
}

package v1beta1

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/pkg/auditlog"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

func (s server) Remove(ctx context.Context, req *ccpb.RemoveRequest) (retResponse *ccpb.CardControlResponse, retError error) {
	serviceData := initServiceData(req)

	defer func() {
		if err := serviceData.Validate(); err != nil {
			logf.Error(ctx, err, "invalid service data payload")
		}
		s.AuditLog.Publish(ctx, auditlog.EventRemoveVisaControl, retResponse, retError, serviceData)
	}()

	entitledCard, err := s.Entitlements.GetEntitledCard(ctx, req.GetTokenizedCardNumber(), entitlements.OPERATION_CARDCONTROLS)
	if err != nil {
		return nil, serviceErr(err, "remove failed")
	}
	serviceData.AccountNumbers = entitledCard.GetAccountNumbers()

	cardNumber, existingControlDocument, err := s.getControlDocument(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, "remove failed")
	}

	serviceData.Last_4Digits = (*cardNumber)[12:]

	if !existingControlDocument.Enrolled() {
		logf.Info(ctx, "Card Not Enrolled")
		return &ccpb.CardControlResponse{}, nil
	}

	request, documentChanged := createUpdateRequest(req.ControlTypes, existingControlDocument)
	if !documentChanged {
		return getCardControlResponse(existingControlDocument), nil
	}

	removeControlResponse, err := s.Visa.UpdateControls(ctx, existingControlDocument.DocumentID, request)
	if err != nil {
		return nil, serviceErr(err, "remove failed")
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardControlsChange)

	return getCardControlResponse(removeControlResponse), nil
}

func createUpdateRequest(controlTypes []ccpb.ControlType, document *visa.Resource) (*visa.Request, bool) {
	documentChanged := false
	for _, controlType := range controlTypes {
		documentChanged = document.RemoveControlByType(controlType)
	}
	if !documentChanged {
		return nil, documentChanged
	}
	return &visa.Request{
		GlobalControls:      document.GlobalControls,
		MerchantControls:    document.MerchantControls,
		TransactionControls: document.TransactionControls,
	}, documentChanged
}

func initServiceData(req *ccpb.RemoveRequest) *servicedata.RemoveVisaControl {
	return &servicedata.RemoveVisaControl{
		TokenizedCardNumber: req.TokenizedCardNumber,
		ControlType:         getRemoveControlTypes(req.ControlTypes),
	}
}

func getRemoveControlTypes(controlRequests []ccpb.ControlType) []string {
	var controls []string
	for _, controlType := range controlRequests {
		controls = append(controls, controlType.String())
	}
	return controls
}

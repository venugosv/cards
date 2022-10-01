package v1beta2

import (
	"context"

	"github.com/anzx/pkg/xcontext"

	"github.com/anzx/fabric-cards/pkg/identity"
	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/pkg/feature"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"
	"github.com/anzx/pkg/auditlog"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

const (
	noTimeRemaining = "00:00:00"
	fortyEightHours = "48:00"
)

func (s server) RemoveControls(ctx context.Context, req *ccpb.RemoveControlsRequest) (retResponse *ccpb.CardControlResponse, retError error) {
	serviceData := initRemoveVisaControlServiceData(req)

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

	var visaCtx context.Context
	if feature.FeatureGate.Enabled(feature.FORGEROCK_SYSTEM_LOGIN) {
		visaCtx, err = s.Forgerock.SystemJWT(ctx, visaGatewayRead, visaGatewayUpdate, visaGatewayDelete)
		if err != nil {
			return nil, serviceErr(err, "remove failed")
		}
	} else {
		visaCtx = ctx
	}

	cardNumber, existingControlDocument, err := s.getControlDocument(ctx, visaCtx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, "remove failed")
	}

	serviceData.Last_4Digits = (*cardNumber)[12:]

	if !customerrules.Enrolled(existingControlDocument) {
		logf.Info(ctx, "Card Not Enrolled")
		return &ccpb.CardControlResponse{}, nil
	}

	if gamblingBlockRequested(req.ControlTypes) {
		var ok bool
		existingControlDocument, ok, err = s.handleGamblingControl(visaCtx, existingControlDocument)
		if err != nil {
			logf.Error(ctx, err, "failed to handle gambling control gracefully")
		}
		if ok {
			req.ControlTypes = removeControlType(req.ControlTypes, ccpb.ControlType_MCT_GAMBLING)
		}
	}

	deleteRequest, ok := customerrules.GetDeleteRequest(existingControlDocument, req.GetControlTypes())
	if !ok {
		return getCardControlResponse(existingControlDocument, req.GetTokenizedCardNumber()), nil
	}

	resource, err := s.Visa.Delete(visaCtx, existingControlDocument.GetDocumentId(), deleteRequest)
	if err != nil {
		return nil, serviceErr(err, "remove failed")
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardControlsChange)

	id, err := identity.Get(ctx)
	if err == nil && id.HasDifferentSubject {
		// This request was likely made by a staff member or coach on customer's behalf, so we should notify customer
		go s.sendNotifications(xcontext.Detach(ctx), req.ControlTypes, id.PersonaID, false)
	}

	return getCardControlResponse(resource, req.GetTokenizedCardNumber()), nil
}

func removeControlType(in []ccpb.ControlType, removeType ccpb.ControlType) []ccpb.ControlType {
	for i, control := range in {
		if control == removeType {
			return append(in[:i], in[i+1:]...)
		}
	}
	return in
}

func (s server) handleGamblingControl(ctx context.Context, controlDocument *crpb.Resource) (*crpb.Resource, bool, error) {
	removeGamblingRequest := false

	gamblingControl, ok := getGamblingControlFromDocument(controlDocument.GetMerchantControls())
	if !ok {
		logf.Info(ctx, "no gambling control found, continuing")
		return controlDocument, removeGamblingRequest, nil
	}
	if canRemoveGamblingControl(gamblingControl) {
		return controlDocument, removeGamblingRequest, nil
	}

	removeGamblingRequest = true

	gamblingControl.ImpulseDelayPeriod = util.ToStringPtr(fortyEightHours)
	request := &crpb.ControlRequest{
		MerchantControls: []*crpb.MerchantControl{
			gamblingControl,
		},
	}
	updatedControlDocument, err := s.Visa.Create(ctx, controlDocument.GetDocumentId(), request)
	if err != nil {
		return nil, removeGamblingRequest, err
	}

	return updatedControlDocument, removeGamblingRequest, nil
}

func gamblingBlockRequested(controls []ccpb.ControlType) bool {
	for _, control := range controls {
		if control == ccpb.ControlType_MCT_GAMBLING {
			return true
		}
	}
	return false
}

func getRemoveControlTypes(controlRequests []ccpb.ControlType) []string {
	var controls []string
	for _, controlType := range controlRequests {
		controls = append(controls, controlType.String())
	}
	return controls
}

func initRemoveVisaControlServiceData(req *ccpb.RemoveControlsRequest) *servicedata.RemoveVisaControl {
	return &servicedata.RemoveVisaControl{
		TokenizedCardNumber: req.TokenizedCardNumber,
		ControlType:         getRemoveControlTypes(req.ControlTypes),
	}
}

// get enable gambling control if any from the existing control document
func getGamblingControlFromDocument(controls []*crpb.MerchantControl) (*crpb.MerchantControl, bool) {
	for _, c := range controls {
		if c.ControlType == ccpb.ControlType_MCT_GAMBLING.String() && c.IsControlEnabled {
			return c, true
		}
	}
	return nil, false
}

// needToSetImpulse, canRemove
func canRemoveGamblingControl(c *crpb.MerchantControl) bool {
	if impulseDelayExists(c) {
		return !impulseDelayActive(c)
	}
	return false
}

// MerchantControl Methods
func impulseDelayActive(c *crpb.MerchantControl) bool {
	return c.ImpulseDelayRemaining != nil && *c.ImpulseDelayRemaining != noTimeRemaining
}

func impulseDelayExists(c *crpb.MerchantControl) bool {
	return c.ImpulseDelayStart != nil &&
		c.ImpulseDelayRemaining != nil &&
		c.ImpulseDelayEnd != nil &&
		c.ImpulseDelayPeriod != nil
}

package v1beta2

import (
	"context"
	"reflect"

	"github.com/anzx/pkg/xcontext"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/pkg/feature"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/identity"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"
	"github.com/anzx/pkg/auditlog"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

const setControlFailed = "set control failed"

func (s server) SetControls(ctx context.Context, req *ccpb.SetControlsRequest) (retResponse *ccpb.CardControlResponse, retError error) {
	serviceData := initSetVisaControlServiceData(req)
	defer func() {
		if err := serviceData.Validate(); err != nil {
			logf.Error(ctx, err, "invalid service data payload")
		}
		s.AuditLog.Publish(ctx, auditlog.EventSetVisaControl, retResponse, retError, serviceData)
	}()

	entitledCard, err := s.Entitlements.GetEntitledCard(ctx, req.GetTokenizedCardNumber(), entitlements.OPERATION_CARDCONTROLS)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}
	serviceData.AccountNumbers = entitledCard.GetAccountNumbers()

	var visaCtx context.Context
	if feature.FeatureGate.Enabled(feature.FORGEROCK_SYSTEM_LOGIN) {
		visaCtx, err = s.Forgerock.SystemJWT(ctx, visaGatewayRead, visaGatewayUpdate, visaGatewayCreate)
		if err != nil {
			return nil, serviceErr(err, setControlFailed)
		}
	} else {
		visaCtx = ctx
	}

	cardNumber, existingControlDocument, err := s.getControlDocument(ctx, visaCtx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	serviceData.Last_4Digits = (*cardNumber)[12:]

	documentID := existingControlDocument.GetDocumentId()
	if !customerrules.Enrolled(existingControlDocument) {
		documentID, err = s.Visa.Registration(visaCtx, *cardNumber)
		if err != nil {
			logf.Error(ctx, err, "unable to enrol card")
			return nil, serviceErr(err, setControlFailed)
		}
	}

	id, err := identity.Get(ctx)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	controls, err := customerrules.WithControls(ctx, req.CardControls, id.PersonaID)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	request := customerrules.ControlRequest(controls...)
	if !existingControlsChanged(request, existingControlDocument) {
		return nil, anzerrors.New(
			codes.AlreadyExists,
			setControlFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlAlreadyExists, "control already exists"),
		)
	}

	// send new controls to customerRulesAPI
	documentResponse, err := s.Visa.Create(visaCtx, documentID, request)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardControlsChange)

	if id.HasDifferentSubject {
		// This request was likely made by a staff member or coach on customer's behalf, so we should notify customer
		controlTypes := make([]ccpb.ControlType, len(req.GetCardControls()))
		for i, controlRequest := range req.GetCardControls() {
			controlTypes[i] = controlRequest.ControlType
		}
		go s.sendNotifications(xcontext.Detach(ctx), controlTypes, id.PersonaID, true)
	}

	return getCardControlResponse(documentResponse, req.TokenizedCardNumber), nil
}

func existingControlsChanged(request *crpb.ControlRequest, existingControls *crpb.Resource) bool {
	globalControlsAreTheSame := reflect.DeepEqual(existingControls.GlobalControls, request.GlobalControls)
	merchantControlsAreTheSame := reflect.DeepEqual(existingControls.MerchantControls, request.MerchantControls)
	transactionControlsAreTheSame := reflect.DeepEqual(existingControls.TransactionControls, request.TransactionControls)

	if globalControlsAreTheSame && merchantControlsAreTheSame && transactionControlsAreTheSame {
		return false // Control Document has not been modified
	}

	return true // Control Document has been modified
}

func getSetControlTypes(controlRequests []*ccpb.ControlRequest) []string {
	var controls []string
	for _, cardControl := range controlRequests {
		controls = append(controls, cardControl.GetControlType().String())
	}
	return controls
}

func initSetVisaControlServiceData(req *ccpb.SetControlsRequest) *servicedata.SetVisaControl {
	return &servicedata.SetVisaControl{
		TokenizedCardNumber: req.TokenizedCardNumber,
		ControlType:         getSetControlTypes(req.CardControls),
	}
}

package v1beta1

import (
	"context"
	"reflect"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/identity"

	"github.com/anzx/pkg/auditlog"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
)

const setControlFailed = "set control failed"

func (s server) Set(ctx context.Context, req *ccpb.SetRequest) (retResponse *ccpb.CardControlResponse, retError error) {
	serviceData := initServiceDta(req)
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

	cardNumber, existingControlDocument, err := s.getControlDocument(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	serviceData.Last_4Digits = (*cardNumber)[12:]

	documentID := existingControlDocument.DocumentID
	if !existingControlDocument.Enrolled() {
		documentID, err = s.Visa.Register(ctx, *cardNumber)
		if err != nil {
			logf.Error(ctx, err, "unable to enrol card")
			return nil, serviceErr(err, setControlFailed)
		}
	}

	id, err := identity.Get(ctx)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	controls, err := visa.WithControls(ctx, req.CardControls, id.PersonaID)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	request := visa.ControlRequest(controls...)
	if !existingControlsChanged(request, existingControlDocument) {
		return nil, anzerrors.New(codes.AlreadyExists, setControlFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlAlreadyExists, "control already exists"))
	}

	// send new controls to customerRulesAPI
	documentResponse, err := s.Visa.CreateControls(ctx, documentID, request)
	if err != nil {
		return nil, serviceErr(err, setControlFailed)
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardControlsChange)

	return getCardControlResponse(documentResponse), nil
}

func existingControlsChanged(request *visa.Request, existingControls *visa.Resource) bool {
	globalControlsAreTheSame := reflect.DeepEqual(existingControls.GlobalControls, request.GlobalControls)
	merchantControlsAreTheSame := reflect.DeepEqual(existingControls.MerchantControls, request.MerchantControls)
	transactionControlsAreTheSame := reflect.DeepEqual(existingControls.TransactionControls, request.TransactionControls)

	if globalControlsAreTheSame && merchantControlsAreTheSame && transactionControlsAreTheSame {
		return false // Control Document has not been modified
	}

	return true // Control Document has been modified
}

func initServiceDta(req *ccpb.SetRequest) *servicedata.SetVisaControl {
	return &servicedata.SetVisaControl{
		TokenizedCardNumber: req.TokenizedCardNumber,
		ControlType:         getSetControlTypes(req.CardControls),
	}
}

func getSetControlTypes(controlRequests []*ccpb.ControlRequest) []string {
	var controls []string
	for _, cardControl := range controlRequests {
		controls = append(controls, cardControl.GetControlType().String())
	}
	return controls
}

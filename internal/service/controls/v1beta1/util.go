package v1beta1

import (
	"github.com/anzx/fabric-cards/pkg/integration/visa"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
)

func getCardControlResponse(controlDocument *visa.Resource) *ccpb.CardControlResponse {
	var response []*ccpb.CardControl

	if controlDocument == nil {
		return &ccpb.CardControlResponse{}
	}

	for _, globalControl := range controlDocument.GlobalControls {
		response = append(response, &ccpb.CardControl{
			ControlType:    ccpb.ControlType_GCT_GLOBAL,
			ControlEnabled: globalControl.ControlEnabled,
		})
	}

	for _, transactionControl := range controlDocument.TransactionControls {
		response = append(response, &ccpb.CardControl{
			ControlType:    ccpb.ControlType(ccpb.ControlType_value[transactionControl.ControlType]),
			ControlEnabled: transactionControl.ControlEnabled,
		})
	}

	for _, merchantControl := range controlDocument.MerchantControls {
		controlType := merchantControl.ControlType
		if controlType == ccpb.ControlType_MCT_GAMBLING.String() {
			response = append(response, &ccpb.CardControl{
				ControlType:        ccpb.ControlType(ccpb.ControlType_value[controlType]),
				ControlEnabled:     merchantControl.ControlEnabled,
				ImpulseDelayStart:  merchantControl.GetImpulseDelayStartTimestamp(),
				ImpulseDelayPeriod: merchantControl.GetImpulseDelayPeriodProto(),
			})
		} else {
			response = append(response, &ccpb.CardControl{
				ControlType:    ccpb.ControlType(ccpb.ControlType_value[controlType]),
				ControlEnabled: merchantControl.ControlEnabled,
			})
		}
	}

	return &ccpb.CardControlResponse{
		CardControls: response,
	}
}

func serviceErr(err error, msg string) error {
	return anzerrors.Wrap(err, anzerrors.GetStatusCode(err), msg, anzerrors.GetErrorInfo(err))
}

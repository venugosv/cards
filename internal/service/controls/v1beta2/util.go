package v1beta2

import (
	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
	anzerrors "github.com/anzx/pkg/errors"
)

func getCardControlResponse(controlDocument *crpb.Resource, tokenizedCardNumber string) *ccpb.CardControlResponse {
	if controlDocument == nil {
		return &ccpb.CardControlResponse{}
	}

	// Use a map to make sure we have at most one of each control type
	count := len(controlDocument.GetGlobalControls()) + len(controlDocument.GetMerchantControls()) + len(controlDocument.GetTransactionControls())
	controlSet := make(map[ccpb.ControlType]*ccpb.CardControl, count)

	for _, globalControl := range controlDocument.GetGlobalControls() {
		if _, hasGlobalControl := controlSet[ccpb.ControlType_GCT_GLOBAL]; hasGlobalControl {
			continue
		}
		if globalControl.GetIsControlEnabled() {
			controlSet[ccpb.ControlType_GCT_GLOBAL] = &ccpb.CardControl{
				ControlType: ccpb.ControlType_GCT_GLOBAL,
			}
		}
	}

	for _, transactionControl := range controlDocument.TransactionControls {
		controlType := ccpb.ControlType(ccpb.ControlType_value[transactionControl.ControlType])
		if _, hasControl := controlSet[controlType]; hasControl {
			continue
		}
		if transactionControl.GetIsControlEnabled() {
			controlSet[controlType] = &ccpb.CardControl{
				ControlType: controlType,
			}
		}
	}

	for _, merchantControl := range controlDocument.MerchantControls {
		controlType := ccpb.ControlType(ccpb.ControlType_value[merchantControl.ControlType])
		if _, hasControl := controlSet[controlType]; hasControl {
			continue
		}
		if merchantControl.GetIsControlEnabled() {
			if controlType == ccpb.ControlType_MCT_GAMBLING {
				controlSet[controlType] = &ccpb.CardControl{
					ControlType:        controlType,
					ImpulseDelayStart:  customerrules.GetImpulseDelayStartTimestamp(merchantControl),
					ImpulseDelayPeriod: customerrules.GetImpulseDelayPeriodProto(merchantControl),
				}
			} else {
				controlSet[controlType] = &ccpb.CardControl{
					ControlType: controlType,
				}
			}
		}
	}

	var out []*ccpb.CardControl
	for _, control := range controlSet {
		out = append(out, control)
	}

	return &ccpb.CardControlResponse{
		TokenizedCardNumber: tokenizedCardNumber,
		CardControls:        out,
	}
}

func serviceErr(err error, msg string) error {
	return anzerrors.Wrap(err, anzerrors.GetStatusCode(err), msg, anzerrors.GetErrorInfo(err))
}

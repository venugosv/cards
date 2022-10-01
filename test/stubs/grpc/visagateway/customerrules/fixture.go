package customerrules

import (
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/util"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"
	"github.com/anzx/fabric-cards/test/data"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
)

const (
	timeFormat          = "2006/01/02 15:04:05"
	defaultImpulseDelay = "48:00"
	DocumentID          = "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"
)

func Resource(builders ...func(*crpb.Resource)) *crpb.Resource {
	out := &crpb.Resource{
		GlobalControls:      []*crpb.GlobalControl{},
		MerchantControls:    []*crpb.MerchantControl{},
		TransactionControls: []*crpb.TransactionControl{},
		DocumentId:          DocumentID,
		LastUpdateTimeStamp: time.Time{}.String(),
	}
	for _, build := range builders {
		build(out)
	}
	return out
}

func WithDocumentID(documentId string) func(*crpb.Resource) {
	return func(r *crpb.Resource) {
		r.DocumentId = documentId
	}
}

func WithGlobalControls() func(*crpb.Resource) {
	return func(r *crpb.Resource) {
		control := createGlobalControl(ccpb.ControlType_GCT_GLOBAL)
		// make sure not append nil to GlobalControls
		if control != nil {
			r.GlobalControls = []*crpb.GlobalControl{control}
		}
	}
}

func WithMerchantControls(controlTypes ...ccpb.ControlType) func(*crpb.Resource) {
	return func(r *crpb.Resource) {
		for _, controlType := range controlTypes {
			control := createMerchantControl(controlType)
			// make sure not append nil to merchantControls
			if control != nil {
				r.MerchantControls = append(r.MerchantControls, control)
			}
		}
	}
}

func WithTransactionControls(controlTypes ...ccpb.ControlType) func(*crpb.Resource) {
	return func(r *crpb.Resource) {
		for _, controlType := range controlTypes {
			control := createTransactionControl(controlType)
			// make sure not append nil to transactionControls
			if control != nil {
				r.TransactionControls = append(r.TransactionControls, control)
			}
		}
	}
}

func WithControls(controlTypes ...ccpb.ControlType) []func(*crpb.Resource) {
	var out []func(*crpb.Resource)
	for _, controlType := range controlTypes {
		category := customerrules.GetCategory(controlType)

		switch category {
		case customerrules.GLOBAL:
			out = append(out, WithGlobalControls())
		case customerrules.TRANSACTION:
			out = append(out, WithTransactionControls(controlTypes...))
		case customerrules.MERCHANT:
			out = append(out, WithMerchantControls(controlTypes...))
		}
	}
	return out
}

func TransactionControlDocumentResponse(resource *crpb.Resource) *crpb.TransactionControlDocumentResponse {
	return &crpb.TransactionControlDocumentResponse{
		ReceivedTimestamp:  time.Time{}.String(),
		ProcessingTimeInMs: 10,
		Resource:           resource,
	}
}

func TransactionControlDocumentList(resources ...*crpb.Resource) *crpb.TransactionControlList {
	return &crpb.TransactionControlList{
		ReceivedTimestamp:  time.Time{}.String(),
		ProcessingTimeInMs: 10,
		Resource: &crpb.RepeatedResource{
			ControlDocuments: resources,
		},
	}
}

func createGlobalControl(controlType ccpb.ControlType) *crpb.GlobalControl {
	if customerrules.GetCategory(controlType) != customerrules.GLOBAL {
		return nil
	}
	return &crpb.GlobalControl{
		ShouldDeclineAll:                  util.ToBoolPtr(true),
		IsControlEnabled:                  true,
		UserIdentifier:                    &data.AUser().PersonaID,
		DeclineAllNonTokenizeTransactions: util.ToBoolPtr(true),
		ShouldAlertOnDecline:              util.ToBoolPtr(true),
	}
}

func createMerchantControl(controlType ccpb.ControlType) *crpb.MerchantControl {
	if customerrules.GetCategory(controlType) != customerrules.MERCHANT {
		return nil
	}
	return &crpb.MerchantControl{
		ShouldDeclineAll:     util.ToBoolPtr(true),
		IsControlEnabled:     true,
		UserIdentifier:       &data.AUser().PersonaID,
		ControlType:          controlType.String(),
		ShouldAlertOnDecline: util.ToBoolPtr(true),
	}
}

func createTransactionControl(controlType ccpb.ControlType) *crpb.TransactionControl {
	if customerrules.GetCategory(controlType) != customerrules.TRANSACTION {
		return nil
	}
	return &crpb.TransactionControl{
		ShouldDeclineAll:     util.ToBoolPtr(true),
		IsControlEnabled:     true,
		ControlType:          controlType.String(),
		UserIdentifier:       &data.AUser().PersonaID,
		ShouldAlertOnDecline: util.ToBoolPtr(true),
	}
}

func AccountUpdateResponse(builders ...func(*crpb.UpdateAccountResponse)) *crpb.UpdateAccountResponse {
	out := &crpb.UpdateAccountResponse{
		ReceivedTimestamp:  time.Time{}.String(),
		ProcessingTimeInMs: 10,
		Resource: &crpb.UpdateAccountResponse_UpdatedResource{
			Status: "SUCCESS",
		},
	}
	for _, build := range builders {
		build(out)
	}
	return out
}

func WithStatus(status string) func(*crpb.UpdateAccountResponse) {
	return func(d *crpb.UpdateAccountResponse) {
		d.Resource.Status = status
	}
}

func WithImpulseDelay(start, remaining string) func(*crpb.Resource) {
	return func(resource *crpb.Resource) {
		length := 48 * time.Hour
		for _, mct := range resource.GetMerchantControls() {
			impulseStart, _ := time.Parse(timeFormat, start)
			impulseEnd := impulseStart.Add(length).Format(timeFormat)
			mct.ImpulseDelayStart = &start
			mct.ImpulseDelayPeriod = util.ToStringPtr(defaultImpulseDelay)
			mct.ImpulseDelayRemaining = &remaining
			mct.ImpulseDelayEnd = &impulseEnd
		}
	}
}

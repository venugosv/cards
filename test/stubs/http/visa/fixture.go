package visa

import (
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

const DocumentID = "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"

type DocumentResponseFixture struct {
	documentResponse visa.Resource
}

type TransactionControlDocumentResponseFixture struct {
	transactionControlDocument visa.TransactionControlDocument
}

type AccountUpdateResponseFixture struct {
	AccountUpdateResponse visa.AccountUpdateResponse
}

func NewResourceFixture() DocumentResponseFixture {
	return DocumentResponseFixture{
		documentResponse: visa.Resource{
			GlobalControls:      []*visa.GlobalControl{},
			MerchantControls:    []*visa.MerchantControl{},
			TransactionControls: []*visa.TransactionControl{},
			DocumentID:          DocumentID,
			LastUpdateTimeStamp: time.Time{}.String(),
		},
	}
}

func NewAccountUpdateResponseFixture() AccountUpdateResponseFixture {
	return AccountUpdateResponseFixture{
		AccountUpdateResponse: visa.AccountUpdateResponse{
			ReceivedTimestamp:  time.Time{}.String(),
			ProcessingTimeinMS: 10,
			Resource: visa.StatusResource{
				Status: "SUCCESS",
			},
		},
	}
}

func NewTransactionControlDocumentResponseFixture(resource visa.Resource) TransactionControlDocumentResponseFixture {
	return TransactionControlDocumentResponseFixture{
		transactionControlDocument: visa.TransactionControlDocument{
			ReceivedTimestamp:  time.Time{}.String(),
			ProcessingTimeInMS: 10,
			Resource:           resource,
		},
	}
}

type TransactionControlDocumentResponseListFixture struct {
	transactionControlDocumentList visa.TransactionControlListResponses
}

func NewTransactionControlDocumentListResponseFixture(resource ...visa.Resource) TransactionControlDocumentResponseListFixture {
	return TransactionControlDocumentResponseListFixture{
		transactionControlDocumentList: visa.TransactionControlListResponses{
			ReceivedTimestamp:  time.Time{}.String(),
			ProcessingTimeInMS: 10,
			Resource: visa.ListResource{
				ControlDocuments: resource,
			},
		},
	}
}

func (d DocumentResponseFixture) WithDocumentID(documentId string) *DocumentResponseFixture {
	d.documentResponse.DocumentID = documentId
	return &d
}

func (d DocumentResponseFixture) WithGlobalControls() *DocumentResponseFixture {
	control := createGlobalControl(ccpb.ControlType_GCT_GLOBAL)
	d.documentResponse.GlobalControls = []*visa.GlobalControl{control}
	return &d
}

func (d DocumentResponseFixture) WithMerchantControls(controlTypes ...ccpb.ControlType) *DocumentResponseFixture {
	for _, controlType := range controlTypes {
		control := createMerchantControl(controlType)
		if controlType == ccpb.ControlType_MCT_GAMBLING {
			control.ImpulseDelayStart = func(s string) *string { return &s }("2020-05-18 23:34:50")
			control.ImpulseDelayPeriod = func(s string) *string { return &s }("24:00")
		}
		d.documentResponse.MerchantControls = append(d.documentResponse.MerchantControls, control)
	}
	return &d
}

func (d DocumentResponseFixture) WithTransactionControls(controlTypes ...ccpb.ControlType) *DocumentResponseFixture {
	for _, controlType := range controlTypes {
		control := createTransactionControl(controlType)
		d.documentResponse.TransactionControls = append(d.documentResponse.TransactionControls, control)
	}
	return &d
}

func (d DocumentResponseFixture) Build() visa.Resource {
	return d.documentResponse
}

func (d DocumentResponseFixture) BuildPointer() *visa.Resource {
	return &d.documentResponse
}

func createGlobalControl(controlType ccpb.ControlType) *visa.GlobalControl {
	if visa.GetCategory(controlType) != visa.GLOBAL {
		return nil
	}
	return &visa.GlobalControl{
		ShouldDeclineAll:                  true,
		ControlEnabled:                    true,
		UserIdentifier:                    data.AUser().PersonaID,
		AlertThreshold:                    util.ToFloat64Ptr(15),
		DeclineThreshold:                  util.ToFloat64Ptr(0),
		DeclineAllNonTokenizeTransactions: true,
		ShouldAlertOnDecline:              false,
	}
}

func createMerchantControl(controlType ccpb.ControlType) *visa.MerchantControl {
	if visa.GetCategory(controlType) != visa.MERCHANT {
		return nil
	}
	return &visa.MerchantControl{
		ShouldDeclineAll:     true,
		ControlEnabled:       true,
		UserIdentifier:       data.AUser().PersonaID,
		ControlType:          controlType.String(),
		AlertThreshold:       util.ToFloat64Ptr(15),
		DeclineThreshold:     util.ToFloat64Ptr(0),
		ShouldAlertOnDecline: false,
	}
}

func createTransactionControl(controlType ccpb.ControlType) *visa.TransactionControl {
	if visa.GetCategory(controlType) != visa.TRANSACTION {
		return nil
	}
	return &visa.TransactionControl{
		ShouldDeclineAll:     true,
		ControlEnabled:       true,
		ControlType:          controlType.String(),
		UserIdentifier:       data.AUser().PersonaID,
		AlertThreshold:       util.ToFloat64Ptr(15),
		DeclineThreshold:     util.ToFloat64Ptr(0),
		ShouldAlertOnDecline: false,
	}
}

func (d DocumentResponseFixture) WithControls(controlTypes ...ccpb.ControlType) *DocumentResponseFixture {
	for _, controlType := range controlTypes {
		category := visa.GetCategory(controlType)

		switch category {
		case visa.GLOBAL:
			return d.WithGlobalControls()
		case visa.TRANSACTION:
			return d.WithTransactionControls(controlTypes...)
		case visa.MERCHANT:
			return d.WithMerchantControls(controlTypes...)
		}
	}
	return &d
}

func (d AccountUpdateResponseFixture) WithStatus(status string) *AccountUpdateResponseFixture {
	d.AccountUpdateResponse.Resource.Status = status
	return &d
}

func (d AccountUpdateResponseFixture) Build() *visa.AccountUpdateResponse {
	return &d.AccountUpdateResponse
}

func (d TransactionControlDocumentResponseFixture) Build() *visa.TransactionControlDocument {
	return &d.transactionControlDocument
}

func (d TransactionControlDocumentResponseListFixture) Build() *visa.TransactionControlListResponses {
	return &d.transactionControlDocumentList
}

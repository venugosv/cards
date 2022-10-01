package visa

import (
	"context"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

type StubClient struct {
	testingData  *data.Data
	EnrolError   error
	QueryError   error
	CreateError  error
	ReplaceError error
	UpdateError  error
	Controls     []ccpb.ControlType
}

// NewStubClient creates a CardEntitlementsAPIClient stubs
func NewStubClient(testData *data.Data) StubClient {
	return StubClient{
		testingData: testData,
	}
}

func (d StubClient) Register(_ context.Context, primaryAccountNumber string) (string, error) {
	if d.EnrolError != nil {
		return "", d.EnrolError
	}
	testingCardItem := d.testingData.GetCardByCardNumber(primaryAccountNumber)

	if testingCardItem.CardControls == data.CardControlsPresetNotEnrolled {
		return "NOT_ENROLLED", nil
	}
	return "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b", nil
}

func (d StubClient) QueryControls(_ context.Context, primaryAccountNumber string) (*visa.Resource, error) {
	if d.QueryError != nil {
		return nil, d.QueryError
	}

	var response *visa.Resource

	if d.Controls != nil {
		response := NewResourceFixture().WithControls(d.Controls...).Build()
		return &response, nil
	}

	testingCardItem := d.testingData.GetCardByCardNumber(primaryAccountNumber)

	switch testingCardItem.CardControls {
	case data.CardControlsPresetAllControls: // all controls
		merchantControls := []ccpb.ControlType{ccpb.ControlType_MCT_ALCOHOL, ccpb.ControlType_MCT_ADULT_ENTERTAINMENT, ccpb.ControlType_MCT_SMOKE_AND_TOBACCO}
		transactionControls := []ccpb.ControlType{ccpb.ControlType_TCT_E_COMMERCE, ccpb.ControlType_TCT_AUTO_PAY, ccpb.ControlType_TCT_ATM_WITHDRAW}
		response = NewResourceFixture().WithGlobalControls().WithMerchantControls(merchantControls...).WithTransactionControls(transactionControls...).BuildPointer()

	case data.CardControlsPresetGlobalControls: // global controls
		response = NewResourceFixture().WithGlobalControls().BuildPointer()

	case data.CardControlsPresetContactlessControl: // global controls
		response = NewResourceFixture().WithTransactionControls(ccpb.ControlType_TCT_CONTACTLESS).BuildPointer()

	case data.CardControlsPresetNoControls:
		response = NewResourceFixture().BuildPointer()

	case data.CardControlsPresetNotEnrolled:
		response = NewResourceFixture().WithDocumentID("NOT_ENROLLED").BuildPointer()

	case data.CardControlsPresetCanNotBeEnrolled:
		response = nil

	default:
		response = NewResourceFixture().WithDocumentID("NOT_ENROLLED").BuildPointer()
	}
	return response, nil
}

func (d StubClient) CreateControls(_ context.Context, documentID string, request *visa.Request) (*visa.Resource, error) {
	if d.CreateError != nil {
		return nil, d.CreateError
	}
	controls := replaceControls(documentID, request)
	return &controls, nil
}

func (d StubClient) UpdateControls(_ context.Context, documentID string, request *visa.Request) (*visa.Resource, error) {
	if d.UpdateError != nil {
		return nil, d.UpdateError
	}
	controls := replaceControls(documentID, request)

	return &controls, nil
}

func (d StubClient) DeleteControls(_ context.Context, documentID string, request *visa.Request) (*visa.Resource, error) {
	if d.UpdateError != nil {
		return nil, d.UpdateError
	}
	controls := replaceControls(documentID, request)

	return &controls, nil
}

func (d StubClient) ReplaceCard(_ context.Context, _, newAccountID string) (bool, error) {
	if d.ReplaceError != nil {
		return false, d.ReplaceError
	}

	newCardItem := d.testingData.GetCardByCardNumber(newAccountID)

	switch newCardItem.CardControls {
	case data.CardControlsPresetCanNotBeEnrolled:
		return false, nil
	default:
		return true, nil
	}
}

func replaceControls(documentID string, controlDocument *visa.Request) visa.Resource {
	response := NewResourceFixture().WithDocumentID(documentID).Build()
	response.GlobalControls = controlDocument.GlobalControls
	response.MerchantControls = controlDocument.MerchantControls
	response.TransactionControls = controlDocument.TransactionControls
	return response
}

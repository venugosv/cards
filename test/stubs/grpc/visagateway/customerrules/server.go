package customerrules

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"

	"github.com/anzx/fabric-cards/test/data"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
)

type StubServer struct {
	crpb.UnimplementedCustomerRulesAPIServer
	data                          *data.Data
	Controls                      []ccpb.ControlType
	GamblingImpulseDelayStart     string
	GamblingImpulseDelayRemaining string
}

func (s StubServer) Register(ctx context.Context, request *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error) {
	testingCardItem := s.data.GetCardByCardNumber(request.GetPrimaryAccountNumber())

	var docID string
	if testingCardItem.CardControls == data.CardControlsPresetNotEnrolled {
		docID = "NOT_ENROLLED"
	} else {
		docID = "ctc-vd-857a8766-160b-498d-820f-bf4339949c1b"
	}

	return &crpb.TransactionControlDocumentResponse{
		Resource: &crpb.Resource{
			DocumentId: docID,
		},
	}, nil
}

func (s StubServer) ListControlDocuments(_ context.Context, request *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
	if s.Controls != nil {
		builders := WithControls(s.Controls...)
		if s.GamblingImpulseDelayStart != "" {
			builders = append(builders, WithImpulseDelay(s.GamblingImpulseDelayStart, s.GamblingImpulseDelayRemaining))
		}
		return TransactionControlDocumentList(Resource(builders...)), nil
	}

	switch s.data.GetCardByCardNumber(request.GetPrimaryAccountNumber()).CardControls {
	case data.CardControlsPresetAllControls:
		return TransactionControlDocumentList(Resource(
			WithGlobalControls(),
			WithMerchantControls(
				ccpb.ControlType_MCT_ALCOHOL,
				ccpb.ControlType_MCT_ADULT_ENTERTAINMENT,
				ccpb.ControlType_MCT_SMOKE_AND_TOBACCO),
			WithTransactionControls(
				ccpb.ControlType_TCT_E_COMMERCE,
				ccpb.ControlType_TCT_AUTO_PAY,
				ccpb.ControlType_TCT_ATM_WITHDRAW))), nil

	case data.CardControlsPresetGlobalControls:
		return TransactionControlDocumentList(Resource(
			WithGlobalControls())), nil

	case data.CardControlsPresetContactlessControl:
		return TransactionControlDocumentList(Resource(
			WithTransactionControls(ccpb.ControlType_TCT_CONTACTLESS))), nil

	case data.CardControlsPresetNoControls:
		return TransactionControlDocumentList(Resource()), nil

	case data.CardControlsPresetNotEnrolled:
		return TransactionControlDocumentList(Resource(
			WithDocumentID("NOT_ENROLLED"))), nil

	case data.CardControlsPresetCanNotBeEnrolled:
		return TransactionControlDocumentList(nil), nil

	default:
		return TransactionControlDocumentList(Resource(
			WithDocumentID("NOT_ENROLLED"))), nil
	}
}

func (s StubServer) CreateControls(_ context.Context, request *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
	return replaceControls(request.GetDocumentId(), request.GetRequest()), nil
}

func (s StubServer) UpdateControls(_ context.Context, request *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
	return updateControls(request.GetDocumentId(), request.GetRequest()), nil
}

func (s StubServer) DeleteControls(_ context.Context, request *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
	in := request.GetRequest()
	builders := WithControls(s.Controls...)
	if s.GamblingImpulseDelayStart != "" {
		builders = append(builders, WithImpulseDelay(s.GamblingImpulseDelayStart, s.GamblingImpulseDelayRemaining))
	}
	resource := Resource(builders...)
	if in.GlobalControls != nil {
		resource.GlobalControls = []*crpb.GlobalControl{}
	}

	for _, inControl := range in.MerchantControls {
		for i, serverControl := range resource.MerchantControls {
			if inControl.GetControlType() == serverControl.GetControlType() {
				resource.MerchantControls = append(resource.MerchantControls[:i], resource.MerchantControls[i+1:]...)
				break
			}
		}
	}

	for _, inControl := range in.TransactionControls {
		for i, serverControl := range resource.TransactionControls {
			if inControl.GetControlType() == serverControl.GetControlType() {
				resource.TransactionControls = append(resource.TransactionControls[:i], resource.TransactionControls[i+1:]...)
				break
			}
		}
	}

	replaceRequest := &crpb.ControlRequest{
		GlobalControls:      resource.GlobalControls,
		MerchantControls:    resource.MerchantControls,
		TransactionControls: resource.TransactionControls,
	}
	return replaceControls(request.GetDocumentId(), replaceRequest), nil
}

func (s StubServer) UpdateAccount(_ context.Context, request *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
	newCardItem := s.data.GetCardByCardNumber(request.GetNewAccountId())

	switch newCardItem.CardControls {
	case data.CardControlsPresetCanNotBeEnrolled:
		return AccountUpdateResponse(WithStatus("FAILED")), nil
	default:
		return AccountUpdateResponse(WithStatus("SUCCESS")), nil
	}
}

// NewStubServer creates a CardEntitlementsAPIClient stubs
func NewStubServer(data *data.Data) StubServer {
	return StubServer{
		data: data,
	}
}

func replaceControls(documentID string, in *crpb.ControlRequest) *crpb.TransactionControlDocumentResponse {
	response := Resource(WithDocumentID(documentID))
	response.GlobalControls = in.GlobalControls
	response.MerchantControls = in.MerchantControls
	response.TransactionControls = in.TransactionControls

	return TransactionControlDocumentResponse(response)
}

func updateControls(documentID string, in *crpb.ControlRequest) *crpb.TransactionControlDocumentResponse {
	for _, v := range in.MerchantControls {
		if v.ControlType == ccpb.ControlType_MCT_GAMBLING.String() {
			v.ImpulseDelayStart = util.ToStringPtr("2020-05-19 07:34:50")
			v.ImpulseDelayPeriod = util.ToStringPtr("48:00")
		}
	}

	return replaceControls(documentID, in)
}

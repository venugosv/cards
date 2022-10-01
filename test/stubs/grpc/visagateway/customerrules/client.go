package customerrules

import (
	"context"

	"github.com/anzx/fabric-cards/test/data"

	"google.golang.org/grpc"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
)

type StubClient struct {
	RegistrationError      error
	ListError              error
	GetError               error
	CreateError            error
	UpdateError            error
	DeleteError            error
	ReplaceError           error
	FixedResponse          *crpb.Resource
	CustomerRulesAPIServer StubServer
}

// NewStubClient creates a CustomerRulesAPIClient stubs
func NewStubClient(data *data.Data) StubClient {
	return StubClient{
		CustomerRulesAPIServer: NewStubServer(data),
	}
}

func (s StubClient) Register(ctx context.Context, in *crpb.RegisterRequest, _ ...grpc.CallOption) (*crpb.TransactionControlDocumentResponse, error) {
	if s.RegistrationError != nil {
		return nil, s.RegistrationError
	}
	return s.CustomerRulesAPIServer.Register(ctx, in)
}

func (s StubClient) ListControlDocuments(ctx context.Context, in *crpb.ListControlDocumentsRequest, _ ...grpc.CallOption) (*crpb.TransactionControlList, error) {
	if s.ListError != nil {
		return nil, s.ListError
	}
	if s.FixedResponse != nil {
		return &crpb.TransactionControlList{
			Resource: &crpb.RepeatedResource{
				ControlDocuments: []*crpb.Resource{
					s.FixedResponse,
				},
			},
		}, nil
	}
	return s.CustomerRulesAPIServer.ListControlDocuments(ctx, in)
}

func (s StubClient) GetControlDocument(ctx context.Context, in *crpb.GetControlDocumentRequest, _ ...grpc.CallOption) (*crpb.TransactionControlDocumentResponse, error) {
	if s.GetError != nil {
		return nil, s.GetError
	}
	return s.CustomerRulesAPIServer.GetControlDocument(ctx, in)
}

func (s StubClient) CreateControls(ctx context.Context, in *crpb.TransactionControlDocumentRequest, _ ...grpc.CallOption) (*crpb.TransactionControlDocumentResponse, error) {
	if s.CreateError != nil {
		return nil, s.CreateError
	}
	return s.CustomerRulesAPIServer.CreateControls(ctx, in)
}

func (s StubClient) UpdateControls(ctx context.Context, in *crpb.TransactionControlDocumentRequest, _ ...grpc.CallOption) (*crpb.TransactionControlDocumentResponse, error) {
	if s.UpdateError != nil {
		return nil, s.UpdateError
	}
	return s.CustomerRulesAPIServer.UpdateControls(ctx, in)
}

func (s StubClient) DeleteControls(ctx context.Context, in *crpb.TransactionControlDocumentRequest, _ ...grpc.CallOption) (*crpb.TransactionControlDocumentResponse, error) {
	if s.DeleteError != nil {
		return nil, s.DeleteError
	}
	return s.CustomerRulesAPIServer.DeleteControls(ctx, in)
}

func (s StubClient) UpdateAccount(ctx context.Context, in *crpb.UpdateAccountRequest, _ ...grpc.CallOption) (*crpb.UpdateAccountResponse, error) {
	if s.ReplaceError != nil {
		return nil, s.ReplaceError
	}
	return s.CustomerRulesAPIServer.UpdateAccount(ctx, in)
}

package accounts

import (
	"context"

	"google.golang.org/grpc"

	apb "github.com/anzx/fabricapis/pkg/fabric/service/accounts/v1alpha6"

	"github.com/anzx/fabric-cards/test/data"
)

type StubClient struct {
	testingData *data.Data
	response    *apb.GetAccountListResponse
	error       error
}

// NewStubClient creates a PartyAPIClient stubs
func NewStubClient(testData *data.Data) apb.AccountAPIClient {
	return &StubClient{
		testingData: testData,
	}
}

func (s StubClient) GetAccountList(_ context.Context, _ *apb.GetAccountListRequest, _ ...grpc.CallOption) (*apb.GetAccountListResponse, error) {
	if s.error != nil {
		return nil, s.error
	}

	if s.response != nil {
		return s.response, nil
	}

	return &apb.GetAccountListResponse{
		AccountList: []*apb.AccountDetails{
			{
				AccountNumber: "1234567890",
			},
		},
	}, nil
}

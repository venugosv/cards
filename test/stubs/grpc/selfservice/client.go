package selfservice

import (
	"context"

	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"

	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/test/data"
)

type StubClient struct {
	testingData      *data.Data
	GetPartyResponse *sspb.GetPartyResponse
	GetPartyError    error
}

// NewStubClient creates a PartyAPIClient stubs
func NewStubClient(testData *data.Data) sspb.PartyAPIClient {
	return &StubClient{
		testingData: testData,
	}
}

func (s StubClient) GetParty(_ context.Context, _ *sspb.GetPartyRequest, _ ...grpc.CallOption) (*sspb.GetPartyResponse, error) {
	if s.GetPartyError != nil {
		return nil, s.GetPartyError
	}

	if s.GetPartyResponse != nil {
		return s.GetPartyResponse, nil
	}

	return &sspb.GetPartyResponse{
		LegalName: &sspb.Name{
			Name:       "Ms. Oprah Gail Winfrey",
			Prefix:     "Queen",
			Title:      "Ms",
			FirstName:  "Oprah",
			MiddleName: "Gail",
			LastName:   "Winfrey",
		},
		ResidentialAddress: &sspb.Address{
			LineOne:    "Level 13",
			LineTwo:    "839 Collins Street",
			City:       "Docklands",
			PostalCode: "3008",
			State:      "VIC",
			Country:    "AUS",
		},
		MailingAddress: &sspb.Address{
			LineOne:    "Mailroom",
			LineTwo:    "833 Collins Street",
			City:       "Docklands",
			PostalCode: "3008",
			State:      "VIC",
			Country:    "AUS",
		},
	}, nil
}

func (s StubClient) UpdateParty(ctx context.Context, in *sspb.UpdatePartyRequest, opts ...grpc.CallOption) (*sspb.UpdatePartyResponse, error) {
	panic("implement me")
}

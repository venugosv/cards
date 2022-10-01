package selfservice

import (
	"context"

	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
)

type StubServer struct {
	sspb.UnimplementedPartyAPIServer
}

// NewStubServer creates a CardEntitlementsAPIClient stubs
func NewStubServer() sspb.PartyAPIServer {
	return &StubServer{}
}

func (s *StubServer) GetParty(_ context.Context, _ *sspb.GetPartyRequest) (*sspb.GetPartyResponse, error) {
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

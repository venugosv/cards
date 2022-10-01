package eligibility

import (
	"context"

	anzcodes "github.com/anzx/pkg/errors/errcodes"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/identity"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/test/stubs/http/ctm"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type StubClient struct {
	*mock.Mock
	testingData *data.Data
	CanErr      error
}

func NewStubClient(testingData *data.Data) StubClient {
	return StubClient{
		testingData: testingData,
	}
}

func (m StubClient) Can(ctx context.Context, in *epb.CanRequest, _ ...grpc.CallOption) (*epb.CanResponse, error) {
	if m.CanErr != nil {
		return nil, m.CanErr
	}

	user, err := identity.Get(ctx)
	if err != nil {
		return nil, err
	}

	cardDetails, err := ctm.GetCardDetails(m.testingData, in.TokenizedCardNumber, user.PersonaID)
	if err != nil {
		return nil, err
	}

	if eligible := cardDetails.HasEligibility(in.Eligibility); !eligible {
		return nil, anzerrors.New(codes.PermissionDenied, "eligibility failed", anzerrors.NewErrorInfo(context.Background(), anzcodes.CardIneligible, "card not eligible"))
	}
	return &epb.CanResponse{}, nil
}

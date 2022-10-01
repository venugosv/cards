package entitlements

import (
	"context"

	anzcodes "github.com/anzx/pkg/errors/errcodes"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/identity"
	"google.golang.org/grpc"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
)

type StubClient struct {
	testingData *data.Data
	StubCardsClient
	StubControlAPIClient
}

type StubCardsClient struct {
	entpb.CardEntitlementsAPIClient
	GetEntitledCardErr   error
	ListEntitledCardsErr error
}

type StubControlAPIClient struct {
	entpb.EntitlementsControlAPIClient
	RegisterCardToPersonaErr error
	ForcePartyToLatestErr    error
}

// NewStubClient creates a CardEntitlementsAPIClient stubs
func NewStubClient(testData *data.Data) StubClient {
	return StubClient{
		testingData: testData,
	}
}

func (s StubClient) GetEntitledCard(ctx context.Context, in *entpb.GetEntitledCardRequest, _ ...grpc.CallOption) (*entpb.EntitledCard, error) {
	if s.GetEntitledCardErr != nil {
		return nil, s.GetEntitledCardErr
	}

	user, err := identity.Get(ctx)
	if err != nil {
		return nil, err
	}

	userID, card := s.testingData.GetCardByToken(in.TokenizedCardNumber)

	if card == nil || (userID != user.PersonaID) {
		return nil, anzerrors.New(codes.PermissionDenied, "entitlements failed", anzerrors.NewErrorInfo(ctx, anzcodes.CardIneligible, "user not entitled"))
	}

	return &entpb.EntitledCard{
		TokenizedCardNumber: card.Token,
		AccountNumbers:      card.AccountNumbers,
	}, nil
}

func (s StubClient) ListEntitledCards(ctx context.Context, in *entpb.ListEntitledCardsRequest, _ ...grpc.CallOption) (*entpb.ListEntitledCardsResponse, error) {
	if s.ListEntitledCardsErr != nil {
		return nil, s.ListEntitledCardsErr
	}

	id, err := identity.Get(ctx)
	if err != nil {
		return nil, err
	}

	user := s.testingData.GetUserByPersonaID(id.PersonaID)
	if user == nil {
		return nil, anzerrors.New(codes.NotFound, "entitlements failed", anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, "cannot find user data"))
	}

	cards := make([]*entpb.EntitledCard, 0, len(user.Cards))
	for _, card := range user.Cards {
		cards = append(cards, &entpb.EntitledCard{
			TokenizedCardNumber: card.Token,
			AccountNumbers:      card.AccountNumbers,
		})
	}

	return &entpb.ListEntitledCardsResponse{
		Cards: cards,
	}, nil
}

func (s StubClient) ListPersonaForCardToken(ctx context.Context, _ *entpb.ListPersonaForCardTokenRequest, _ ...grpc.CallOption) (*entpb.ListPersonaForCardTokenResponse, error) {
	return nil, nil
}

func NewControlAPIStubClient() StubControlAPIClient {
	return StubControlAPIClient{}
}

func (e StubControlAPIClient) RegisterCardToPersona(ctx context.Context, in *entpb.RegisterCardToPersonaRequest, opts ...grpc.CallOption) (*entpb.RegisterCardToPersonaResponse, error) {
	return nil, e.RegisterCardToPersonaErr
}

func (e StubControlAPIClient) ForcePartyToLatest(ctx context.Context, in *entpb.ForcePartyToLatestRequest, opts ...grpc.CallOption) (*entpb.ForcePartyToLatestResponse, error) {
	return nil, e.ForcePartyToLatestErr
}

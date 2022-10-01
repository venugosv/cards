package entitlements

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/anzdata"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/identity"
	"github.com/anzx/fabric-cards/test/stubs/utils"
	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
)

type StubServer struct {
	StubCardEntitlementsAPIServer
	StubEntitlementsControlAPIServer
}

type StubCardEntitlementsAPIServer struct {
	entpb.UnimplementedCardEntitlementsAPIServer
	store *utils.Store
}

type StubEntitlementsControlAPIServer struct {
	entpb.UnimplementedEntitlementsControlAPIServer
	store *utils.Store
}

// NewStubServer creates a CardEntitlementsAPIClient stubs
func NewStubServer(ctx context.Context) entpb.CardEntitlementsAPIServer {
	return &StubCardEntitlementsAPIServer{
		store: utils.GetStore(ctx),
	}
}

func NewStubControlAPIServer(ctx context.Context) entpb.EntitlementsControlAPIServer {
	return &StubEntitlementsControlAPIServer{
		store: utils.GetStore(ctx),
	}
}

func (e StubCardEntitlementsAPIServer) GetEntitledCard(ctx context.Context, request *entpb.GetEntitledCardRequest) (*entpb.EntitledCard, error) {
	user, err := getUserData(ctx, e.store)
	if err != nil {
		return nil, err
	}

	var response *entpb.EntitledCard
	for _, card := range user.Cards {
		if card.Token == request.GetTokenizedCardNumber() {
			response = &entpb.EntitledCard{
				TokenizedCardNumber: card.Token,
				AccountNumbers:      card.AccountNumbers,
			}
		}
	}

	if response == nil {
		return nil, anzerrors.New(codes.NotFound, "entitlements failed", anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, "cannot find user data"))
	}

	return response, nil
}

func (e StubCardEntitlementsAPIServer) ListEntitledCards(ctx context.Context, _ *entpb.ListEntitledCardsRequest) (*entpb.ListEntitledCardsResponse, error) {
	user, err := getUserData(ctx, e.store)
	if err != nil {
		return nil, err
	}

	var cards []*entpb.EntitledCard
	for _, item := range user.Cards {
		card := &entpb.EntitledCard{
			TokenizedCardNumber: item.Token,
			AccountNumbers:      item.AccountNumbers,
		}
		cards = append(cards, card)
	}

	if len(cards) < 1 {
		return nil, anzerrors.New(codes.NotFound, "entitlements failed", anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, "cannot find user data"))
	}

	return &entpb.ListEntitledCardsResponse{Cards: cards}, nil
}

func getUserData(ctx context.Context, store *utils.Store) (anzdata.User, error) {
	id, err := identity.Get(ctx)
	if err != nil {
		logf.Error(ctx, err, "failed to get user data")
		return anzdata.User{}, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), "entitlements failed", anzerrors.GetErrorInfo(err))
	}

	return store.GetUser(id.PersonaID), nil
}

func (e StubEntitlementsControlAPIServer) RegisterCardToPersona(ctx context.Context, in *entpb.RegisterCardToPersonaRequest) (*entpb.RegisterCardToPersonaResponse, error) {
	user, err := getUserData(ctx, e.store)
	if err != nil {
		return nil, err
	}

	user.Cards = append(user.Cards, anzdata.Card{
		Number:         in.GetTokenizedCardNumber(),
		Token:          in.GetTokenizedCardNumber(),
		AccountNumbers: user.Cards[0].AccountNumbers,
	})

	e.store.SaveUser(user.PersonaID, user)

	return &entpb.RegisterCardToPersonaResponse{}, nil
}

func (e StubEntitlementsControlAPIServer) ForcePartyToLatest(ctx context.Context, request *entpb.ForcePartyToLatestRequest) (*entpb.ForcePartyToLatestResponse, error) {
	return &entpb.ForcePartyToLatestResponse{}, nil
}

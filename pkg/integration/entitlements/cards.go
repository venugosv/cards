package entitlements

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"google.golang.org/grpc/codes"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/pkg/errors"
)

func (c Client) GetEntitledCard(ctx context.Context, tokenizedCardNumber string, operations ...string) (*entpb.EntitledCard, error) {
	req := &entpb.GetEntitledCardRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		Operations:          operations,
		CardView:            entpb.CardView_CARD_VIEW_ACCOUNT_NUMBERS,
	}

	entitledCard, err := c.CardEntitlementsAPIClient.GetEntitledCard(ctx, req)
	if err != nil {
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), "Entitlements/GetEntitledCard failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, anzerrors.GetErrorInfo(err).GetReason()))
	}

	return entitledCard, nil
}

func (c Client) ListEntitledCards(ctx context.Context) ([]*entpb.EntitledCard, error) {
	req := &entpb.ListEntitledCardsRequest{
		Operations: []string{
			OPERATION_VIEW_CARD,
		},
		CardView: entpb.CardView_CARD_VIEW_ACCOUNT_NUMBERS,
	}

	response, err := c.CardEntitlementsAPIClient.ListEntitledCards(ctx, req)
	if err != nil {
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), "Entitlements ListEntitledCards failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, anzerrors.GetErrorInfo(err).GetReason()))
	}

	if len(response.GetCards()) == 0 {
		err := errors.New("Empty List of cards returned from Entitlements")
		logf.Err(ctx, err)
		return nil, anzerrors.New(codes.NotFound, "Get Cards Failed", anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, "User has no cards"))
	}

	return response.GetCards(), nil
}

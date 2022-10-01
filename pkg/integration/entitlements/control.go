package entitlements

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/identity"
	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
)

func (c Client) Register(ctx context.Context, tokenizedCardNumber string) error {
	id, err := identity.Get(ctx)
	if err != nil {
		return err
	}

	req := &entpb.RegisterCardToPersonaRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		ExternalPersonaId:   id.PersonaID,
	}

	_, err = c.EntitlementsControlAPIClient.RegisterCardToPersona(ctx, req)
	if err != nil {
		return anzerrors.Wrap(err, anzerrors.GetStatusCode(err), "Entitlements/RegisterCardToPersona failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, anzerrors.GetErrorInfo(err).GetReason()))
	}

	return nil
}

func (c Client) Latest(ctx context.Context) error {
	id, err := identity.Get(ctx)
	if err != nil {
		return err
	}

	req := &entpb.ForcePartyToLatestRequest{ExternalPartyId: id.OcvID}

	_, err = c.EntitlementsControlAPIClient.ForcePartyToLatest(ctx, req)
	if err != nil {
		return anzerrors.Wrap(err, anzerrors.GetStatusCode(err), "Entitlements/ForcePartyToLatestRequest failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, anzerrors.GetErrorInfo(err).GetReason()))
	}

	return nil
}

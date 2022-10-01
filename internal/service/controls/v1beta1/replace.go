package v1beta1

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzcodes "github.com/anzx/pkg/errors/errcodes"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

const replaceFailed = "replace failed"

func (s server) Replace(ctx context.Context, req *ccpb.ReplaceRequest) (*ccpb.ReplaceResponse, error) {
	operations := []string{entitlements.OPERATION_MANAGE_CARD, entitlements.OPERATION_CARDCONTROLS}
	if _, err := s.Entitlements.GetEntitledCard(ctx, req.NewTokenizedCardNumber, operations...); err != nil {
		return nil, serviceErr(err, replaceFailed)
	}

	if _, err := s.Entitlements.GetEntitledCard(ctx, req.CurrentTokenizedCardNumber, operations...); err != nil {
		return nil, serviceErr(err, replaceFailed)
	}

	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_CARD_CONTROLS, req.NewTokenizedCardNumber); err != nil {
		return nil, serviceErr(err, replaceFailed)
	}

	decodedCardNumbers, err := s.Vault.DecodeCardNumbers(ctx, []string{req.CurrentTokenizedCardNumber, req.NewTokenizedCardNumber})
	if err != nil {
		logf.Err(ctx, err)
		return nil, serviceErr(err, replaceFailed)
	}

	ok, err := s.Visa.ReplaceCard(ctx, decodedCardNumbers[req.CurrentTokenizedCardNumber], decodedCardNumbers[req.NewTokenizedCardNumber])
	if err != nil {
		return nil, serviceErr(err, replaceFailed)
	}
	if !ok {
		return nil, anzerrors.New(codes.Aborted, replaceFailed, anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to replace card"))
	}
	return &ccpb.ReplaceResponse{Status: true}, nil
}

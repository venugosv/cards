package v1beta1

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

const serviceFailed = "service failed"

func (s server) Query(ctx context.Context, req *ccpb.QueryRequest) (*ccpb.CardControlResponse, error) {
	if _, err := s.Entitlements.GetEntitledCard(ctx, req.GetTokenizedCardNumber(), entitlements.OPERATION_CARDCONTROLS); err != nil {
		return nil, serviceErr(err, "query failed")
	}

	_, controlDocument, err := s.getControlDocument(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, "query failed")
	}

	return getCardControlResponse(controlDocument), nil
}

func (s server) getControlDocument(ctx context.Context, tokenizedCardNumber string) (*string, *visa.Resource, error) {
	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_CARD_CONTROLS, tokenizedCardNumber); err != nil {
		return nil, nil, serviceErr(err, serviceFailed)
	}

	cardNumber, err := s.Vault.DecodeCardNumber(ctx, tokenizedCardNumber)
	if err != nil {
		return nil, nil, serviceErr(err, serviceFailed)
	}
	// query controls by pan
	visaResponse, err := s.Visa.QueryControls(ctx, cardNumber)
	if err != nil {
		return nil, nil, serviceErr(err, serviceFailed)
	}

	// get Control Document
	if visaResponse == nil {
		return nil, nil, anzerrors.New(codes.NotFound, serviceFailed, anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "no control document found"))
	}

	return &cardNumber, visaResponse, nil
}

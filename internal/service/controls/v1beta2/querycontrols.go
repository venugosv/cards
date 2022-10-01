package v1beta2

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/pkg/feature"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
)

const serviceFailed = "service failed"

func (s server) QueryControls(ctx context.Context, req *ccpb.QueryControlsRequest) (*ccpb.CardControlResponse, error) {
	var (
		err     error
		visaCtx context.Context
	)
	if _, err := s.Entitlements.GetEntitledCard(ctx, req.GetTokenizedCardNumber(), entitlements.OPERATION_CARDCONTROLS); err != nil {
		return nil, serviceErr(err, "query failed")
	}

	if feature.FeatureGate.Enabled(feature.FORGEROCK_SYSTEM_LOGIN) {
		visaCtx, err = s.Forgerock.SystemJWT(ctx, visaGatewayRead)
		if err != nil {
			return nil, serviceErr(err, "query failed")
		}
	} else {
		visaCtx = ctx
	}

	_, controlDocument, err := s.getControlDocument(ctx, visaCtx, req.TokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, "query failed")
	}

	return getCardControlResponse(controlDocument, req.TokenizedCardNumber), nil
}

func (s server) getControlDocument(ctx context.Context, visaCtx context.Context, tokenizedCardNumber string) (*string, *crpb.Resource, error) {
	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_CARD_CONTROLS, tokenizedCardNumber); err != nil {
		return nil, nil, serviceErr(err, serviceFailed)
	}

	cardNumber, err := s.Vault.DecodeCardNumber(ctx, tokenizedCardNumber)
	if err != nil {
		return nil, nil, serviceErr(err, serviceFailed)
	}
	// query controls by pan
	visaResponse, err := s.Visa.ListControlDocuments(visaCtx, cardNumber)
	if err != nil {
		return nil, nil, serviceErr(err, serviceFailed)
	}

	// get Control Document
	if visaResponse == nil {
		return nil, nil, anzerrors.New(codes.NotFound, serviceFailed, anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "no control document found"))
	}

	return &cardNumber, visaResponse, nil
}

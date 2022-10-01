package v1beta2

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/xcontext"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
)

const replaceFailed = "replace failed"

func (s server) TransferControls(ctx context.Context, req *ccpb.TransferControlsRequest) (*ccpb.TransferControlsResponse, error) {
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

	var visaCtx context.Context
	if feature.FeatureGate.Enabled(feature.FORGEROCK_SYSTEM_LOGIN) {
		visaCtx, err = s.Forgerock.SystemJWT(ctx, visaGatewayUpdate)
		if err != nil {
			return nil, serviceErr(err, listFailed)
		}
	} else {
		visaCtx = ctx
	}

	ok, err := s.Visa.Replace(visaCtx, decodedCardNumbers[req.CurrentTokenizedCardNumber], decodedCardNumbers[req.NewTokenizedCardNumber])
	if err != nil {
		return nil, serviceErr(err, replaceFailed)
	}
	if !ok {
		return nil, anzerrors.New(codes.Aborted, replaceFailed, anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to replace card"))
	}

	s.CommandCentre.PublishEventAsync(ctx, event.CardControlsChange)

	// Remove global control as part of the transferControls
	detachCtx := xcontext.Detach(visaCtx)
	go func() {
		removeReq := &ccpb.RemoveControlsRequest{
			TokenizedCardNumber: req.GetNewTokenizedCardNumber(),
			ControlTypes:        []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
		}
		if _, err := s.RemoveControls(detachCtx, removeReq); err != nil {
			logf.Error(detachCtx, err, "failed to remove global control after replace")
		}
	}()

	return &ccpb.TransferControlsResponse{}, nil
}

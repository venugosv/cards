package cards

import (
	"context"
	"fmt"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/date"
	types "github.com/anzx/fabricapis/pkg/fabric/type"

	"github.com/anzx/fabric-cards/pkg/feature"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"google.golang.org/grpc/codes"

	"github.com/anzx/pkg/errors"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

const (
	getDetailsFailed = "get details failed"
	dcvv2Create      = "https://fabric.anz.com/scopes/visaGateway:dcvv2:create"
)

func (s server) GetDetails(ctx context.Context, req *cpb.GetDetailsRequest) (*cpb.GetDetailsResponse, error) {
	card, err := s.getCard(ctx, req.TokenizedCardNumber, "get details")
	if err != nil {
		return nil, err
	}

	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_GET_DETAILS, req.TokenizedCardNumber); err != nil {
		return nil, anzerrors.Wrap(err, codes.PermissionDenied, getDetailsFailed, anzerrors.GetErrorInfo(err))
	}

	cardNumber, err := s.Vault.DecodeCardNumber(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, errors.Wrap(err, codes.Internal, getDetailsFailed, errors.NewErrorInfo(ctx, anzcodes.CardTokenizationFailed, serviceUnavailable))
	}

	expiry := date.GetDate(ctx, date.YYMM, card.ExpiryDate)

	dcvv2 := s.getDCVV2(ctx, cardNumber, expiry)

	return &cpb.GetDetailsResponse{
		Name:          fmt.Sprintf("%s %s", card.FirstName, card.LastName),
		CardNumber:    cardNumber,
		ExpiryDate:    date.GetDate(ctx, date.YYMM, card.ExpiryDate),
		Cvc:           dcvv2,
		Eligibilities: card.Eligibility(),
	}, nil
}

func (s server) getDCVV2(ctx context.Context, cardNumber string, expiry *types.Date) string {
	if !feature.FeatureGate.Enabled(feature.DCVV2) {
		return ""
	}
	if feature.FeatureGate.Enabled(feature.FORGEROCK_SYSTEM_LOGIN) {
		visaCtx, err := s.Forgerock.SystemJWT(ctx, dcvv2Create)
		if err != nil {
			logf.Error(ctx, err, "unable to log into forgerock")
			return ""
		}
		ctx = visaCtx
	}

	dcvv2Item, err := s.DCVV2.Generate(ctx, date.ProtoToDate(expiry).String(), cardNumber)
	if err != nil {
		logf.Error(ctx, err, "unable to generate DCVV2")
		return ""
	}
	return dcvv2Item.GetDcvv2Value()
}

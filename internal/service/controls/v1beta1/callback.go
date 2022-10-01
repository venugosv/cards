package v1beta1

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
)

func (s server) Enrol(ctx context.Context, req *ccpb.CallbackRequest) (*ccpb.CallbackResponse, error) {
	return s.Flag(ctx, req, s.CTM, true)
}

func (s server) Disenrol(ctx context.Context, req *ccpb.CallbackRequest) (*ccpb.CallbackResponse, error) {
	return s.Flag(ctx, req, s.CTM, false)
}

func (s server) Flag(ctx context.Context, req *ccpb.CallbackRequest, c ctm.CardMaintenanceAPI, flag bool) (*ccpb.CallbackResponse, error) {
	if len(req.BulkEnrollmentObjectList) == 0 {
		return nil, anzerrors.New(codes.InvalidArgument, "callback failed", anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "empty BulkEnrollmentObjectList provided"))
	}

	for _, bulkEnrollmentObject := range req.BulkEnrollmentObjectList {
		tokenizedCardNumber, err := s.Vault.EncodeCardNumber(ctx, bulkEnrollmentObject.CardNumber)
		if err != nil {
			return nil, anzerrors.Wrap(err, codes.Unavailable, "callback failed", anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "service unavailable"))
		}

		preference := &ctm.UpdatePreferencesRequest{
			CardControlPreference: &flag,
		}

		ok, err := c.UpdatePreferences(ctx, preference, tokenizedCardNumber)
		if !ok || err != nil {
			return nil, anzerrors.Wrap(err, codes.Unavailable, "callback failed", anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "service unavailable"))
		}

		logf.Info(ctx, "successfully set flag for %v", tokenizedCardNumber)
	}

	return &ccpb.CallbackResponse{}, nil
}

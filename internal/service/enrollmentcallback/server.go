package enrollmentcallback

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/fakerock"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"

	"github.com/anzx/fabric-cards/pkg/integration/vault"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	ecpb "github.com/anzx/fabricapis/pkg/visa/service/enrollmentcallback"
)

const (
	callbackFailed = "callback failed"
	updateScope    = "AU.RETAIL.DEBITCARDS.UPDATE"
)

type server struct {
	ecpb.UnimplementedEnrollmentCallbackAPIServer
	ctm       ctm.CardMaintenanceAPI
	vault     vault.Client
	fakerock  *fakerock.Client
	forgerock forgerock.Clienter
}

// NewServer constructs a new CustomerRulesAPI from configured clients
func NewServer(ctm ctm.ControlAPI, vault vault.Client, fakerock *fakerock.Client, forgerock forgerock.Clienter) ecpb.EnrollmentCallbackAPIServer {
	return &server{
		ctm:       ctm,
		vault:     vault,
		fakerock:  fakerock,
		forgerock: forgerock,
	}
}

func (s server) Enroll(ctx context.Context, request *ecpb.Request) (*ecpb.Response, error) {
	if !feature.FeatureGate.Enabled(feature.ENROLLMENT_CALLBACK_INTEGRATED) {
		logf.Info(ctx, "enrollment not integrated")
		return &ecpb.Response{}, nil
	}
	return s.flag(ctx, request, true)
}

func (s server) Disenroll(ctx context.Context, request *ecpb.Request) (*ecpb.Response, error) {
	if !feature.FeatureGate.Enabled(feature.ENROLLMENT_CALLBACK_INTEGRATED) {
		logf.Info(ctx, "disenrollment not integrated")
		return &ecpb.Response{}, nil
	}
	return s.flag(ctx, request, false)
}

func (s server) flag(ctx context.Context, request *ecpb.Request, flag bool) (*ecpb.Response, error) {
	size := len(request.GetBulkEnrollmentObjectList())
	if size == 0 {
		return nil, anzerrors.New(codes.InvalidArgument, callbackFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "empty BulkEnrollmentObjectList provided"))
	}

	pans := make([]string, 0, size)
	for _, bulkEnrollmentObject := range request.BulkEnrollmentObjectList {
		pans = append(pans, bulkEnrollmentObject.GetPrimaryAccountNumber())
	}

	tokenizedCardNumbers, err := s.vault.EncodeCardNumbers(ctx, pans)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, callbackFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "service unavailable"))
	}

	if feature.FeatureGate.Enabled(feature.FORGEROCK_SYSTEM_LOGIN) {
		ctx, err = s.forgerock.SystemJWT(ctx, updateScope)
		if err != nil {
			return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), callbackFailed, anzerrors.GetErrorInfo(err))
		}
	} else {
		ctx, err = s.fakerock.ElevateContext(ctx)
		if err != nil {
			return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), callbackFailed, anzerrors.GetErrorInfo(err))
		}
	}

	for _, tokenizedCardNumber := range tokenizedCardNumbers {
		preference := &ctm.UpdatePreferencesRequest{
			CardControlPreference: &flag,
		}

		ok, err := s.ctm.UpdatePreferences(ctx, preference, tokenizedCardNumber)
		if !ok || err != nil {
			return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), callbackFailed, anzerrors.GetErrorInfo(err))
		}

		logf.Info(ctx, "successfully set flag for %v", tokenizedCardNumber)
	}

	return &ecpb.Response{}, nil
}

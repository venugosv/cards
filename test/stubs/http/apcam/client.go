package apcam

import (
	"context"
	"regexp"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/apcam"
)

type StubClient struct {
	Err error
}

// NewStubClient creates a APCAMClient client stubs
func NewStubClient() StubClient {
	return StubClient{}
}

func (e StubClient) PushProvision(ctx context.Context, in apcam.Request) (*apcam.Response, error) {
	if e.Err != nil {
		return nil, e.Err
	}

	if !regexp.MustCompile(YYYYMM).MatchString(in.CardInfo.ExpiryDate) {
		return nil, anzerrors.New(codes.InvalidArgument, "Push Provision failed", anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "invalid expiry"))
	}

	return getResponse(&in), nil
}

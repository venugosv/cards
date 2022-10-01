package v1beta2

import (
	"context"
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/assert"
)

const token = "token"

func TestTransferControls(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, personaID string
		builder         *fixtures.ServerBuilder
		req             *ccpb.TransferControlsRequest
		want            *ccpb.TransferControlsResponse
		wantErr         error
	}{
		{
			name:    "Successful Card Replace Response",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			want: &ccpb.TransferControlsResponse{},
		},
		{
			name:    "unable to call Entitlements",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=service unavailable"),
		},
		{
			name:    "unable to verify ownership of current card",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.RandomUser().Token(),
				NewTokenizedCardNumber:     data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=replace failed, reason=user not entitled"),
		},
		{
			name:    "unable to verify ownership of new card",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     data.RandomUser().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=replace failed, reason=user not entitled"),
		},
		{
			name:    "unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen)).AddACard(data.WithAToken(token), data.WithStatus(ctm.StatusTemporaryBlock))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=replace failed, reason=card not eligible"),
		},
		{
			name:    "unable to replace card with customerRulesAPI error",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen)).AddACard(data.WithAToken(token))).WithVisaGatewayReplaceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "unable to replace card without customerRulesAPI error",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithVisaGatewayReplaceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "unable to tokenize card",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=service unavailable"),
		},
		{
			name:    "unable to replace card with Visa",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithVisaGatewayReplaceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "unable to remove global control",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithVisaGatewayDeleteError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.TransferControlsRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			want: &ccpb.TransferControlsResponse{},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardControlsServer(test.builder)

			ctx := fixtures.GetTestContext()
			if test.personaID != "" {
				ctx = fixtures.GetTestContextWithJWT(test.personaID)
			}

			got, err := s.TransferControls(ctx, test.req)
			if test.wantErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, test.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

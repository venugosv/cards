package v1beta1

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/test/data"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const token = "token"

func TestReplace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, personaID string
		builder         *fixtures.ServerBuilder
		req             *ccpb.ReplaceRequest
		want            *ccpb.ReplaceResponse
		wantErr         error
	}{
		{
			name:    "Successful Card Replace Response",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			want: &ccpb.ReplaceResponse{
				Status: true,
			},
		},
		{
			name:    "unable to call Entitlements",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=service unavailable"),
		},
		{
			name:    "unable to verify ownership of current card",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.RandomUser().Token(),
				NewTokenizedCardNumber:     data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=replace failed, reason=user not entitled"),
		},
		{
			name:    "unable to verify ownership of new card",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     data.RandomUser().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=replace failed, reason=user not entitled"),
		},
		{
			name:    "unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen)).AddACard(data.WithAToken(token), data.WithStatus(ctm.StatusTemporaryBlock))),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=replace failed, reason=card not eligible"),
		},
		{
			name:    "unable to replace card with customerRulesAPI error",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen)).AddACard(data.WithAToken(token))).WithVtcReplaceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=service unavailable"),
		},
		{
			name:    "unable to replace card without customerRulesAPI error",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithVtcReplaceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=service unavailable"),
		},
		{
			name:    "unable to tokenize card",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=service unavailable"),
		},
		{
			name:    "unable to replace card with Visa",
			builder: fixtures.AServer().WithData(data.AUserWithACard().AddACard(data.WithAToken(token))).WithVtcReplaceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.ReplaceRequest{
				CurrentTokenizedCardNumber: data.AUserWithACard().Token(),
				NewTokenizedCardNumber:     token,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=replace failed, reason=service unavailable"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildCardControlsServer(tt.builder)

			ctx := fixtures.GetTestContext()
			if tt.personaID != "" {
				ctx = fixtures.GetTestContextWithJWT(tt.personaID)
			}

			got, err := s.Replace(ctx, tt.req)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

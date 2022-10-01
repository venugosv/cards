package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/test/data"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"

	"github.com/stretchr/testify/assert"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

func TestQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.QueryRequest
		want    *ccpb.CardControlResponse
		wantErr error
	}{
		{
			name:    "successful call to query controls",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))),
			req: &ccpb.QueryRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			want: &ccpb.CardControlResponse{
				CardControls: []*ccpb.CardControl{
					{
						ControlType:    ccpb.ControlType_GCT_GLOBAL,
						ControlEnabled: true,
					},
				},
			},
		},
		{
			name:    "Invalid ENT call",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.QueryRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=query failed, reason=service unavailable"),
		},
		{
			name:    "Invalid VTC call",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVtcQueryError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.QueryRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=query failed, reason=service unavailable"),
		},
		{
			name:    "Unable to verify ownership",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))),
			req: &ccpb.QueryRequest{
				TokenizedCardNumber: data.RandomUser().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=query failed, reason=user not entitled"),
		},
		{
			name:    "Unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))),
			req: &ccpb.QueryRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=query failed, reason=card not eligible"),
		},
		{
			name:    "Unable to tokenize card",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.QueryRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=query failed, reason=service unavailable"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildCardControlsServer(tt.builder)
			got, err := s.Query(fixtures.GetTestContext(), tt.req)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

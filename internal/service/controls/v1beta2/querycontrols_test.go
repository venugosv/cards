package v1beta2

import (
	"context"
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/test/data"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"

	"github.com/stretchr/testify/assert"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
)

func TestServer_QueryControls(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.QueryControlsRequest
		want    *ccpb.CardControlResponse
		wantErr error
	}{
		{
			name:    "successful call to query controls",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))),
			req: &ccpb.QueryControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			want: &ccpb.CardControlResponse{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				CardControls: []*ccpb.CardControl{
					{
						ControlType: ccpb.ControlType_GCT_GLOBAL,
					},
				},
			},
		},
		{
			name:    "Invalid ENT call",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.QueryControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=query failed, reason=service unavailable"),
		},
		{
			name:    "Invalid VisaGateway call",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVisaGatewayListError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.QueryControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=query failed, reason=invalid response from visa gateway"),
		},
		{
			name:    "Unable to verify ownership",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))),
			req: &ccpb.QueryControlsRequest{
				TokenizedCardNumber: data.RandomUser().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=2, message=query failed, reason=user not entitled"),
		},
		{
			name:    "Unable to verify Eligibility",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))),
			req: &ccpb.QueryControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=query failed, reason=card not eligible"),
		},
		{
			name:    "Unable to tokenize card",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithControls(data.CardControlsPresetGlobalControls))).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.QueryControlsRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=query failed, reason=service unavailable"),
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardControlsServer(test.builder)
			got, err := s.QueryControls(fixtures.GetTestContext(), test.req)
			if test.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestServer_QueryControlsMultipleUsers(t *testing.T) {
	var resource crpb.Resource
	require.NoError(t, protojson.Unmarshal(sampleControlDoc, &resource))

	sb := fixtures.AServer().WithData(data.AUserWithACard()).WithVisaGatewayResource(&resource)
	s := buildCardControlsServer(sb)

	request := &ccpb.QueryControlsRequest{
		TokenizedCardNumber: data.AUserWithACard().Token(),
	}

	got, err := s.QueryControls(fixtures.GetTestContext(), request)
	require.NoError(t, err)

	require.Len(t, got.GetCardControls(), 3)
}

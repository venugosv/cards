package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/test/util"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"

	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

func TestBlock(t *testing.T) {
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.BlockRequest
		want    *ccpb.BlockResponse
		wantErr error
	}{
		{
			name:    "Successfully apply block",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_BLOCK,
			},
			want: &ccpb.BlockResponse{
				Status: true,
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_UNBLOCK,
				},
			},
		},
		{
			name:    "audit log does not affect response",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_BLOCK,
			},
			want: &ccpb.BlockResponse{
				Status: true,
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_UNBLOCK,
				},
			},
		},
		{
			name:    "Entitlements call returns not entitled",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_BLOCK,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=block failed, reason=service unavailable"),
		},
		{
			name:    "Eligibility call returns not eligible",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_BLOCK,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=block failed, reason=card not eligible"),
		},
		{
			name:    "CTM call fails to update status",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmUpdateError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_BLOCK,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=block failed, reason=service unavailable"),
		},
		{
			name:    "CTM inquiry call failed",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmInquiryError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_BLOCK,
			},
			want: &ccpb.BlockResponse{
				Status: true,
			},
		},
		{
			name:    "eligibility failed, unblock return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_UNBLOCK,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=unblock failed, reason=card not eligible"),
		},
		{
			name:    "Eligibility call returns not eligible due to card is not yet activated",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.Inactive, data.WithStatus(ctm.StatusIssued))),
			req: &ccpb.BlockRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockRequest_ACTION_BLOCK,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=block failed, reason=card not eligible"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardControlsServer(test.builder)
			got, err := s.Block(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func TestBlockAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.BlockCard{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}

		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogHook(hook)
		request := &ccpb.BlockRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			Action:              ccpb.BlockRequest_ACTION_BLOCK,
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		_, _ = s.Block(ctx, request)
		require.NoError(t, sd.Validate())
		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

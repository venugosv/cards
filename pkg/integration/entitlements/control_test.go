package entitlements

import (
	"context"
	"testing"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/test/fixtures"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"

	"google.golang.org/grpc"
)

type mockEntitlementsControlClient struct {
	entpb.EntitlementsControlAPIClient
	RegisterCardToPersonaCall func(ctx context.Context, in *entpb.RegisterCardToPersonaRequest, opts ...grpc.CallOption) (*entpb.RegisterCardToPersonaResponse, error)
	forcePartyToLatestCall    func(ctx context.Context, in *entpb.ForcePartyToLatestRequest, opts ...grpc.CallOption) (*entpb.ForcePartyToLatestResponse, error)
}

func (m mockEntitlementsControlClient) RegisterCardToPersona(ctx context.Context, in *entpb.RegisterCardToPersonaRequest, opts ...grpc.CallOption) (*entpb.RegisterCardToPersonaResponse, error) {
	return m.RegisterCardToPersonaCall(ctx, in, opts...)
}

func (m mockEntitlementsControlClient) ForcePartyToLatest(ctx context.Context, in *entpb.ForcePartyToLatestRequest, opts ...grpc.CallOption) (*entpb.ForcePartyToLatestResponse, error) {
	return m.forcePartyToLatestCall(ctx, in, opts...)
}

func TestClient_Register(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		ent  entpb.EntitlementsControlAPIClient
		want string
	}{
		{
			name: "happy path",
			ctx:  fixtures.GetTestContext(),
			ent: &mockEntitlementsControlClient{
				RegisterCardToPersonaCall: func(ctx context.Context, in *entpb.RegisterCardToPersonaRequest, opts ...grpc.CallOption) (*entpb.RegisterCardToPersonaResponse, error) {
					return &entpb.RegisterCardToPersonaResponse{}, nil
				},
			},
		}, {
			name: "unable to get persona",
			ctx:  context.Background(),
			want: "fabric error: status_code=Internal, error_code=5, message=identity error, reason=could not retrieve user identification",
		}, {
			name: "entitlements failed",
			ctx:  fixtures.GetTestContext(),
			ent: &mockEntitlementsControlClient{
				RegisterCardToPersonaCall: func(ctx context.Context, in *entpb.RegisterCardToPersonaRequest, opts ...grpc.CallOption) (*entpb.RegisterCardToPersonaResponse, error) {
					return nil, anzerrors.New(codes.NotFound, "entitlements failed", anzerrors.NewErrorInfo(context.Background(), anzcodes.Unknown, "cannot find persona"))
				},
			},
			want: "fabric error: status_code=NotFound, error_code=2, message=Entitlements/RegisterCardToPersona failed, reason=cannot find persona",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Client{EntitlementsControlAPIClient: test.ent}
			err := c.Register(test.ctx, tokenizedCardNumber)
			if test.want == "" {
				require.NoError(t, err)
			} else {
				assert.Equal(t, test.want, err.Error())
			}
		})
	}
}

func TestClient_Latest(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		ent  entpb.EntitlementsControlAPIClient
		want string
	}{
		{
			name: "happy path",
			ctx:  fixtures.GetTestContext(),
			ent: &mockEntitlementsControlClient{
				forcePartyToLatestCall: func(ctx context.Context, in *entpb.ForcePartyToLatestRequest, opts ...grpc.CallOption) (*entpb.ForcePartyToLatestResponse, error) {
					return &entpb.ForcePartyToLatestResponse{}, nil
				},
			},
		}, {
			name: "unable to get persona",
			ctx:  context.Background(),
			want: "fabric error: status_code=Internal, error_code=5, message=identity error, reason=could not retrieve user identification",
		}, {
			name: "entitlements failed",
			ctx:  fixtures.GetTestContext(),
			ent: &mockEntitlementsControlClient{
				forcePartyToLatestCall: func(ctx context.Context, in *entpb.ForcePartyToLatestRequest, opts ...grpc.CallOption) (*entpb.ForcePartyToLatestResponse, error) {
					return nil, anzerrors.New(codes.NotFound, "entitlements failed", anzerrors.NewErrorInfo(context.Background(), anzcodes.Unknown, "cannot find persona"))
				},
			},
			want: "fabric error: status_code=NotFound, error_code=2, message=Entitlements/ForcePartyToLatestRequest failed, reason=cannot find persona",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Client{EntitlementsControlAPIClient: test.ent}
			err := c.Latest(test.ctx)
			if test.want == "" {
				require.NoError(t, err)
			} else {
				assert.Equal(t, test.want, err.Error())
			}
		})
	}
}

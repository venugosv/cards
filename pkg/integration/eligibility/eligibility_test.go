package eligibility

import (
	"context"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"github.com/anzx/pkg/jwtauth"
	"google.golang.org/grpc/codes"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/stretchr/testify/require"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockEligibilityClient struct {
	canCall func(ctx context.Context, in *epb.CanRequest, opts ...grpc.CallOption) (*epb.CanResponse, error)
}

func (m *mockEligibilityClient) Can(ctx context.Context, in *epb.CanRequest, opts ...grpc.CallOption) (*epb.CanResponse, error) {
	return m.canCall(ctx, in, opts...)
}

func TestNewEligibilityClient(t *testing.T) {
	tests := []struct {
		name        string
		input       *Config
		errContents []string
	}{
		{
			name: "Valid config",
			input: &Config{
				BaseURL: "localhost:9090",
			},
		},
		{
			name: "nil config",
		},
		{
			name:        "Invalid config",
			input:       &Config{BaseURL: "%%"},
			errContents: []string{"status_code=Internal", "fabric error", "error_code=1", "failed to create eligibility adapter", "unable to parse configured url"},
		},
	}
	for _, test := range tests {
		tt := test
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(context.Background(), tt.input, opts...)
			if test.errContents != nil {
				require.Nil(t, got)
				require.Error(t, err)
				for _, val := range tt.errContents {
					assert.Contains(t, err.Error(), val)
				}
			} else {
				require.NoError(t, err)
				if test.input == nil {
					assert.Nil(t, got)
				} else {
					require.NotNil(t, got)
				}
			}
		})
	}
}

func TestCan(t *testing.T) {
	ctx := context.Background()

	ctxWithClaims := jwtauth.AddClaimsToContext(ctx, jwtauth.NewClaims(jwtauth.BaseClaims{
		Claims: jwt.Claims{
			Subject: "fake subject UUID",
		},
	}))

	for _, tt := range []struct {
		name        string
		context     context.Context
		err         error
		response    *epb.CanResponse
		expectError func(*testing.T, error)
	}{{
		name:     "basic happy",
		context:  ctxWithClaims,
		response: &epb.CanResponse{},
	}, {
		name:    "Not Eligible",
		context: ctxWithClaims,
		err:     anzerrors.New(codes.PermissionDenied, "Ineligible", anzerrors.NewErrorInfo(context.Background(), anzcodes.CardIneligible, "card not eligible")),
		expectError: func(t *testing.T, err error) {
			assert.Contains(t, err.Error(), "fabric error: status_code=Internal, error_code=0, message=eligibility failed, reason=card not eligible")
		},
	}} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ent := &mockEligibilityClient{}

			ent.canCall = func(ctx context.Context, req *epb.CanRequest, opts ...grpc.CallOption) (*epb.CanResponse, error) {
				operations := req.GetEligibility()
				assert.Equal(t, epb.Eligibility_ELIGIBILITY_APPLE_PAY, operations)
				assert.Equal(t, "1234567890123456", req.GetTokenizedCardNumber())
				return test.response, test.err
			}

			c := &Client{CardEligibilityAPIClient: ent}
			err := c.Can(test.context, epb.Eligibility_ELIGIBILITY_APPLE_PAY, "1234567890123456")
			if test.expectError != nil {
				require.Error(t, err)
				assert.Equal(t, test.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

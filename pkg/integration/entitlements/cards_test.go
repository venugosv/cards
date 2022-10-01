package entitlements

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
	"github.com/anzx/pkg/jwtauth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"gopkg.in/square/go-jose.v2/jwt"
)

const tokenizedCardNumber = "token"

func TestClient_Get(t *testing.T) {
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
		response    *entpb.EntitledCard
		expectError func(*testing.T, error)
	}{
		{
			name:    "basic happy",
			context: ctxWithClaims,
			response: &entpb.EntitledCard{
				TokenizedCardNumber: tokenizedCardNumber,
			},
		},
		{
			name:    "Not Entitled",
			context: ctxWithClaims,
			response: &entpb.EntitledCard{
				TokenizedCardNumber: tokenizedCardNumber,
			},
			err: anzerrors.New(codes.Internal, "big bad fail", anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "some bad reason")),
			expectError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "fabric error: status_code=Internal, error_code=2, message=Entitlements/GetEntitledCard failed, reason=some bad reason")
			},
		},
		{
			name:    "Entitlements Error",
			context: ctxWithClaims,
			err:     errors.New("SensitiveInfo"),
			expectError: func(t *testing.T, err error) {
				assert.NotContains(t, err.Error(), "SensitiveInfo")
			},
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ent := &mockEntitlementsClient{}
			c := &Client{CardEntitlementsAPIClient: ent}

			ent.getCall = func(ctx context.Context, req *entpb.GetEntitledCardRequest, opts ...grpc.CallOption) (*entpb.EntitledCard, error) {
				assert.Equal(t, tokenizedCardNumber, req.GetTokenizedCardNumber())
				return test.response, test.err
			}

			_, err := c.GetEntitledCard(test.context, tokenizedCardNumber, OPERATION_VIEW_CARD)
			if test.expectError != nil {
				if err == nil {
					t.Fatal("Expected error, got none")
				} else {
					test.expectError(t, err)
				}
			} else {
				if err != nil {
					t.Fatal(err.Error())
				}
			}
		})
	}
}

func TestGetUserCardNumbers(t *testing.T) {
	tests := []struct {
		name string
		ent  entpb.CardEntitlementsAPIClient
		want []*entpb.EntitledCard
	}{
		{
			name: "SuccessfullyListAccountNumbers",
			ent: &mockEntitlementsClient{
				listCall: func(ctx context.Context, in *entpb.ListEntitledCardsRequest, opts ...grpc.CallOption) (*entpb.ListEntitledCardsResponse, error) {
					return &entpb.ListEntitledCardsResponse{
						Cards: []*entpb.EntitledCard{
							{
								TokenizedCardNumber: tokenizedCardNumber,
								AccountNumbers: []string{
									"accountNumber1",
									"accountNumber2",
								},
							},
						},
					}, nil
				},
			},
			want: []*entpb.EntitledCard{
				{
					TokenizedCardNumber: tokenizedCardNumber,
					AccountNumbers: []string{
						"accountNumber1",
						"accountNumber2",
					},
				},
			},
		},
	}

	ctxWithClaims := jwtauth.AddClaimsToContext(context.Background(), jwtauth.NewClaims(jwtauth.BaseClaims{
		Claims: jwt.Claims{
			Subject: "fake subject UUID",
		},
	}))

	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{CardEntitlementsAPIClient: test.ent}
			got, err := c.ListEntitledCards(ctxWithClaims)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetUserCardNumbersError(t *testing.T) {
	tests := []struct {
		name string
		ent  entpb.CardEntitlementsAPIClient
		want error
	}{
		{
			name: "failed to make ListEntitledCards Me request",
			ent: &mockEntitlementsClient{
				listCall: func(ctx context.Context, in *entpb.ListEntitledCardsRequest, opts ...grpc.CallOption) (*entpb.ListEntitledCardsResponse, error) {
					return nil, anzerrors.New(codes.Internal, "big bad fail", anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "some bad reason"))
				},
			},
			want: fmt.Errorf("fabric error: status_code=Internal, error_code=2, message=Entitlements ListEntitledCards failed, reason=some bad reason"),
		},
		{
			name: "empty ListEntitledCards Me response",
			ent: &mockEntitlementsClient{
				listCall: func(ctx context.Context, in *entpb.ListEntitledCardsRequest, opts ...grpc.CallOption) (*entpb.ListEntitledCardsResponse, error) {
					return &entpb.ListEntitledCardsResponse{
						Cards: []*entpb.EntitledCard{},
					}, nil
				},
			},
			want: fmt.Errorf("fabric error: status_code=NotFound, error_code=20000, message=Get Cards Failed, reason=User has no cards"),
		},
	}

	ctxWithClaims := jwtauth.AddClaimsToContext(context.Background(), jwtauth.NewClaims(jwtauth.BaseClaims{
		Claims: jwt.Claims{
			Subject: "fake subject UUID",
		},
	}))

	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{CardEntitlementsAPIClient: test.ent}
			got, err := c.ListEntitledCards(ctxWithClaims)
			require.Nil(t, got)
			assert.Equal(t, tt.want.Error(), err.Error())
		})
	}
}

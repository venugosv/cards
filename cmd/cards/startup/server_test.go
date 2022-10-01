package startup

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/anzx/fabric-cards/internal/service/wallet"

	"google.golang.org/grpc/credentials/insecure"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/jwtauth/jwtgrpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/anzx/fabric-cards/cmd/cards/config/app"
	"github.com/anzx/fabric-cards/internal/service/cards"
	"github.com/anzx/fabric-cards/internal/service/eligibility"
	"github.com/anzx/pkg/jwtauth"
	"github.com/anzx/pkg/jwtauth/jwttest"
	"github.com/stretchr/testify/assert"
)

const (
	cardsCreate            = "https://fabric.anz.com/scopes/cards:create"
	cardsRead              = "https://fabric.anz.com/scopes/cards:read"
	cardsUpdate            = "https://fabric.anz.com/scopes/cards:update"
	entitlementsRead       = "https://fabric.anz.com/scopes/entitlements:read"
	selfServiceProfileRead = "https://fabric.anz.com/scopes/selfservice:profile:read"
	pinInfoUpdate          = "AU.RETAIL.PININFO.UPDATE"
)

func TestRunAPIServer(t *testing.T) {
	t.Run("failed to start server but all checks pass", func(t *testing.T) {
		cfg := app.Spec{
			AppName: "test-app",
			Port:    65536, // highest possible port number + 1
			Auth:    jwtauth.Config{},
		}

		ctx := context.Background()

		decider := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return false }

		cardServer := cards.NewServer(cards.Fabric{}, cards.Internal{}, cards.External{})
		eligibilityServer := eligibility.NewServer(nil, nil, nil)
		walletServer := wallet.NewServer(nil, nil, nil, nil, nil, nil, nil, nil)

		assert.Error(t, RunAPIServer(ctx, cfg, decider, &jwtauth.InsecureAuthenticator{}, cardServer, eligibilityServer, walletServer)())
	})
}

func TestAuthRules(t *testing.T) {
	// Setup the grpc server
	issuer, err := jwttest.NewIssuer("test", 1024)
	require.NoError(t, err)
	svr := grpc.NewServer(
		grpc.ChainUnaryInterceptor(jwtgrpc.UnaryServerInterceptor(issuer)), // issuer doubles as its own jwtauth.Authenticator
	)

	// Register unimplemented services
	cpb.RegisterCardAPIServer(svr, &cpb.UnimplementedCardAPIServer{})
	epb.RegisterCardEligibilityAPIServer(svr, &epb.UnimplementedCardEligibilityAPIServer{})
	cpb.RegisterWalletAPIServer(svr, &cpb.UnimplementedWalletAPIServer{})

	// Set up a listener and serve
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() {
		if err := svr.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// Dial it. WithBlock causes it to wait for the goroutine to actually setup the server
	cc, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	require.NoError(t, err)
	c := cpb.NewCardAPIClient(cc)
	e := epb.NewCardEligibilityAPIClient(cc)
	w := cpb.NewWalletAPIClient(cc)

	// Tests
	t.Run("ActivateAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsUpdate, entitlementsRead)

		// Call
		_, err = c.Activate(ctx, &cpb.ActivateRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("AuditTrailAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsRead, entitlementsRead)

		// Call
		_, err = c.AuditTrail(ctx, &cpb.AuditTrailRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("ChangePINAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsUpdate, entitlementsRead)

		// Call
		_, err = c.ChangePIN(ctx, &cpb.ChangePINRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("GetDetailsAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsRead, entitlementsRead)

		// Call
		_, err = c.GetDetails(ctx, &cpb.GetDetailsRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("GetWrappingKeyAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsRead, entitlementsRead)

		// Call
		_, err = c.GetWrappingKey(ctx, &cpb.GetWrappingKeyRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("ListAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsRead, entitlementsRead)

		// Call
		_, err = c.List(ctx, &cpb.ListRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("ReplaceAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsUpdate, entitlementsRead, selfServiceProfileRead)

		// Call
		_, err = c.Replace(ctx, &cpb.ReplaceRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("ResetPINAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsUpdate, entitlementsRead, pinInfoUpdate)

		// Call
		_, err = c.ResetPIN(ctx, &cpb.ResetPINRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("SetPINAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsCreate, entitlementsRead, pinInfoUpdate)

		// Call
		_, err = c.SetPIN(ctx, &cpb.SetPINRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("VerifyPINAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsCreate, entitlementsRead)

		// Call
		_, err = c.VerifyPIN(ctx, &cpb.VerifyPINRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("CanAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardsCreate, entitlementsRead)

		// Call
		_, err = e.Can(ctx, &epb.CanRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("CreateApplePaymentTokenAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, "https://fabric.anz.com/scopes/cards:create", "https://fabric.anz.com/scopes/entitlements:read")

		// Call
		_, err = w.CreateApplePaymentToken(ctx, &cpb.CreateApplePaymentTokenRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
}

func getToken(t *testing.T, issuer jwttest.Issuer, scopes ...string) context.Context {
	// Create token
	token, err := issuer.Issue(jwtauth.BaseClaims{Scopes: scopes})
	require.NoError(t, err)

	// Create context
	return metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
}

// check that auth passed, since we use unimplemented servers, we auth passed means the error we get back is UnImplemented, not an authentication error
func assertAuthPassed(t *testing.T, err error) {
	ferr, ok := errors.FromStatusError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, ferr.GetStatusCode(), fmt.Sprintf("expected: %v, got: %v", codes.Unimplemented, ferr.GetStatusCode()))
}

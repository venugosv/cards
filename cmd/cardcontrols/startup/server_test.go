package startup

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/anzx/fabric-cards/internal/service/controls/v1beta1"

	v1beta2pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"

	"github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/jwtauth/jwtgrpc"
	"github.com/anzx/pkg/jwtauth/jwttest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/anzx/fabric-cards/pkg/servers"
	v1beta1pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/cmd/cardcontrols/config/app"
	"github.com/anzx/pkg/jwtauth"
	"github.com/stretchr/testify/assert"
)

const (
	cardControlsCreate = "https://fabric.anz.com/scopes/cardControls:create"
	cardControlsRead   = "https://fabric.anz.com/scopes/cardControls:read"
	cardControlsUpdate = "https://fabric.anz.com/scopes/cardControls:update"
	cardControlsDelete = "https://fabric.anz.com/scopes/cardControls:delete"
	entitlementsRead   = "https://fabric.anz.com/scopes/entitlements:read"
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

		cardControlsServer := v1beta1.NewServer(v1beta1.Fabric{}, v1beta1.Internal{}, v1beta1.External{})
		registrations := []servers.GRPCRegistration{
			func(server *grpc.Server) {
				v1beta1pb.RegisterCardControlsAPIServer(server, cardControlsServer)
			},
		}

		assert.Error(t, RunAPIServer(ctx, cfg, decider, registrations, nil, nil)())
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
	v1beta2pb.RegisterCardControlsAPIServer(svr, &v1beta2pb.UnimplementedCardControlsAPIServer{})

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
	v1beta2 := v1beta2pb.NewCardControlsAPIClient(cc)

	// Tests
	t.Run("SetControlsAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardControlsUpdate, entitlementsRead)

		// Call
		_, err = v1beta2.BlockCard(ctx, &v1beta2pb.BlockCardRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("ListControlsAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardControlsRead, entitlementsRead)

		// Call
		_, err = v1beta2.ListControls(ctx, &v1beta2pb.ListControlsRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("QueryControlsAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardControlsRead, entitlementsRead)

		// Call
		_, err = v1beta2.QueryControls(ctx, &v1beta2pb.QueryControlsRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("RemoveControlsAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardControlsDelete, entitlementsRead)

		// Call
		_, err = v1beta2.RemoveControls(ctx, &v1beta2pb.RemoveControlsRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("SetControlsAcceptsCorrectScopes", func(t *testing.T) {
		ctx := getToken(t, issuer, cardControlsCreate, entitlementsRead)

		// Call
		_, err = v1beta2.SetControls(ctx, &v1beta2pb.SetControlsRequest{})

		// Assert
		assertAuthPassed(t, err)
	})
	t.Run("TransferControlsAcceptsCorrectScopes", func(t *testing.T) {
		// TODO: change to cardcontrols scope in GH-1811
		ctx := getToken(t, issuer, "https://fabric.anz.com/scopes/cardControls:update", entitlementsRead)

		// Call
		_, err = v1beta2.TransferControls(ctx, &v1beta2pb.TransferControlsRequest{})

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
	assert.Equal(t, codes.Unimplemented, ferr.GetStatusCode())
}

package startup

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/internal/service/controls/v1beta1"
	"github.com/anzx/fabric-cards/pkg/servers"
	v1beta1pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/cmd/callback/config/app"
	"github.com/stretchr/testify/assert"
)

func TestRunAPIServer(t *testing.T) {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	decider := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return false }

	cardControlsServer := v1beta1.NewServer(v1beta1.Fabric{}, v1beta1.Internal{}, v1beta1.External{})
	registrations := []servers.GRPCRegistration{
		func(server *grpc.Server) {
			v1beta1pb.RegisterCardControlsAPIServer(server, cardControlsServer)
		},
	}

	t.Run("failed to start server but all checks pass", func(t *testing.T) {
		cfg := app.Spec{
			AppName: "test-app",
			Port:    65536, // highest possible port number + 1
		}
		assert.Error(t, RunAPIServer(ctx, cfg, decider, registrations, nil)())
	})
	t.Run("failed to start server but all checks pass", func(t *testing.T) {
		cfg := app.Spec{
			AppName: "test-app",
			Port:    65536, // highest possible port number + 1
		}
		assert.Error(t, RunAPIServer(ctx, cfg, decider, registrations, nil)())
	})
}

package commandcentre

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/anzx/fabric-cards/test/fixtures"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
)

type testCommandCentre struct {
	hasError bool
}

func (t *testCommandCentre) PublishSync(ctx context.Context, request *sdk.PublishSyncRequest) (*sdk.PublishSyncResponse, error) {
	panic("implement me")
}

func (t *testCommandCentre) Publish(_ context.Context, _ sdk.PublishRequester) (*sdk.PublishResponse, error) {
	if t.hasError {
		return nil, fmt.Errorf("something wrong")
	}
	return &sdk.PublishResponse{
		Status: sdk.PublishResponsePublished,
	}, nil
}

func TestPublishEvent(t *testing.T) {
	// pkg/log does not yet have the ability to add hooks to allow us to check the logs after async call
	// Therefore we call publishEvent instead of PublishEventAsync
	t.Run("logs message when successfully published an event", func(t *testing.T) {
		cc := &testCommandCentre{hasError: false}

		ctx, b := fixtures.GetTestContextWithLogger(nil)

		c := &Client{Publisher: cc}

		c.publishEvent(ctx, event.CardStatusChange)

		assert.Contains(t, b.String(), "successfully published event to CommandCentre: Published")
	})

	t.Run("logs error when failed to published an event", func(t *testing.T) {
		cc := &testCommandCentre{hasError: true}

		ctx, b := fixtures.GetTestContextWithLogger(nil)

		c := &Client{Publisher: cc}

		c.publishEvent(ctx, event.CardStatusChange)

		assert.Contains(t, b.String(), "failed to publish event to CommandCenter: something wrong")
	})
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	env := sdk.EnvironmentLocal
	t.Run("valid config", func(t *testing.T) {
		got, err := NewClient(ctx, &Config{
			Env: &env,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
	})
	t.Run("invalid config", func(t *testing.T) {
		got, err := NewClient(ctx, nil)
		require.NoError(t, err)
		require.Nil(t, got)
	})
	t.Run("sets emulator env variable", func(t *testing.T) {
		got, err := NewClient(ctx, &Config{
			Env:                &env,
			PubsubEmulatorHost: "localhost:8080",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "localhost:8080", os.Getenv("PUBSUB_EMULATOR_HOST"))
	})
}

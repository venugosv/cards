package commandcentre

//go:generate mockgen -build_flags=-mod=mod -destination=mocks/publisher.go -package=commandcentre github.com/anzx/fabric-commandcentre-sdk/pkg/sdk Publisher

import (
	"context"
	"fmt"
	"os"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	anzerrors "github.com/anzx/pkg/errors"

	"google.golang.org/grpc/codes"

	"github.com/anzx/pkg/xcontext"

	"github.com/anzx/fabric-cards/pkg/identity"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"
)

const pubsubEmulatorHostKey = "PUBSUB_EMULATOR_HOST"

type Config struct {
	PubsubEmulatorHost string           `json:"pubsubEmulatorHost" yaml:"pubsubEmulatorHost" mapstructure:"pubsubEmulatorHost"`
	Env                *sdk.Environment `json:"env" yaml:"env" mapstructure:"env"`
}

type Client struct {
	sdk.Publisher
}

func NewClient(ctx context.Context, config *Config) (*Client, error) {
	if config == nil {
		logf.Debug(ctx, "CommandCentre config not provided %v", config)
		return nil, nil
	}

	if config.PubsubEmulatorHost != "" {
		err := os.Setenv(pubsubEmulatorHostKey, config.PubsubEmulatorHost)
		if err != nil {
			logf.Error(ctx, err, "failed to set %s to %s", pubsubEmulatorHostKey, config.PubsubEmulatorHost)
		}
	}

	logf.Debug(ctx, "attempting to connect to command centre")
	commandCentreClient, err := sdk.NewCommandCentre(ctx, *config.Env)
	if err != nil {
		logf.Error(ctx, err, "unable to create client provided in cmdcntr config: %v", config.Env)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create cmdcntr adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "environment not found"))
	}

	return &Client{
		Publisher: commandCentreClient,
	}, nil
}

func (c Client) publishEvent(ctx context.Context, eventType event.Type) {
	id, err := identity.Get(ctx)
	if err != nil {
		logf.Err(ctx, err)
		return
	}

	req := &sdk.EventForPersona{
		Event:     eventType,
		PersonaID: id.PersonaID,
	}

	res, err := c.Publisher.Publish(ctx, req)
	if err != nil {
		logf.Error(ctx, err, fmt.Sprintf("failed to publish event to CommandCenter: %s", err.Error()))
		return
	}
	logf.Info(ctx, fmt.Sprintf("successfully published event to CommandCentre: %v", res.Status))
}

// PublishEventAsync an event to CommandCentre
func (c Client) PublishEventAsync(ctx context.Context, eventType event.Type) {
	go c.publishEvent(xcontext.Detach(ctx), eventType)
}

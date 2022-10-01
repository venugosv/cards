package auditlogger

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/xcontext"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/anzx/pkg/auditlog"
	"github.com/anzx/pkg/auditlog/pubsub"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
)

type Client struct {
	auditlog.Publisher
}

func (c Client) Publish(ctx context.Context, event auditlog.Event, retResponse proto.Message, retError error, serviceData proto.Message) {
	detachedCtx := xcontext.Detach(ctx)
	data, err := anypb.New(serviceData)
	if err != nil {
		logf.Error(ctx, err, "Error marshalling auditlog service data")
	}

	anyResp, err := anypb.New(retResponse)
	if err != nil {
		logf.Error(ctx, err, "Error marshalling auditlog response")
	}

	if err := auditlog.Publish(detachedCtx, c.Publisher,
		auditlog.WithEventName(event),
		auditlog.WithEventFor(auditlog.EventForFalconDetectionNonMon),
		auditlog.WithResponse(anyResp),
		auditlog.WithServiceData(data),
		auditlog.WithError(retError)); err != nil {
		logf.Error(detachedCtx, err, "Error publishing auditLog message")
		return
	}
	logf.Info(detachedCtx, "AuditLog Published Successfully")
}

func NewClient(ctx context.Context, config *auditlog.Config) (*Client, error) {
	if config == nil || config.PubSub == nil {
		logf.Debug(ctx, "audit logger config not provided %v", config)
		return nil, nil
	}

	auditlogPublisher, err := pubsub.NewPubSubClient(ctx, config.PubSub)
	if err != nil {
		logf.Error(ctx, err, "unable to create auditlog pubsub client with config %v", config.PubSub)
		return nil, anzerrors.Wrap(err, codes.Unavailable, "failed to create auditlog adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to create pubsub client"))
	}

	return &Client{
		Publisher: auditlogPublisher,
	}, nil
}

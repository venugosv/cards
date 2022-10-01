package cardcontrols

import (
	"context"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	v1beta2pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc"
)

type Client struct {
	v1beta2pb.CardControlsAPIClient
}

type Config struct {
	BaseURL string `yaml:"baseURL" validate:"required"`
}

func NewClient(ctx context.Context, config *Config, opts ...grpc.DialOption) (*Client, error) {
	if config == nil {
		logf.Debug(ctx, "card controls config not provided %v", config)
		return nil, nil
	}

	opts = append(opts, grpc.WithChainUnaryInterceptor(extractor.MonitorGRPCClientUnaryInterceptor(names.FabricEligibility)))

	cardControlsURL, err := url.Parse(config.BaseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in card controls config: %v", config.BaseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create card controls adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	logf.Debug(ctx, "attempting to dial card controls: %s", cardControlsURL.Host)
	conn, err := grpc.Dial(cardControlsURL.Host, opts...)
	if err != nil {
		logf.Error(ctx, err, "unable to dial card controls service")
		return nil, anzerrors.Wrap(err, codes.Unavailable, "failed to create card controls adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to make successful connection"))
	}

	return &Client{
		CardControlsAPIClient: v1beta2pb.NewCardControlsAPIClient(conn),
	}, nil
}

func (c Client) TransferControls(ctx context.Context, currentTokenizedCardNumber, newTokenizedCardNumber string) error {
	req := &v1beta2pb.TransferControlsRequest{
		CurrentTokenizedCardNumber: currentTokenizedCardNumber,
		NewTokenizedCardNumber:     newTokenizedCardNumber,
	}

	_, err := c.CardControlsAPIClient.TransferControls(ctx, req)

	return err
}

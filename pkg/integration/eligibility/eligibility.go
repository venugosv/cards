package eligibility

import (
	"context"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"google.golang.org/grpc"
)

type Client struct {
	epb.CardEligibilityAPIClient
}

type Config struct {
	BaseURL string `yaml:"baseURL" validate:"required"`
}

func NewClient(ctx context.Context, config *Config, opts ...grpc.DialOption) (*Client, error) {
	if config == nil {
		logf.Debug(ctx, "eligibility config not provided %v", config)
		return nil, nil
	}

	opts = append(opts, grpc.WithChainUnaryInterceptor(extractor.MonitorGRPCClientUnaryInterceptor(names.FabricEligibility)))

	eligibilityURL, err := url.Parse(config.BaseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in eligibility config: %v", config.BaseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create eligibility adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	logf.Debug(ctx, "attempting to dial eligibility: %s", eligibilityURL.Host)
	conn, err := grpc.Dial(eligibilityURL.Host, opts...)
	if err != nil {
		logf.Error(ctx, err, "unable to dial eligibility service")
		return nil, anzerrors.Wrap(err, codes.Unavailable, "failed to create eligibility adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to make successful connection"))
	}

	return &Client{
		CardEligibilityAPIClient: epb.NewCardEligibilityAPIClient(conn),
	}, nil
}

func (c Client) Can(ctx context.Context, operation epb.Eligibility, tokenizedCardNumber string) error {
	req := &epb.CanRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		Eligibility:         operation,
	}

	_, err := c.CardEligibilityAPIClient.Can(ctx, req)

	return err
}

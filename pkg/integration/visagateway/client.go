package visagateway

import (
	"context"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/cardonfile"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway/dcvv2"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Client struct {
	DCVV2         *dcvv2.Client
	CustomerRules *customerrules.Client
	CardOnFile    *cardonfile.Client
}

type Config struct {
	BaseURL  string `yaml:"baseURL" validate:"required"`
	ClientID string `yaml:"clientID"`
}

func NewClient(ctx context.Context, config *Config, opts ...grpc.DialOption) (*Client, error) {
	if config == nil {
		logf.Debug(ctx, "visa gateway config not provided %v", config)
		return nil, nil
	}

	visaGatewayURL, err := url.Parse(config.BaseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in visa gateway config: %v", config.BaseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create visa gateway adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}
	grpcDialOption := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(extractor.MonitorGRPCClientUnaryInterceptor(names.FabricVisaGateway)),
	}
	opts = append(opts, grpcDialOption...)

	logf.Debug(ctx, "attempting to dial visa gateway: %s", visaGatewayURL.Host)
	conn, err := grpc.DialContext(ctx, visaGatewayURL.Host, opts...)
	if err != nil {
		logf.Error(ctx, err, "unable to dial visa gateway service")
		return nil, anzerrors.Wrap(err, codes.Unavailable, "failed to create visa gateway adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to make successful connection"))
	}

	return &Client{
		DCVV2:         dcvv2.NewClient(config.ClientID, conn),
		CustomerRules: customerrules.NewClient(conn),
		CardOnFile:    cardonfile.NewClient(conn),
	}, nil
}

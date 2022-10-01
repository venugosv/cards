package entitlements

import (
	"context"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
	"google.golang.org/grpc"
)

type Carder interface {
	GetEntitledCard(ctx context.Context, tokenizedCardNumber string, operations ...string) (*entpb.EntitledCard, error)
	ListEntitledCards(ctx context.Context) ([]*entpb.EntitledCard, error)
}

type Controller interface {
	Register(ctx context.Context, tokenizedCardNumber string) error
	Latest(ctx context.Context) error
}

type API interface {
	Carder
	Controller
}

type Client struct {
	entpb.CardEntitlementsAPIClient
	entpb.EntitlementsControlAPIClient
}

type Config struct {
	BaseURL string `yaml:"baseURL" validate:"required"`
}

const (
	OPERATION_VIEW_CARD    = "com.anz.x.card.view"
	OPERATION_MANAGE_CARD  = "com.anz.x.card.manage.write"
	OPERATION_CARDCONTROLS = "com.anz.x.card.controls.write"
)

func NewClient(ctx context.Context, config *Config, opts ...grpc.DialOption) (*Client, error) {
	if config == nil {
		logf.Debug(ctx, "entitlements config not provided %v", config)
		return nil, nil
	}
	grpcDialOption := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(extractor.MonitorGRPCClientUnaryInterceptor(names.FabricEntitlements)),
	}
	opts = append(opts, grpcDialOption...)

	entitlementsURL, err := url.Parse(config.BaseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in entitlements config: %v", config.BaseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create entitlements adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	logf.Debug(ctx, "attempting to dial entitlements: %s", entitlementsURL.Host)
	conn, err := grpc.DialContext(ctx, entitlementsURL.Host, opts...)
	if err != nil {
		logf.Error(ctx, err, "unable to dial entitlements service")
		return nil, anzerrors.Wrap(err, codes.Unavailable, "failed to create entitlements adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to make successful connection"))
	}

	return &Client{
		CardEntitlementsAPIClient:    entpb.NewCardEntitlementsAPIClient(conn),
		EntitlementsControlAPIClient: entpb.NewEntitlementsControlAPIClient(conn),
	}, nil
}

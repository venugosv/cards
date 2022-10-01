package selfservice

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"

	anzerrors "github.com/anzx/pkg/errors"

	"google.golang.org/grpc/codes"

	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
	"google.golang.org/grpc"
)

type Client struct {
	sspb.PartyAPIClient
}

type Config struct {
	BaseURL string `yaml:"baseURL" validate:"required"`
}

func NewClient(ctx context.Context, config *Config, opts ...grpc.DialOption) (*Client, error) {
	if config == nil {
		logf.Debug(ctx, "selfservice config not provided %v", config)
		return nil, nil
	}

	selfServiceURL, err := url.Parse(config.BaseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in SelfService config: %v", config.BaseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create SelfService adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	grpcDialOption := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(extractor.MonitorGRPCClientUnaryInterceptor(names.FabricSelfService)),
	}
	opts = append(opts, grpcDialOption...)

	logf.Debug(ctx, "attempting to dial selfservice: %s", selfServiceURL.Host)
	conn, err := grpc.DialContext(ctx, selfServiceURL.Host, opts...)
	if err != nil {
		logf.Error(ctx, err, "unable to dial SelfService service")
		return nil, anzerrors.Wrap(err, codes.Unavailable, "failed to create SelfService adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to make successful connection"))
	}

	return &Client{
		PartyAPIClient: sspb.NewPartyAPIClient(conn),
	}, nil
}

func (c Client) GetParty(ctx context.Context) (*Party, error) {
	getResponse, err := c.PartyAPIClient.GetParty(ctx, &sspb.GetPartyRequest{})
	if err != nil {
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), "SelfService failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, anzerrors.GetErrorInfo(err).GetReason()))
	}

	return &Party{GetPartyResponse: getResponse}, nil
}

type Party struct {
	*sspb.GetPartyResponse
}

func (c Party) GetAddress(ctx context.Context) (*sspb.Address, error) {
	switch {
	case c.GetMailingAddress() != nil:
		return c.GetMailingAddress(), nil
	case c.GetResidentialAddress() != nil:
		return c.GetResidentialAddress(), nil
	default:
		return nil, anzerrors.New(codes.NotFound, "Party Incomplete",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardInvalidAddress, "address not found"))
	}
}

func (c Party) GetName(ctx context.Context) (string, error) {
	firstName := c.GetLegalName().GetFirstName()
	middleName := c.GetLegalName().GetMiddleName()
	lastName := c.GetLegalName().GetLastName()
	if firstName == "" && middleName == "" && lastName == "" {
		return "", anzerrors.New(codes.NotFound, "Party Incomplete",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "empty name returned"))
	}
	name := fmt.Sprintf("%s %s %s", firstName, middleName, lastName)

	return strings.Join(strings.Fields(name), " "), nil
}

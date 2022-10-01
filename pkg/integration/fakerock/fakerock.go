package fakerock

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"

	"google.golang.org/grpc/credentials"

	"github.com/pkg/errors"

	"google.golang.org/grpc"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/metadata"

	frpb "github.com/anzx/fabricapis/pkg/fabric/service/fakerock/v1alpha1"
)

const (
	failMsg = "failed to create fakerock adapter"
)

type Config struct {
	BaseURL         string `yaml:"baseURL"`
	ClientID        string `yaml:"-"`
	ClientSecretKey string `yaml:"clientSecretKey"`
}

type Client struct {
	frpb.FakerockAPIClient
	basicAuth string
}

func NewClient(ctx context.Context, config *Config, opts ...grpc.DialOption) (*Client, error) {
	if config == nil {
		logf.Debug(ctx, "fakerock config not provided %v", config)
		return nil, nil
	}

	fakerockURL, err := url.Parse(config.BaseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in fakerock config: %v", config.BaseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, failMsg,
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	opts = append(opts, grpc.WithChainUnaryInterceptor(extractor.MonitorGRPCClientUnaryInterceptor(names.FabricFakerock)))
	conn, err := setup(ctx, fakerockURL.Scheme == "http", fakerockURL.Host, opts)
	if err != nil {
		logf.Error(ctx, err, "unable to dial fakerock service")
		return nil, anzerrors.Wrap(err, codes.Unavailable, failMsg,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to make successful connection"))
	}

	clientSecret := os.Getenv(config.ClientSecretKey)
	if clientSecret == "" {
		err := errors.New(fmt.Sprintf("unable to load clientSecret from env with key %s", config.ClientSecretKey))
		logf.Err(ctx, err)
		return nil, anzerrors.Wrap(err, codes.Internal, failMsg,
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to find clientSecret"))
	}

	credentials := fmt.Sprintf("%s:%s", config.ClientID, clientSecret)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(credentials))

	return &Client{
		FakerockAPIClient: frpb.NewFakerockAPIClient(conn),
		basicAuth:         basicAuth,
	}, nil
}

func (c Client) ElevateContext(ctx context.Context) (context.Context, error) {
	token, err := c.Token(ctx)
	if err != nil {
		return ctx, err
	}
	in := map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", token),
	}
	return metadata.NewIncomingContext(ctx, metadata.New(in)), nil
}

func (c Client) Token(ctx context.Context) (string, error) {
	auth := fmt.Sprintf("Basic %s", c.basicAuth)
	outgoingContext := metadata.AppendToOutgoingContext(ctx, "Authorization", auth)

	login, err := c.FakerockAPIClient.SystemLogin(outgoingContext, &frpb.SystemLoginRequest{})
	if err != nil {
		return "", anzerrors.Wrap(err, codes.Internal, "failed to get token",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "system login failed"))
	}

	return login.GetToken(), nil
}

func setup(ctx context.Context, plaintext bool, targetURL string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}...)
	if !plaintext {
		cp, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		opts = []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(cp, "")),
		}
	}

	logf.Debug(ctx, "attempting to dial fakerock: %s", targetURL)
	cc, err := grpc.DialContext(ctx, targetURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("%v: failed to connect to server", err)
	}
	return cc, nil
}

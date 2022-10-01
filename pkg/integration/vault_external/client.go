package vault_external

import (
	"context"
	"fmt"
	"net/http"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/errors/errcodes"
	"github.com/cenkalti/backoff/v4"
	"google.golang.org/grpc/codes"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
)

// Client is the interface to Vault API operations
type Client interface {
	Transform(ctx context.Context, kind TransformKind, role string, values []*TransformRequest) ([]*TransformResult, error)
}

// client contains the data needed to integrate with the Vault API
type client struct {
	config             *Config
	api                VaultAPIer
	auth               auth
	jwtSigner          JwtSigner
	metadataHttpClient *http.Client
	backoff            *backoff.ExponentialBackOff
}

// NewClient creates a client for using the Vault API, and initializes a vaultLogin auth
func NewClient(ctx context.Context, httpClient *http.Client, config *Config) (*client, error) {
	if config == nil {
		return nil, errors.New(codes.InvalidArgument,
			"could not create vault client",
			errors.NewErrorInfo(ctx, errcodes.StartupFailure, "config was nil"),
		)
	}

	if httpClient == nil {
		return nil, errors.New(codes.InvalidArgument,
			"could not create vault client",
			errors.NewErrorInfo(ctx, errcodes.StartupFailure, "http client was nil"))
	}

	// The API interface handles the HTTP calls
	api := VaultAPI{
		httpClient: httpClient,
		address:    config.Address,
		namespace:  config.NameSpace,
		loginPath:  fmt.Sprintf("%s%s", config.AuthPath, authLoginPath),
	}

	c := &client{
		config:             config,
		api:                &api,
		metadataHttpClient: httpClient,
		auth: auth{
			renewed:   make(chan interface{}),
			blockTime: config.BlockForTokenTime,
		},
	}

	c.backoff = &backoff.ExponentialBackOff{
		InitialInterval:     config.TokenErrorRetryFirstTime,
		MaxElapsedTime:      config.TokenErrorRetryMaxTime,
		MaxInterval:         backoff.DefaultMaxInterval,
		Multiplier:          backoff.DefaultMultiplier,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Clock:               backoff.SystemClock,
	}

	c.backoff.Reset()

	if config.NoGoogleCredentialsClient {
		c.jwtSigner = &FixedSignedJwt{
			jwt: "foo",
			key: "bar",
		}
	} else {
		credsClient, err := credentials.NewIamCredentialsClient(ctx)
		if err != nil {
			return nil, err
		}
		c.jwtSigner = credsClient
	}

	logf.Debug(ctx, "Vault attempting first vaultLogin/auth")

	loginResponse, err := c.login(ctx)
	if err != nil {
		return nil, err
	}

	duration := time.Second * time.Duration(loginResponse.LeaseDuration)
	c.auth.set(loginResponse.ClientToken, duration)

	keepAuthValid(ctx, c)

	return c, nil
}

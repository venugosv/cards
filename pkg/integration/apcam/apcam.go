package apcam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"

	"go.opentelemetry.io/otel/trace"

	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/names"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/rest"
	"github.com/anzx/fabric-cards/pkg/util/apic"
)

const pushProvisionEndpoint = "/internal/in-app-provisioning-data"

type Config struct {
	BaseURL        string `json:"baseURL"             yaml:"baseURL"             mapstructure:"baseURL"        validate:"required"`
	ClientIDEnvKey string `json:"clientIDEnvKey"      yaml:"clientIDEnvKey"      mapstructure:"clientIDEnvKey" validate:"required"`
	ClientID       string `json:"clientID,omitempty"  yaml:"clientID,omitempty"  mapstructure:"clientID"`
	MaxRetries     int    `json:"maxRetries"          yaml:"maxRetries"          mapstructure:"maxRetries"     validate:"required"`
}

type Client interface {
	PushProvision(ctx context.Context, in Request) (*Response, error)
}

type apcam struct {
	baseURL    string
	apicClient apic.Clienter
}

func ClientFromConfig(ctx context.Context, httpClient *http.Client, config *Config, gsmClient *gsm.Client) (Client, error) {
	if config == nil {
		logf.Debug(ctx, "apcam config not provided %v", config)
		return nil, nil
	}

	return NewClient(ctx, config.BaseURL, config.ClientIDEnvKey, httpClient, config.MaxRetries, gsmClient)
}

func NewClient(ctx context.Context, baseURL string, clientIDEnvKey string, httpClient *http.Client, maxRetries int, gsmClient *gsm.Client) (Client, error) {
	if httpClient == nil {
		// TODO: (GH-1936) update `name.Service` to be new name.apcam once new enum value is added to monitoring package
		httpClient = rest.NewHTTPClientWithLogAndRetry(maxRetries, nil, names.Unknown)
	}

	destination, err := url.Parse(baseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in APCAM config: %v", baseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create APCAM adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	apicClient, err := apic.NewAPICClient(ctx, clientIDEnvKey, httpClient, gsmClient)
	if err != nil {
		return nil, err
	}

	return &apcam{
		baseURL:    destination.String(),
		apicClient: apicClient,
	}, nil
}

// This service is responsible for the post processing of a provisioning
// event where a card has been provisioned with a token / provisioned in a wallet
const failedRequest = "failed push provision request"

func (c *apcam) PushProvision(ctx context.Context, in Request) (*Response, error) {
	traceID := trace.SpanContextFromContext(ctx).TraceID().String()

	in.TraceInfo = TraceInfo{
		MessageID:      traceID,
		ConversationID: traceID,
	}

	req, err := json.Marshal(in)
	if err != nil {
		logf.Error(ctx, err, "apcam:PushProvision failed to marshall request ")
		return nil, anzerrors.Wrap(err, codes.InvalidArgument, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to marshall request"))
	}

	target := fmt.Sprintf("%s%s", c.baseURL, pushProvisionEndpoint)

	body, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, target, req), "apcam:PushProvision")
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, errMsg(body)))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		logf.Error(ctx, err, "apcam:PushProvision failed to unmarshalling response")
		return nil, anzerrors.Wrap(err, codes.Internal, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}
	return &response, nil
}

func errMsg(in []byte) string {
	var resp ErrorInfo
	if err := json.Unmarshal(in, &resp); err != nil {
		return "unexpected response from downstream"
	}
	return fmt.Sprintf("error code: %s - %s", resp.ErrorCode, resp.ErrorDescription)
}

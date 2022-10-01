package ctm

import (
	"context"
	"net/http"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"

	"github.com/anzx/fabric-cards/pkg/rest"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/util/apic"
)

const (
	debitCardInqUrlTemplate     = "%s/debit-card-inquiry/debit-cards/%s?detail=true"
	replaceAPIUrlTemplate       = "%s/debit-card-maintenance/debit-cards/%s/replace"
	preferenceAPIUrlTemplate    = "%s/debit-card-maintenance/debit-cards/%s/preferences"
	updateDetailsAPIUrlTemplate = "%s/debit-card-maintenance/debit-cards/%s/details"
	activationAPIUrlTemplate    = "%s/debit-card-status/debit-cards/%s/activate"
	statusAPIUrlTemplate        = "%s/debit-card-status/debit-cards/%s/status"
	pinInfoUpdateUrlTemplate    = "%s/debit-card-pin-info-update/debit-cards/%s/pin-info/update"
)

type Config struct {
	BaseURL        string `json:"baseURL"             yaml:"baseURL"             mapstructure:"baseURL"         validate:"required"`
	ClientIDEnvKey string `json:"clientIDEnvKey"      yaml:"clientIDEnvKey"      mapstructure:"clientIDEnvKey"  validate:"required"`
	MaxRetries     int    `json:"maxRetries"          yaml:"maxRetries"          mapstructure:"maxRetries"      validate:"required"`
}

type CardMaintenanceAPI interface {
	ReplaceCard(context.Context, *ReplaceCardRequest, string) (string, error)
	UpdatePreferences(ctx context.Context, req *UpdatePreferencesRequest, tokenizedCardNumber string) (bool, error)
	UpdateDetails(ctx context.Context, req *UpdateDetailsRequest, tokenizedCardNumber string) (bool, error)
}

type CardInquiryAPI interface {
	DebitCardInquiry(context.Context, string) (*DebitCardResponse, error)
}

type StatusAPI interface {
	Activate(context.Context, string) (bool, error)
	UpdateStatus(ctx context.Context, tokenizedCardNumber string, status Status) (bool, error)
}

type PINInfoUpdateAPI interface {
	UpdatePINInfo(ctx context.Context, tokenizedCardNumber string) (bool, error)
}

type Client interface {
	CardMaintenanceAPI
	CardInquiryAPI
	StatusAPI
	PINInfoUpdateAPI
}

type ControlAPI interface {
	CardMaintenanceAPI
	CardInquiryAPI
	StatusAPI
}

type client struct {
	baseURL    string
	apicClient apic.Clienter
}

func ClientFromConfig(ctx context.Context, httpClient *http.Client, config *Config, gsmClient *gsm.Client) (Client, error) {
	if config == nil {
		logf.Debug(ctx, "ctm config not provided %v", config)
		return nil, nil
	}

	return NewClient(ctx, config.BaseURL, config.ClientIDEnvKey, httpClient, config.MaxRetries, gsmClient)
}

func NewClient(ctx context.Context, baseURL string, clientIDEnvKey string, httpClient *http.Client, maxRetries int, gsmClient *gsm.Client) (Client, error) {
	if httpClient == nil {
		httpClient = rest.NewHTTPClientWithLogAndRetry(maxRetries, nil, names.CTM)
	}

	destination, err := url.Parse(baseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in CTM config: %v", baseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create CTM adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	apicClient, err := apic.NewAPICClient(ctx, clientIDEnvKey, httpClient, gsmClient)
	if err != nil {
		return nil, err
	}

	return &client{
		baseURL:    destination.String(),
		apicClient: apicClient,
	}, nil
}

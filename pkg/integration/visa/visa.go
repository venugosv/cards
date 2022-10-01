package visa

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/rest"

	"github.com/anzx/fabric-cards/pkg/util/apic"
)

type Config struct {
	BaseURL        string `json:"baseURL"             yaml:"baseURL"              mapstructure:"baseURL"              validate:"required"`
	ClientIDEnvKey string `json:"clientIDEnvKey"      yaml:"clientIDEnvKey"       mapstructure:"clientIDEnvKey"       validate:"required"`
	ClientID       string `json:"clientID,omitempty"  yaml:"clientID,omitempty"   mapstructure:"clientID,omitempty"`
	MaxRetries     int    `json:"maxRetries"          yaml:"maxRetries"           mapstructure:"maxRetries"           validate:"required"`
}

const (
	queryByPanEndpoint      = "customerrules/consumertransactioncontrols/inquiries/cardinquiry"
	enrolByPanEndpoint      = "customerrules/consumertransactioncontrols"
	cardReplacementEndpoint = "customerrules/consumertransactioncontrols/accounts/accountupdate"
	setControlPrefix        = "customerrules/consumertransactioncontrols"
	setControlSuffix        = "rules"
)

// CustomerRulesAPI - Retrieve, create, modify or delete controls on a registered primaryAccountNumber or paymentToken
// The Customer Rules API is used to enroll, configure and retrieve an account's card control settings.
type CustomerRulesAPI interface {
	Register(ctx context.Context, primaryAccountNumber string) (string, error)
	QueryControls(ctx context.Context, primaryAccountNumber string) (*Resource, error)
	CreateControls(ctx context.Context, documentID string, request *Request) (*Resource, error)
	UpdateControls(ctx context.Context, documentID string, request *Request) (*Resource, error)
	DeleteControls(ctx context.Context, documentID string, request *Request) (*Resource, error)
	ReplaceCard(ctx context.Context, currentAccountID, newAccountID string) (bool, error)
}

type client struct {
	apicClient apic.Clienter
	baseURL    string
}

func ClientFromConfig(ctx context.Context, httpClient *http.Client, config *Config, gsmClient *gsm.Client) (CustomerRulesAPI, error) {
	if config == nil {
		logf.Debug(ctx, "visa config not provided %v", config)
		return nil, nil
	}

	return NewClient(ctx, config.BaseURL, config.ClientIDEnvKey, httpClient, config.MaxRetries, gsmClient)
}

func NewClient(ctx context.Context, baseURL string, clientIDEnvKey string, httpClient *http.Client, maxRetries int, gsmClient *gsm.Client) (CustomerRulesAPI, error) {
	if httpClient == nil {
		httpClient = rest.NewHTTPClientWithLogAndRetry(maxRetries, nil, names.VISA)
	}

	destination, err := url.Parse(baseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in entitlements config: %v", baseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create visa adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to parse configured url"))
	}

	apicClient, err := apic.NewAPICClient(ctx, clientIDEnvKey, httpClient, gsmClient)
	if err != nil {
		return nil, err
	}

	return &client{
		apicClient: apicClient,
		baseURL:    destination.String(),
	}, nil
}

// CreateControls Creates a new control(s) within the Transaction Control Document
func (c client) CreateControls(ctx context.Context, documentID string, request *Request) (*Resource, error) {
	if ok := checkRequest(documentID, request); !ok {
		return nil, anzerrors.New(codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to make request with provided value"))
	}

	destinationURL := fmt.Sprintf("%s/%s/%s/%s", c.baseURL, setControlPrefix, documentID, setControlSuffix)

	return sendControlRequest(ctx, c.apicClient, http.MethodPost, destinationURL, request)
}

// UpdateControls updates existing control(s) within the Transaction Control Document
func (c client) UpdateControls(ctx context.Context, documentID string, request *Request) (*Resource, error) {
	if ok := checkRequest(documentID, request); !ok {
		return nil, anzerrors.New(codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to make request with provided value"))
	}

	destinationURL := fmt.Sprintf("%s/%s/%s/%s", c.baseURL, setControlPrefix, documentID, setControlSuffix)

	return sendControlRequest(ctx, c.apicClient, http.MethodPut, destinationURL, request)
}

// DeleteControls deletes existing control(s) within the Transaction Control Document
func (c client) DeleteControls(ctx context.Context, documentID string, request *Request) (*Resource, error) {
	if ok := checkRequest(documentID, request); !ok {
		return nil, anzerrors.New(codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to make request with provided value"))
	}

	destinationURL := fmt.Sprintf("%s/%s/%s/%s", c.baseURL, setControlPrefix, documentID, setControlSuffix)

	return sendControlRequest(ctx, c.apicClient, http.MethodDelete, destinationURL, request)
}

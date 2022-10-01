package echidna

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"

	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/names"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/rest"
	"github.com/anzx/fabric-cards/pkg/util/apic"
)

const (
	fabric             = "Fabric"
	format             = "%s/%s/%s"
	cardPINServicesAPI = "card-and-pin-services"
)

type Config struct {
	BaseURL        string `json:"baseURL"             yaml:"baseURL"             mapstructure:"baseURL"        validate:"required"`
	ClientIDEnvKey string `json:"clientIDEnvKey"      yaml:"clientIDEnvKey"      mapstructure:"clientIDEnvKey" validate:"required"`
	ClientID       string `json:"clientID,omitempty"  yaml:"clientID,omitempty"  mapstructure:"clientID"`
	MaxRetries     int    `json:"maxRetries"          yaml:"maxRetries"          mapstructure:"maxRetries"     validate:"required"`
}

type Echidna interface {
	GetWrappingKey(context.Context) (string, error)
	SelectPIN(context.Context, IncomingRequest) error
	VerifyPIN(context.Context, IncomingRequest) error
	ChangePIN(context.Context, IncomingChangePINRequest) error
}

type client struct {
	baseURL    string
	apicClient apic.Clienter
}

func ClientFromConfig(ctx context.Context, httpClient *http.Client, config *Config, gsmClient *gsm.Client) (Echidna, error) {
	if config == nil {
		logf.Debug(ctx, "echidna config not provided %v", config)
		return nil, nil
	}

	return NewClient(ctx, config.BaseURL, config.ClientIDEnvKey, httpClient, config.MaxRetries, gsmClient)
}

func NewClient(ctx context.Context, baseURL string, clientIDEnvKey string, httpClient *http.Client, maxRetries int, gsmClient *gsm.Client) (Echidna, error) {
	if httpClient == nil {
		httpClient = rest.NewHTTPClientWithLogAndRetry(maxRetries, nil, names.Echidna)
	}

	destination, err := url.Parse(baseURL)
	if err != nil {
		logf.Error(ctx, err, "unable to parse url provided in Echidna config: %v", baseURL)
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create Echidna adapter",
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

// Retrieves the public key information from Salt Echidna-RemotePIN Server. The public key information are required to
// pass into the Salt client side SDK for PIN Block protection. The client side application should be retrieving the
// public key information every time prior to calling the Salt client side SDK for PIN Block protection. This ensures
// the PIN Block is protected by the latest public key information, especially when dynamic key rollover feature is
// enabled on the Echidna-RemotePIN application.
func (c client) GetWrappingKey(ctx context.Context) (string, error) {
	getWrappingKeyRequest := GetWrappingKeyRequest{
		OperatorID: fabric,
		RespFormat: formatJson,
		LogLevel:   LoglevelInfo,
	}

	body, _ := json.Marshal(getWrappingKeyRequest)

	respBody, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, c.getURL(ActionGetWrappingKey), body), "echidna:GetWrappingKey")
	if err != nil {
		return "", err
	}

	var response GetWrappingKeyResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", unexpectedResp(ctx, err, ActionGetWrappingKey)
	}

	if err := checkCode(ctx, response.Response.Result); err != nil {
		return "", err
	}

	return *response.Response.Result.EncodedKey, nil
}

// Submits a Select-PIN or Set-PIN request for a card via the RemotePIN service. The encrypted PIN block is should be
// obtained from the client side application, which utilises the RemotePIN client SDK. It is recommended to obtain the
// PAN from internal card management system. The txnInfo parameters are additional transaction details for the Select
// PIN operation.
func (c client) SelectPIN(ctx context.Context, in IncomingRequest) error {
	body, err := c.send(ctx, in, ActionSelect)
	if err != nil {
		return err
	}

	var response SetPINResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return unexpectedResp(ctx, err, ActionSelect)
	}

	return checkCode(ctx, response.Response.Result)
}

// Submits a Verify-PIN request for a card via the RemotePIN service. The encrypted PIN block should be obtained from
// the client side application, which utilises the RemotePIN client SDK. It is recommended to obtain the PAN from
// internal card management system. The txnInfo parameters are additional transaction details for the Verify PIN operation.
func (c client) VerifyPIN(ctx context.Context, in IncomingRequest) error {
	body, err := c.send(ctx, in, ActionVerify)
	if err != nil {
		return err
	}

	var response VerifyPINResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return unexpectedResp(ctx, err, ActionVerify)
	}

	return checkCode(ctx, response.Response.Result)
}

func (c client) send(ctx context.Context, in IncomingRequest, action Action) ([]byte, error) {
	if in.PlainPAN == "" || in.EncryptedPINBlock == "" {
		return nil, anzerrors.New(codes.InvalidArgument, "echidna failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "mandatory request fields not supplied"))
	}

	req := newRequest(in.PlainPAN, in.EncryptedPINBlock)

	return c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, c.getURL(action), req), fmt.Sprintf("echidna:%s", action))
}

// Submits a Change-PIN request for a card via the RemotePIN service. The encrypted PIN blocks should be obtained from
// the client side application, which utilises the RemotePIN client SDK. This operation requires submitting both the
// Old-PIN and the New-PIN. It is recommended to obtain the PAN from internal card management system. The txnInfo
// parameters are additional transaction details for the Change PIN operation.
func (c client) ChangePIN(ctx context.Context, in IncomingChangePINRequest) error {
	if in.PlainPAN == "" || in.EncryptedPINBlockOld == "" || in.EncryptedPINBlockNew == "" {
		return anzerrors.New(codes.InvalidArgument, "echidna failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "mandatory request fields not supplied"))
	}

	req := newChangePINRequest(in.PlainPAN, in.EncryptedPINBlockOld, in.EncryptedPINBlockNew)

	body, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, c.getURL(ActionChange), req), "echidna:changePIN")
	if err != nil {
		return err
	}

	var response ChangePINResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return unexpectedResp(ctx, err, ActionChange)
	}

	return checkCode(ctx, response.Response.Result)
}

func checkCode(ctx context.Context, result Result) error {
	if result.Code != 0 {
		logf.Error(ctx, fmt.Errorf("(%d) %s", result.Code, result.Message), "echidna call failed")
		return anzerrors.New(GetGRPCError(result.Code), "failed request",
			anzerrors.NewErrorInfo(ctx, GetANZError(result.Code), GetErrorMsg(result.Code)))
	}
	return nil
}

func (c client) getURL(action Action) string {
	return fmt.Sprintf(format, c.baseURL, cardPINServicesAPI, action)
}

func unexpectedResp(ctx context.Context, err error, action Action) error {
	logf.Error(ctx, err, "echidna %s failed unexpected response from downstream", action)
	return anzerrors.Wrap(err, codes.Internal, "failed request",
		anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
}

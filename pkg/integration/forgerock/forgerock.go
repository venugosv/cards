package forgerock

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"

	"github.com/anzx/fabric-cards/pkg/rest"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"

	"github.com/anzx/pkg/log"
)

// Clienter is the interface for calling the forgerock
type Clienter interface {
	SystemJWT(context.Context, ...string) (context.Context, error)
}

const (
	headerContentType = "application/x-www-form-urlencoded"

	startupFailure        = "failed to create forgerock client"
	failure               = "could not get token"
	requestBuilderFailure = "could not create downstream request"
	downstreamFailure     = "downstream service failure"
	invalidResponse       = "downstream returned invalid response"

	contentType = "Content-Type"
	xRequestID  = "x-request-id"

	path  = "token"
	query = "requested_token_type=SystemJWT"

	getTokenClientID     = "client_id"
	getTokenClientSecret = "client_secret"
	getTokenScope        = "scope"
)

// TokenReq is the request to forgerock token endpoint.
type TokenReq struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
}

// TokenResp is the response to forgerock token endpoint.
type TokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Config struct {
	BaseURL         string `yaml:"baseURL,omitempty"         json:"baseURL,omitempty"`
	ClientID        string `yaml:"clientID,omitempty"        json:"clientID,omitempty"`
	ClientSecretKey string `yaml:"clientSecretKey,omitempty" json:"clientSecretKey,omitempty"`
}

// ClientFromConfig creates and returns a new Clienter instance.
func ClientFromConfig(ctx context.Context, httpClient *http.Client, config *Config, gsmClient *gsm.Client) (Clienter, error) {
	if config == nil {
		return nil, nil
	}
	return NewClient(ctx, httpClient, config.BaseURL, config.ClientID, config.ClientSecretKey, gsmClient)
}

// NewClient creates and returns a new Clienter instance.
func NewClient(ctx context.Context, httpClient *http.Client, clientURL string, clientID string, clientSecretKey string, gsmClient *gsm.Client) (Clienter, error) {
	if httpClient == nil {
		httpClient = rest.NewHTTPClientWithLog(http.DefaultTransport, nil, names.ForgeRock)
	}

	if clientSecretKey == "" {
		err := anzerrors.New(codes.Internal, startupFailure,
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "forgerock secret key was empty"))
		logf.Error(ctx, err, startupFailure)
		return nil, err
	}

	clientSecret, err := gsmClient.AccessSecret(ctx, clientSecretKey)
	if err != nil || clientSecret == "" {
		log.Error(ctx, err, "unable to load forgerock client secret from gsm", log.Str("key", clientSecretKey))
		return nil, anzerrors.Wrap(err, codes.Internal, startupFailure,
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to find client secret"))
	}

	return &Client{
		clientURL:    clientURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		client:       httpClient,
	}, nil
}

type Client struct {
	clientURL    string
	clientID     string
	clientSecret string
	client       *http.Client
}

func (f Client) SystemJWT(ctx context.Context, scope ...string) (context.Context, error) {
	token, err := f.getToken(ctx, scope...)
	if err != nil {
		return ctx, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md.Set("authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	return metadata.NewIncomingContext(ctx, md), nil
}

func (f Client) getToken(ctx context.Context, scope ...string) (*TokenResp, error) {
	logf.Info(ctx, "Getting Token from Forgerock")

	form := url.Values{}
	form.Set(getTokenClientID, f.clientID)
	form.Set(getTokenClientSecret, f.clientSecret)
	form.Set(getTokenScope, strings.Join(scope, " "))
	reqBody := form.Encode()

	target := fmt.Sprintf("%s/%s?%s", f.clientURL, path, query)

	httpReq, err := http.NewRequest(http.MethodPost, target, strings.NewReader(reqBody))
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, failure,
			anzerrors.NewErrorInfo(ctx, anzcodes.Unknown, requestBuilderFailure))
	}
	httpReq.Header.Add(contentType, headerContentType)
	httpReq.Header.Add(xRequestID, getOrCreateRequestID(ctx))

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, failure,
			anzerrors.NewErrorInfo(ctx, anzcodes.Unknown, downstreamFailure))
	}
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode {
	case http.StatusOK:
		logf.Debug(ctx, "forge token received")
		return parseResponse(ctx, resp)
	default:
		logf.Error(ctx, err, "failed to fetch forge token")
		return nil, anzerrors.Wrap(handleErrorResponse(resp), codes.Internal, failure,
			anzerrors.NewErrorInfo(ctx, anzcodes.Unknown, downstreamFailure))
	}
}

func parseResponse(ctx context.Context, res *http.Response) (*TokenResp, error) {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, failure,
			anzerrors.NewErrorInfo(ctx, anzcodes.Unknown, invalidResponse))
	}
	defer res.Body.Close()

	var payload TokenResp
	if err = json.Unmarshal(resBody, &payload); err != nil {
		logf.Error(ctx, err, "Cannot json unmarshal response, http status %d: %s", res.StatusCode, string(resBody))
		return nil, anzerrors.Wrap(err, codes.Internal, failure,
			anzerrors.NewErrorInfo(ctx, anzcodes.Unknown, invalidResponse))
	}
	return &payload, nil
}

func handleErrorResponse(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("received status code %d, could not handle error response, could not read error body: %w", res.StatusCode, err)
	}
	defer res.Body.Close()

	if len(body) == 0 {
		return fmt.Errorf("received status code %d, could not handle error response, response contained no error body", res.StatusCode)
	}
	return fmt.Errorf("received status code %d from downstream service with body %s", res.StatusCode, strings.TrimSuffix(string(body), "\n"))
}

func getOrCreateRequestID(ctx context.Context) string {
	incoming, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.New().String()
	}
	header, ok := incoming[xRequestID]
	if !ok || len(header) == 0 {
		return uuid.New().String()
	}

	return header[0]
}

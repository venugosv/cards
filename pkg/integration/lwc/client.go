package lwc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/rest"
	"github.com/anzx/pkg/gsm"
	"github.com/anzx/pkg/monitoring/names"
	"github.com/pkg/errors"
)

const (
	defaultDialTimeout         = time.Second * 3
	defaultIdleConnTimeout     = time.Second * 10
	defaultTimeToHeaderTimeout = time.Second * 5
	retrieveMerchantsEndpoint  = "/api/v4/search/banktransaction"
)

type Config struct {
	BaseURL string       `json:"baseURL"             yaml:"baseURL"             mapstructure:"baseURL"        validate:"required"`
	Proxy   *ProxyConfig `json:"proxy"               yaml:"proxy"               mapstructure:"proxy"`
}

type ProxyConfig struct {
	Host              string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host,omitempty"`
	Username          string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username,omitempty"`
	PasswordSecretKey string `json:"password_secret_key,omitempty" yaml:"password_secret_key,omitempty" mapstructure:"password_secret_key,omitempty"`
}

type Client interface {
	RetrieveMerchants(ctx context.Context, in Request) ([]MerchantDetails, error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(ctx context.Context, cfg *Config, httpClient *http.Client, secrets gsm.Client) (Client, error) {
	if cfg == nil {
		logf.Debug(ctx, "Accounts config not provided %v", cfg)
		return nil, nil
	}

	destination, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to configure proxy URL")
	}

	proxy, err := cfg.Proxy.new(ctx, &secrets)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		transport := newDefaultTransportWithTimeout(nil, proxy)
		httpClient = rest.NewHTTPClientWithLog(transport, nil, names.Unknown)
	}

	return &client{
		baseURL:    destination.String(),
		httpClient: httpClient,
	}, nil
}

func newDefaultDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: defaultDialTimeout}
	return dialer.DialContext
}

func newDefaultTransportWithTimeout(tls *tls.Config, proxy func(*http.Request) (*url.URL, error)) *http.Transport {
	return &http.Transport{
		Proxy:                 proxy,
		DialContext:           newDefaultDialContext(),
		IdleConnTimeout:       defaultIdleConnTimeout,
		ResponseHeaderTimeout: defaultTimeToHeaderTimeout,
		TLSClientConfig:       tls,
	}
}

func (c *ProxyConfig) new(ctx context.Context, secrets *gsm.Client) (func(*http.Request) (*url.URL, error), error) {
	if c == nil {
		return func(request *http.Request) (*url.URL, error) { return nil, nil }, nil
	}

	password, err := secrets.AccessSecret(ctx, c.PasswordSecretKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to access secret")
	}

	logf.Info(ctx, "building proxy URL for host: %s", c.Host)
	proxy := fmt.Sprintf("http://%s:%s@%s", c.Username, password, c.Host)

	u, err := url.Parse(proxy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to configure proxy URL")
	}

	return http.ProxyURL(u), nil
}

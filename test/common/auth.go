package common

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/test/config"
	"github.com/google/uuid"

	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/anzx/pkg/jwtauth"
	"github.com/anzx/pkg/jwtauth/jwttest"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/metadata"

	"github.com/anzx/anzdata"
	"github.com/anzx/utils/forgejwt/v2"
	"github.com/googleapis/gax-go/v2"
)

const (
	envHeader           = "env"
	authorizationHeader = "Authorization"
	cookieHeader        = "Cookie"
)

var (
	once           sync.Once
	authOptCreator func(anzdata.Auther) anzdata.Auther
)

func authCache(opt anzdata.Auther) anzdata.Auther {
	once.Do(func() {
		authOptCreator, _ = anzdata.AuthCacherOpt()
	})
	return authOptCreator(opt)
}

type authHeaders struct {
	headers  map[string]string
	env      forgejwt.Env
	backoff  gax.Backoff
	deadline time.Time
	user     anzdata.User
}

func newGRPCHeaders(env forgejwt.Env, user anzdata.User) *authHeaders {
	return &authHeaders{
		headers:  map[string]string{},
		env:      env,
		backoff:  gax.Backoff{Initial: time.Second, Multiplier: 2},
		deadline: time.Now().Add(time.Second * 30),
		user:     user,
	}
}

func GetAuthHeaders(t *testing.T, user anzdata.User, cfg config.AuthConfig, scopes ...string) *authHeaders {
	t.Logf("Target Environment: %s\n", cfg.Env)

	out := newGRPCHeaders(cfg.Env, user)

	out.setEnvHeader(cfg.Env)

	switch cfg.Method {
	case config.AuthMethodForgejwt:
		out.authMethodForgeJWT(t, scopes...)
	case config.AuthMethodForgesso:
		out.authMethodForgeSSO(t)
	case config.AuthMethodBasic:
		out.authMethodBasic(t)
	case config.AuthMethodFakeJWT:
		out.authMethodFakeJWT(t, scopes...)
	case config.AuthMethodCustomJWT:
		out.authMethodCustomJWT(t, scopes...)
	case config.AuthMethodNone:
		return out
	default:
		t.Fatalf("unrecognised auth method %s", cfg.Method)
	}

	return out
}

func (a authHeaders) authMethodForgeJWT(t *testing.T, scopes ...string) {
	t.Log("header added for auth method forge JWT")
	opt := anzdata.AuthRetry{
		Logger: log.Printf,
		//nolint:misspell
		Auther: authCache(anzdata.AuthForgeJWT{
			Env:    a.env,
			Scopes: scopes,
		}),
		Bo:       a.backoff,
		Deadline: a.deadline,
	}

	jwt, err := a.user.Auth(opt)
	if err != nil {
		t.Fatal("Failed to get ForgeJWT with err", err)
	}

	a.headers[authorizationHeader] = fmt.Sprintf("Bearer %s", jwt)
}

func (a authHeaders) authMethodForgeSSO(t *testing.T) {
	t.Log("header added for auth method forge SSO")
	opt := anzdata.AuthRetry{
		Logger: log.Printf,
		//nolint:misspell
		Auther: authCache(anzdata.AuthForgeSSO{
			Env: a.env,
		}),
		Bo:       a.backoff,
		Deadline: a.deadline,
	}

	token, err := a.user.Auth(opt)
	if err != nil {
		t.Fatal("Failed to get ForgeSSO with err", err)
	}

	a.headers[cookieHeader] = fmt.Sprintf("anzssotoken=%s", token)
}

func (a authHeaders) authMethodBasic(t *testing.T) {
	t.Log("header added for auth method basic")
	a.headers[authorizationHeader] = a.user.BasicAuth()
}

func (a authHeaders) authMethodFakeJWT(t *testing.T, scopes ...string) {
	t.Log("header added for auth method fake JWT")
	auth := a.user.MustAuth(anzdata.AuthJWT{Claims: map[string]interface{}{"scopes": scopes}})
	a.headers[authorizationHeader] = fmt.Sprintf("Bearer %s", auth)
}

func (a authHeaders) authMethodCustomJWT(t *testing.T, scopes ...string) {
	t.Log("header added for auth method custom JWT")
	issuer, err := jwttest.NewIssuer("test", 1024)
	require.NoError(t, err)
	token, err := issuer.Issue(jwtauth.BaseClaims{
		Claims:           jwt.Claims{},
		AuthContextClass: "",
		AuthMethods:      nil,
		Scopes:           scopes,
		DeviceInfo:       nil,
		Device:           "",
		OCVID:            "",
		Persona:          nil,
		Actor:            nil,
		RequestedFor:     nil,
	})
	require.NoError(t, err)
	a.headers[authorizationHeader] = fmt.Sprintf("Bearer %s", token)
}

func (a authHeaders) setEnvHeader(env forgejwt.Env) {
	if env == forgejwt.SitL {
		a.headers[envHeader] = "sit"
	}
}

// Context sets an environment specific 'authorization' header into context
func (a authHeaders) Context(t *testing.T, ctx context.Context, headers ...string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", uuid.NewString())
	t.Log("adding headers:")
	for key, val := range a.headers {
		t.Logf("- %s: %s", key, val)
		ctx = metadata.AppendToOutgoingContext(ctx, key, val)
	}
	t.Logf("custom headers: %v", headers)
	return metadata.AppendToOutgoingContext(ctx, headers...)
}

// GetHeadersHTTP generates environment specific 'authorization' headers for a request
func (a authHeaders) GetHeadersHTTP(headers ...string) http.Header {
	out := http.Header{}
	for key, val := range a.headers {
		out.Add(key, val)
	}
	for i := 0; i < len(headers); i += 2 {
		out.Add(strings.ToLower(headers[i]), headers[i+1])
	}
	out.Add("Accept", "*/*")
	return out
}

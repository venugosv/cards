package forgerock

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/anzx/fabric-cards/test/stubs/utils"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"
)

var visaGatewayScopes = []string{
	"https://fabric.anz.com/scopes/visaGateway:create",
	"https://fabric.anz.com/scopes/visaGateway:read",
	"https://fabric.anz.com/scopes/visaGateway:update",
	"https://fabric.anz.com/scopes/visaGateway:delete",
}

type StubServer struct{}

func NewStubServer(_ context.Context) *StubServer {
	return &StubServer{}
}

func AppendRoutes(ctx context.Context, router *http.ServeMux) {
	router.Handle("/forgerock/", handler(ctx))
}

func handler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewStubServer(ctx).ServeHTTP(w, r)
	})
}

func (s *StubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenURL := regexp.MustCompile("/forgerock/")

	switch {
	case tokenURL.MatchString(r.URL.Path):
		s.tokenHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *StubServer) tokenHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := utils.GetRequestBody(w, r, http.MethodPost, http.MethodGet)
	if !ok {
		return
	}
	defer r.Body.Close()

	token, err := getToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(out)
}

func getToken() (forgerock.TokenResp, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return forgerock.TokenResp{}, err
	}
	key := jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}

	signerOpts := jose.SignerOptions{}
	signerOpts.WithType("JWT")
	rsaSigner, err := jose.NewSigner(key, &signerOpts)
	if err != nil {
		return forgerock.TokenResp{}, err
	}

	claims := map[string]interface{}{
		"scopes": visaGatewayScopes,
	}
	token, err := jwt.Signed(rsaSigner).Claims(claims).CompactSerialize()
	if err != nil {
		return forgerock.TokenResp{}, err
	}

	return getJwt(token), nil
}

func getJwt(in string) forgerock.TokenResp {
	return forgerock.TokenResp{
		AccessToken: in,
		TokenType:   "SYSTEM",
		ExpiresIn:   300,
	}
}

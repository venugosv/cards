package vault

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/anzx/fabric-cards/pkg/integration/vault_external"
	"github.com/brianvoe/gofakeit/v6"

	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

type StubServer struct{}

func NewStubServer() *StubServer {
	return &StubServer{}
}

func AppendRoutes(_ context.Context, router *http.ServeMux) {
	router.Handle("/vault/", handle())
}

func handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewStubServer().ServeHTTP(w, r)
	})
}

func (s *StubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		login     = regexp.MustCompile(".*/login.*")
		lookup    = regexp.MustCompile(".*/v1/auth/token/lookup-self.*")
		transform = regexp.MustCompile(".*/v1/int/au/transform/data/.*")
		metadata  = regexp.MustCompile(".*/computeMetadata/v1/instance/service-accounts/default/email.*")
	)

	switch {
	case login.MatchString(r.URL.Path):
		s.login(w, r)
	case lookup.MatchString(r.URL.Path):
		s.lookup(w, r)
	case transform.MatchString(r.URL.Path):
		s.transform(w, r)
	case metadata.MatchString(r.URL.Path):
		s.meta(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s StubServer) login(w http.ResponseWriter, r *http.Request) {
	resp := vault_external.Secret{
		Auth: &vault_external.SecretAuth{
			ClientToken: "vault-test-token",
			// 900 seconds = 15 minutes
			LeaseDuration: 900,
		},
	}
	respond(w, r, http.MethodPost, resp)
}

func (s StubServer) lookup(w http.ResponseWriter, r *http.Request) {
	resp := &vault_external.Secret{
		Data: map[string]interface{}{
			"meta": map[string]interface{}{
				"role": "gcpiamrole-fabric-encdec.common",
			},
		},
	}
	respond(w, r, http.MethodGet, resp)
}

func respond(w http.ResponseWriter, r *http.Request, method string, response interface{}) {
	if r.Method == method {
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s StubServer) transform(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer r.Body.Close()

	var req request
	_ = json.Unmarshal(bytes, &req) //nolint

	in := make([]string, 0, len(req.BatchInput))
	for _, token := range req.BatchInput {
		in = append(in, token.Value)
	}

	var batchResult []BatchResults
	switch {
	case strings.Contains(r.URL.String(), "decode"):
		batchResult = make([]BatchResults, 0, len(in))
		for _, result := range in {
			batchResult = append(batchResult, BatchResults{DecodedValue: result})
		}
	case strings.Contains(r.URL.String(), "encode"):
		batchResult = make([]BatchResults, 0, len(in))
		for _, result := range in {
			batchResult = append(batchResult, BatchResults{EncodedValue: result})
		}
	}
	secret := Secret{
		RequestID: gofakeit.UUID(),
		Data: Data{
			BatchResults: batchResult,
		},
	}

	resp, _ := json.Marshal(secret)
	_, _ = w.Write(resp)
}

type Secret struct {
	RequestID     string      `json:"request_id"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int         `json:"lease_duration"`
	Data          Data        `json:"data"`
	WrapInfo      interface{} `json:"wrap_info"`
	Warnings      interface{} `json:"warnings"`
	Auth          interface{} `json:"auth"`
}

type BatchResults struct {
	EncodedValue string `json:"encoded_value,omitempty"`
	DecodedValue string `json:"decoded_value,omitempty"`
}

type Data struct {
	BatchResults []BatchResults `json:"batch_results"`
}

type request struct {
	BatchInput []struct {
		Transformation string `json:"transformation"`
		Value          string `json:"value"`
	} `json:"batch_input"`
}

func (s StubServer) meta(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "fabric@anz.com")
}

type iamClient struct {
	*credentialspb.UnimplementedIAMCredentialsServer
}

func NewIAMServer() credentialspb.IAMCredentialsServer {
	return iamClient{}
}

func (c iamClient) SignJwt(context.Context, *credentialspb.SignJwtRequest) (*credentialspb.SignJwtResponse, error) {
	return &credentialspb.SignJwtResponse{
		KeyId:     "1234567890",
		SignedJwt: "gcp-signed-jwt-for-vault",
	}, nil
}

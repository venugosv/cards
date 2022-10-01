package v1beta1

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/anzx/fabric-cards/test/client/common"

	"github.com/anzx/fabric-cards/pkg/integration/vault"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
)

type RESTTestClient struct {
	httpClient *http.Client
	host       string
	headers    http.Header
	protocol   string
	ctx        context.Context
	state      common.ConnectionState
}

func NewRESTTestClient(ctx context.Context, host string, insecure bool, headers http.Header, vault vault.Client) *RESTTestClient {
	protocol := "https://"
	if insecure {
		protocol = "http://"
	}
	return &RESTTestClient{
		httpClient: http.DefaultClient,
		host:       host,
		headers:    headers,
		protocol:   protocol,
		ctx:        ctx,
		state: common.ConnectionState{
			Vault: vault,
		},
	}
}

func (r *RESTTestClient) LoadCard(t *testing.T) {
	listResponse, err := r.List()
	if err != nil {
		t.Fatalf("LoadCard: failed to ListCards: %v", err)
	}

	r.state.GetCard(listResponse.GetCards(), r.state.CurrentCard.GetTokenizedCardNumber())
}

func (r *RESTTestClient) GetCurrentCard() *cpb.Card {
	return r.state.CurrentCard
}

func (r *RESTTestClient) List() (*cpb.ListResponse, error) {
	url := fmt.Sprintf("%s%s/api/v1beta1/cards", r.protocol, r.host)

	b, err := common.Run(r.ctx, http.MethodGet, url, nil, r.headers)
	if err != nil {
		return nil, err
	}
	var resp cpb.ListResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) Can(eligibility epbv1beta1.Eligibility) bool {
	return r.state.Can(eligibility)
}

func (r *RESTTestClient) GetDetails() (*cpb.GetDetailsResponse, error) {
	url := fmt.Sprintf("%s%s/api/v1beta1/cards/%s", r.protocol, r.host, r.state.CurrentCard.GetTokenizedCardNumber())
	req := &cpb.GetDetailsRequest{
		TokenizedCardNumber: r.state.CurrentCard.GetTokenizedCardNumber(),
	}
	b, err := common.Run(r.ctx, http.MethodGet, url, req, r.headers)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var resp cpb.GetDetailsResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) Activate() (*cpb.ActivateResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta1/cards/%s/activate", r.protocol, r.host, tokenizedCardNumber)

	plaintext := tokenizedCardNumber
	if r.state.Vault != nil {
		var err error
		plaintext, err = r.state.Vault.DecodeCardNumber(r.ctx, tokenizedCardNumber)
		if err != nil {
			return nil, fmt.Errorf("unable to getRequest last 6 digits %v", err)
		}
	}
	req := &cpb.ActivateRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		Last_6Digits:        plaintext[len(plaintext)-6:],
	}

	b, err := common.Run(r.ctx, http.MethodPost, url, req, r.headers)
	if err != nil {
		return nil, err
	}

	var resp cpb.ActivateResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) GetWrappingKey() (*cpb.GetWrappingKeyResponse, error) {
	url := fmt.Sprintf("%s%s/api/v1beta1/cards/pin/wrappingkey", r.protocol, r.host)
	b, err := common.Run(r.ctx, http.MethodGet, url, nil, r.headers)
	if err != nil {
		return nil, err
	}

	var resp cpb.GetWrappingKeyResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) ResetPIN() (*cpb.ResetPINResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta1/cards/%s/pin/reset", r.protocol, r.host, tokenizedCardNumber)
	req := &cpb.ResetPINRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		EncryptedPinBlock:   pinBlock1,
	}

	b, err := common.Run(r.ctx, http.MethodPost, url, req, r.headers)
	if err != nil {
		return nil, err
	}

	var resp cpb.ResetPINResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) SetPIN() (*cpb.SetPINResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta1/cards/%s/pin/set", r.protocol, r.host, tokenizedCardNumber)
	req := &cpb.ResetPINRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		EncryptedPinBlock:   pinBlock1,
	}

	b, err := common.Run(r.ctx, http.MethodPost, url, req, r.headers)
	if err != nil {
		return nil, err
	}

	var resp cpb.SetPINResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) Replace(reason cpb.ReplaceRequest_Reason) (*cpb.ReplaceResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta1/cards/%s/replace", r.protocol, r.host, tokenizedCardNumber)
	req := &cpb.ReplaceRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		Reason:              reason,
	}

	b, err := common.Run(r.ctx, http.MethodPost, url, req, r.headers)
	if err != nil {
		return nil, err
	}
	var resp cpb.ReplaceResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) AuditTrail() (*cpb.AuditTrailResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta1/cards/%s/audittrail", r.protocol, r.host, tokenizedCardNumber)
	req := &cpb.AuditTrailRequest{
		TokenizedCardNumber: tokenizedCardNumber,
	}

	b, err := common.Run(r.ctx, http.MethodGet, url, req, r.headers)
	if err != nil {
		return nil, err
	}

	var resp cpb.AuditTrailResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTTestClient) CreateApplePaymentToken() (*cpb.CreateApplePaymentTokenResponse, error) {
	panic("implement me")
}

func (r *RESTTestClient) CreateGooglePaymentToken() (*cpb.CreateGooglePaymentTokenResponse, error) {
	panic("implement me")
}

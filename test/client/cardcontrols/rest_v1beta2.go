package cardcontrols

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/anzx/fabric-cards/test/client/common"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	ccpbv1beta2 "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
)

type RESTV1beta2Client struct {
	httpClient  *http.Client
	cardHost    string
	cardHeaders http.Header
	host        string
	headers     http.Header
	protocol    string
	ctx         context.Context
	state       common.ConnectionState
}

func NewRESTV1beta2Client(ctx context.Context, cardHost, host string, insecure bool, headers, cardsHeaders http.Header) *RESTV1beta2Client {
	protocol := "https://"
	if insecure {
		protocol = "http://"
	}
	return &RESTV1beta2Client{
		httpClient:  http.DefaultClient,
		cardHost:    cardHost,
		cardHeaders: cardsHeaders,
		host:        host,
		headers:     headers,
		protocol:    protocol,
		ctx:         ctx,
	}
}

func (r *RESTV1beta2Client) LoadCard(t *testing.T) {
	listResponse, err := r.List()
	if err != nil {
		t.Fatalf("LoadCard: failed to ListCards: %v", err)
	}

	r.state.GetCard(listResponse.GetCards(), r.state.CurrentCard.GetTokenizedCardNumber())
}

func (r *RESTV1beta2Client) GetCurrentCard() *cpb.Card {
	return r.state.CurrentCard
}

func (r *RESTV1beta2Client) Can(eligibility epbv1beta1.Eligibility) bool {
	return r.state.Can(eligibility)
}

func (r *RESTV1beta2Client) List() (*cpb.ListResponse, error) {
	url := fmt.Sprintf("%s%s/api/v1beta1/cards", r.protocol, r.cardHost)

	b, err := common.Run(r.ctx, http.MethodGet, url, nil, r.cardHeaders)
	if err != nil {
		return nil, err
	}
	var resp cpb.ListResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTV1beta2Client) ListControls() (*ccpbv1beta2.ListControlsResponse, error) {
	url := fmt.Sprintf("%s%s/api/v1beta2/cardcontrols", r.protocol, r.host)
	b, err := common.Run(r.ctx, http.MethodGet, url, nil, r.headers)
	if err != nil {
		return nil, err
	}
	var resp ccpbv1beta2.ListControlsResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTV1beta2Client) QueryControls() (*ccpbv1beta2.CardControlResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta2/cardcontrols/%s", r.protocol, r.host, tokenizedCardNumber)
	req := &ccpbv1beta2.QueryControlsRequest{
		TokenizedCardNumber: tokenizedCardNumber,
	}
	b, err := common.Run(r.ctx, http.MethodGet, url, req, r.headers)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var resp ccpbv1beta2.CardControlResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTV1beta2Client) SetControls(controls ...ccpbv1beta2.ControlType) (*ccpbv1beta2.CardControlResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta2/cardcontrols/%s/set", r.protocol, r.host, tokenizedCardNumber)
	var controlRequest []*ccpbv1beta2.ControlRequest
	for _, control := range controls {
		controlRequest = append(controlRequest, &ccpbv1beta2.ControlRequest{
			ControlType: control,
		})
	}
	req := &ccpbv1beta2.SetControlsRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		CardControls:        controlRequest,
	}
	b, err := common.Run(r.ctx, http.MethodPost, url, req, r.headers)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var resp ccpbv1beta2.CardControlResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTV1beta2Client) RemoveControls(controls ...ccpbv1beta2.ControlType) (*ccpbv1beta2.CardControlResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta2/cardcontrols/%s/remove", r.protocol, r.host, tokenizedCardNumber)
	req := &ccpbv1beta2.RemoveControlsRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		ControlTypes:        controls,
	}
	b, err := common.Run(r.ctx, http.MethodPost, url, req, r.headers)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var resp ccpbv1beta2.CardControlResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTV1beta2Client) TransferControls(newTokenizedCardNumber string) (*ccpbv1beta2.TransferControlsResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta2/cardcontrols/%s/transfer/%s", r.protocol, r.host, tokenizedCardNumber, newTokenizedCardNumber)
	req := &ccpbv1beta2.TransferControlsRequest{
		CurrentTokenizedCardNumber: tokenizedCardNumber,
		NewTokenizedCardNumber:     newTokenizedCardNumber,
	}
	b, err := common.Run(r.ctx, http.MethodPatch, url, req, r.headers)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var resp ccpbv1beta2.TransferControlsResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

func (r *RESTV1beta2Client) BlockCard(action ccpbv1beta2.BlockCardRequest_Action) (*ccpbv1beta2.BlockCardResponse, error) {
	tokenizedCardNumber := r.state.CurrentCard.GetTokenizedCardNumber()
	url := fmt.Sprintf("%s%s/api/v1beta2/cardcontrols/%s/%s", r.protocol, r.host, tokenizedCardNumber, action)
	req := &ccpbv1beta2.BlockCardRequest{
		TokenizedCardNumber: tokenizedCardNumber,
		Action:              action,
	}
	b, err := common.Run(r.ctx, http.MethodPatch, url, req, r.headers)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var resp ccpbv1beta2.BlockCardResponse
	err = protojson.Unmarshal(b, &resp)
	return &resp, err
}

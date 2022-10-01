package v1beta2

import (
	"context"
	"sync"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
)

const (
	listFailed        = "list controls failed"
	visaGatewayCreate = "https://fabric.anz.com/scopes/visaGateway:create"
	visaGatewayRead   = "https://fabric.anz.com/scopes/visaGateway:read"
	visaGatewayUpdate = "https://fabric.anz.com/scopes/visaGateway:update"
	visaGatewayDelete = "https://fabric.anz.com/scopes/visaGateway:delete"
)

func (s server) ListControls(ctx context.Context, _ *ccpb.ListControlsRequest) (*ccpb.ListControlsResponse, error) {
	entitledCards, err := s.Entitlements.ListEntitledCards(ctx)
	if err != nil {
		return nil, serviceErr(err, listFailed)
	}

	var tokenizedCardNumbers []string
	for _, entitledCard := range entitledCards {
		tokenizedCardNumbers = append(tokenizedCardNumbers, entitledCard.GetTokenizedCardNumber())
	}

	cardNumbers, err := s.Vault.DecodeCardNumbers(ctx, tokenizedCardNumbers)
	if err != nil {
		logf.Err(ctx, err)
		return nil, serviceErr(err, listFailed)
	}

	var visaCtx context.Context
	if feature.FeatureGate.Enabled(feature.FORGEROCK_SYSTEM_LOGIN) {
		visaCtx, err = s.Forgerock.SystemJWT(ctx, visaGatewayRead)
		if err != nil {
			return nil, serviceErr(err, listFailed)
		}
	} else {
		visaCtx = ctx
	}

	ch := make(chan []*ccpb.CardControlResponse)
	var wg sync.WaitGroup

	for _, entitledCard := range entitledCards {
		wg.Add(1)
		go s.asyncGetDoc(ctx, visaCtx, entitledCard.TokenizedCardNumber, cardNumbers, &wg, ch)
	}

	// close the channel in the background
	go func() {
		wg.Wait()
		close(ch)
	}()

	// read from channel as they come in until its closed
	responses := make([]*ccpb.CardControlResponse, 0)
	for call := range ch {
		responses = append(responses, call...)
	}

	return &ccpb.ListControlsResponse{
		CardControls: responses,
	}, nil
}

func (s server) asyncGetDoc(ctx context.Context, visaCtx context.Context, tokenizedCardNumber string, cardNumbers map[string]string, wg *sync.WaitGroup, ch chan<- []*ccpb.CardControlResponse) {
	defer wg.Done()

	out := make([]*ccpb.CardControlResponse, 0)

	if _, err := s.Entitlements.GetEntitledCard(ctx, tokenizedCardNumber, entitlements.OPERATION_CARDCONTROLS); err != nil {
		logf.Err(ctx, err)
		return
	}

	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_CARD_CONTROLS, tokenizedCardNumber); err != nil {
		logf.Err(ctx, err)
		return
	}

	if _, ok := cardNumbers[tokenizedCardNumber]; !ok {
		logf.Error(ctx, errors.New("unable to get plaintext card number"), "decoded %s not found", tokenizedCardNumber)
		return
	}

	// query controls by pan
	visaResponse, err := s.Visa.ListControlDocuments(visaCtx, cardNumbers[tokenizedCardNumber])
	if err != nil {
		logf.Err(ctx, err)
		control := &ccpb.CardControlResponse{
			TokenizedCardNumber: tokenizedCardNumber,
		}
		out = append(out, control)
		ch <- out

		return
	}

	control := getCardControlResponse(visaResponse, tokenizedCardNumber)
	out = append(out, control)

	ch <- out
}

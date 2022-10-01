package v1beta1

import (
	"context"
	"sync"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

const listFailed = "list controls failed"

func (s server) List(ctx context.Context, _ *ccpb.ListRequest) (*ccpb.ListResponse, error) {
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

	ch := make(chan map[string]*ccpb.CardControlResponse)
	var wg sync.WaitGroup

	for _, entitledCard := range entitledCards {
		wg.Add(1)
		go s.asyncGetDoc(ctx, ctx, entitledCard.TokenizedCardNumber, cardNumbers, &wg, ch)
	}

	// close the channel in the background
	go func() {
		wg.Wait()
		close(ch)
	}()

	// read from channel as they come in until its closed
	responses := make(map[string]*ccpb.CardControlResponse)
	for call := range ch {
		for tokenizedCardNumber, response := range call {
			responses[tokenizedCardNumber] = response
		}
	}

	return &ccpb.ListResponse{
		CardControls: responses,
	}, nil
}

func (s server) asyncGetDoc(ctx context.Context, jwt context.Context, tokenizedCardNumber string, cardNumbers map[string]string, wg *sync.WaitGroup, ch chan<- map[string]*ccpb.CardControlResponse) {
	defer wg.Done()
	res := make(map[string]*ccpb.CardControlResponse)

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
	visaResponse, err := s.Visa.QueryControls(jwt, cardNumbers[tokenizedCardNumber])
	if err != nil {
		logf.Err(ctx, err)
		res[tokenizedCardNumber] = &ccpb.CardControlResponse{}
		ch <- res
		return
	}

	res[tokenizedCardNumber] = getCardControlResponse(visaResponse)
	ch <- res
}

package cards

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/date"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

const listCardsFailed = "list cards failed"

type cardDetail struct {
	entitledCard      *entpb.EntitledCard
	debitCardResponse *ctm.DebitCardResponse
}

func (s server) List(ctx context.Context, _ *cpb.ListRequest) (*cpb.ListResponse, error) {
	entitledCards, err := s.Entitlements.ListEntitledCards(ctx)
	if err != nil {
		return nil, serviceErr(err, listCardsFailed)
	}

	cardDetails := make([]*cardDetail, 0, len(entitledCards))
	allCards := make(map[string]*ctm.DebitCardResponse, len(entitledCards))
	for _, card := range entitledCards {
		c := card
		tokenizedCardNumber := c.TokenizedCardNumber
		response, err := s.CTM.DebitCardInquiry(ctx, tokenizedCardNumber)
		if err != nil {
			return nil, serviceErr(err, listCardsFailed)
		}
		cardDetails = append(cardDetails, &cardDetail{
			entitledCard:      c,
			debitCardResponse: response,
		})
		allCards[tokenizedCardNumber] = response
	}

	var cards []*cpb.Card
	for _, card := range cardDetails {
		if !card.debitCardResponse.Visible() && newCardIsActive(card.debitCardResponse, allCards) {
			continue
		}

		d := card.debitCardResponse
		c := &cpb.Card{
			Name:                d.EmbossingLine1,
			TokenizedCardNumber: card.entitledCard.GetTokenizedCardNumber(),
			Last_4Digits:        d.CardNumber.Last4Digits,
			Status:              d.Status.String(),
			ExpiryDate:          date.GetDate(ctx, date.YYMM, d.ExpiryDate),
			AccountNumbers:      card.entitledCard.GetAccountNumbers(),
			Eligibilities:       d.Eligibility(),
			Wallets:             getWallet(d.Wallets),
			CardControlsEnabled: d.CardControlPreference,
		}

		if d.NewCardNumber != nil {
			c.NewTokenizedCardNumber = d.NewCardNumber.Token
		}

		cards = append(cards, c)
	}

	return &cpb.ListResponse{
		Cards: cards,
	}, nil
}

func getWallet(in ctm.Wallet) *cpb.Wallets {
	return &cpb.Wallets{
		Other:      in.Other,
		Fitness:    in.Fitness,
		ApplePay:   in.ApplePay,
		ECommerce:  in.ECommerce,
		SamsungPay: in.SamsungPay,
		GooglePay:  in.GooglePay,
	}
}

func newCardIsActive(card *ctm.DebitCardResponse, allCards map[string]*ctm.DebitCardResponse) bool {
	// if new card is issued
	if card.NewCardNumber != nil {
		// and new card is returned from CTM
		if c, ok := allCards[card.NewCardNumber.Token]; ok {
			// return its activation status
			return c.ActivationStatus
		}
	}
	// otherwise assume its inactive so that both cards are returned
	return false
}

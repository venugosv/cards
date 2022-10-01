package common

import (
	"github.com/anzx/fabric-cards/pkg/integration/vault"
	cpbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
)

type ConnectionState struct {
	CurrentCard *cpbv1beta1.Card
	Vault       vault.Client
}

func (c *ConnectionState) GetCard(cards []*cpbv1beta1.Card, target string) {
	if c.CurrentCard == nil {
		c.CurrentCard = cards[0]
		target = c.CurrentCard.GetTokenizedCardNumber()
	}
	if target == "" {
		return
	}
	for _, card := range cards {
		if card.GetTokenizedCardNumber() != target {
			continue
		}
		c.CurrentCard = card
		c.GetCard(cards, c.CurrentCard.GetNewTokenizedCardNumber())
		return
	}
}

func (c *ConnectionState) Can(eligibility epbv1beta1.Eligibility) bool {
	if c.CurrentCard == nil {
		return false
	}

	for _, e := range c.CurrentCard.Eligibilities {
		if e == eligibility {
			return true
		}
	}

	return false
}

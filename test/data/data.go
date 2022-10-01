package data

import (
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
)

const (
	defaultPersonaID        = "9045c12a-5d2c-5ebc-bc1a-64d1551b93ce"
	defaultCardNumber       = "4622390512341000"
	defaultCardToken        = "6688390512341000" // #nosec
	defaultActivationStatus = true
	defaultCardControls     = CardControlsPresetNoControls
	defaultCardStatus       = ctm.StatusIssued
	defaultPINChangedCount  = 1
	defaultAccountNumber    = "1234567890"
)

type Data struct {
	Users []*User
}

func (t Data) GetCardByToken(token string) (string, *Card) {
	for _, user := range t.Users {
		for _, card := range user.Cards {
			if card.Token == token {
				return user.PersonaID, card
			}
		}
	}
	return "", nil
}

func (t Data) GetCardByCardNumber(cardNumber string) *Card {
	for _, user := range t.Users {
		for _, card := range user.Cards {
			if card.CardNumber == cardNumber {
				return card
			}
		}
	}
	return nil
}

func (t Data) GetCardByTokenizedCardNumber(tokenizedCardNumber string) *Card {
	for _, user := range t.Users {
		for _, card := range user.Cards {
			if card.Token == tokenizedCardNumber {
				return card
			}
		}
	}
	return nil
}

func (t Data) GetUserByPersonaID(personaID string) *User {
	for _, user := range t.Users {
		if user.PersonaID == personaID {
			return user
		}
	}
	return nil
}

func (t Data) GetUserItemByCardNumber(cardNumber string, personaID string) *Card {
	user := t.GetUserByPersonaID(personaID)
	for _, card := range user.Cards {
		if card.CardNumber == cardNumber {
			return card
		}
	}
	return nil
}

func (t Data) GetUserItemByTokenizedCardNumber(tokenizedCardNumber string, personaID string) *Card {
	user := t.GetUserByPersonaID(personaID)
	for _, card := range user.Cards {
		if card.Token == tokenizedCardNumber {
			return card
		}
	}
	return nil
}

func (t Data) GetUserAndCardByTokenizedCardNumber(tokenizedCardNumber string) (*User, *Card) {
	for _, user := range t.Users {
		for _, card := range user.Cards {
			if card.Token == tokenizedCardNumber {
				return user, card
			}
		}
	}

	return nil, nil
}

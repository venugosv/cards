package vault

import (
	"context"

	"github.com/anzx/fabric-cards/test/data"
)

type StubClient struct {
	testingData *data.Data
	Err         error
}

func (v StubClient) EncodeCardNumber(ctx context.Context, cardNumber string) (string, error) {
	result, err := v.EncodeCardNumbers(ctx, []string{cardNumber})
	if err != nil {
		return "", err
	}
	return result[cardNumber], nil
}

func (v StubClient) EncodeCardNumbers(ctx context.Context, cardNumbers []string) (map[string]string, error) {
	if v.Err != nil {
		return nil, v.Err
	}
	cardsToToken := make(map[string]string)

	for _, user := range v.testingData.Users {
		for _, card := range user.Cards {
			cardsToToken[card.CardNumber] = card.Token
		}
	}

	result := make(map[string]string)
	for _, cardNumber := range cardNumbers {
		result[cardNumber] = cardsToToken[cardNumber]
	}
	return result, nil
}

func (v StubClient) DecodeCardNumbers(ctx context.Context, tokens []string) (map[string]string, error) {
	if v.Err != nil {
		return nil, v.Err
	}
	tokenToCards := make(map[string]string)

	for _, user := range v.testingData.Users {
		for _, card := range user.Cards {
			tokenToCards[card.Token] = card.CardNumber
		}
	}

	result := make(map[string]string)
	for _, token := range tokens {
		result[token] = tokenToCards[token]
	}
	return result, nil
}

func NewVaultClient(testingData *data.Data) StubClient {
	return StubClient{
		testingData: testingData,
	}
}

func (v StubClient) DecodeCardNumber(ctx context.Context, token string) (string, error) {
	result, err := v.DecodeCardNumbers(ctx, []string{token})
	if err != nil {
		return "", err
	}
	return result[token], nil
}

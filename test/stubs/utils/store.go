package utils

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/anzx/anzdata"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
)

// Store is a singleton instance that hold the data that can be shared across all stubbed services
type Store struct {
	cache              *Cache
	simulatedLatencies SimulatedLatencies
}

// GetSimulatedLatencies Get the simulated latencies data
func (s *Store) GetSimulatedLatencies(r *http.Request) int64 {
	for k, v := range s.simulatedLatencies {
		match, _ := regexp.MatchString(k, r.URL.Path)
		if match {
			return v
		}
	}

	return 0
}

// SaveSimulatedLatencies Save the simulated latencies data
func (s *Store) SaveSimulatedLatencies(delays SimulatedLatencies) {
	s.simulatedLatencies = delays
}

func (s *Store) GetCard(token string) *ctm.DebitCardResponse {
	key := fmt.Sprintf("%s-detail", token)
	for _, val := range s.cache.Dump() {
		for _, card := range val.data.Cards {
			if card.Token != token {
				continue
			}
			if detail, ok := val.data.Other.(map[string]interface{}); ok {
				if out, ok := detail[key]; ok {
					return out.(*ctm.DebitCardResponse)
				}
			}
		}
	}
	card := defaultDebitCardResponse(token)
	s.SaveCard(token, card)
	return card
}

func (s *Store) SaveCard(token string, card *ctm.DebitCardResponse) {
	key := fmt.Sprintf("%s-detail", token)
	for k, v := range s.cache.Dump() {
		for _, c := range v.data.Cards {
			if c.Token != token {
				continue
			}
			if detail, ok := v.data.Other.(map[string]interface{}); ok {
				detail[key] = card
				v.data.Other = detail
			} else {
				v.data.Other = map[string]interface{}{key: card}
			}
			s.cache.Set(k, v.data)
		}
	}
}

func (s *Store) GetUser(personaID string) anzdata.User {
	if cachedUser := s.cache.Get(personaID); cachedUser != nil {
		return cachedUser.(anzdata.User)
	}
	newUser, err := anzdata.AllUsers().Match("PersonaID", personaID)
	if err != nil {
		newUser = anzdata.MustRandomUserFromUUID(personaID)
	}
	s.SaveUser(personaID, newUser)
	return newUser
}

func (s *Store) GetUserByOCV(id string) anzdata.User {
	data := s.cache.Dump()
	for _, cacheValue := range data {
		if cacheValue.data.OCVID == id {
			return cacheValue.data
		}
	}
	return anzdata.User{}
}

func (s *Store) SaveUser(personaID string, user anzdata.User) {
	s.cache.Set(personaID, user)
}

func defaultDebitCardResponse(token string) *ctm.DebitCardResponse {
	now := time.Now().AddDate(0, 0, -1)
	year, month, day := now.Date()
	exp := now.AddDate(5, 0, 0)
	expYear := fmt.Sprintf("%d", exp.Year())
	expMonth := exp.Month()
	yyyymmdd := "%d-%02d-%02d"
	person := gofakeit.Person()
	return &ctm.DebitCardResponse{
		Title:               "MX",
		FirstName:           person.FirstName,
		LastName:            person.LastName,
		ProductCode:         "PDV",
		SubProductCode:      "101",
		StatusCode:          ctm.StatusCodeIssued,
		Status:              ctm.StatusIssued,
		AccountsLinkedCount: 1,
		ExpiryDate:          fmt.Sprintf("%s%02d", expYear[:2], expMonth),
		ActivationStatus:    false,
		EmbossingLine1:      fmt.Sprintf("%s %s", person.FirstName, person.LastName),
		DispatchedMethod:    ctm.DispatchedMethodMail,
		IssueBranch:         4111,
		IssueReason:         "New",
		CollectionStatus:    ctm.CollectionStatusCollected,
		IssueDate:           fmt.Sprintf(yyyymmdd, year, month, day),
		CardNumber: ctm.Card{
			Token:       token,
			Last4Digits: fmt.Sprintf("%d", gofakeit.Number(1000, 9999)),
		},
		PinChangedCount: 0,
	}
}

var (
	once     sync.Once
	instance *Store
)

// A singleton that holds testing data
// x-request-id is used as the unique key to retrieve the testing data, so we can share testing data between stubs as long as they have same x-request-id
func GetStore(ctx context.Context) *Store {
	once.Do(func() {
		instance = &Store{
			// Using a FIFO queue to make sure the stubs do not consume too much memory
			cache:              NewCache(ctx, 120),
			simulatedLatencies: map[string]int64{},
		}
	})
	return instance
}

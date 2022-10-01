package data

import (
	"fmt"
	"math/rand"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/util"
)

type Card struct {
	CardNumber       string
	Token            string
	Status           ctm.Status
	ActivationStatus bool
	CardControls     CardControlsPresetType
	Reason           *ctm.StatusReason
	NewCardNumber    *string
	NewToken         *string
	PinChangedCount  int64
	AccountNumbers   []string
}

type CardControlsPresetType string

const (
	CardControlsPresetAllControls        CardControlsPresetType = "All Controls"
	CardControlsPresetGlobalControls     CardControlsPresetType = "Global Controls"
	CardControlsPresetContactlessControl CardControlsPresetType = "Contactless Control"
	CardControlsPresetNoControls         CardControlsPresetType = "No Controls"
	CardControlsPresetNotEnrolled        CardControlsPresetType = "Not Enrolled"
	CardControlsPresetCanNotBeEnrolled   CardControlsPresetType = "Can not be Enrolled"
)

func aCard(cardBuilders ...func(*Card)) *Card {
	card := &Card{
		CardNumber:       defaultCardNumber,
		Token:            defaultCardToken,
		Status:           defaultCardStatus,
		ActivationStatus: defaultActivationStatus,
		CardControls:     defaultCardControls,
		PinChangedCount:  defaultPINChangedCount,
		AccountNumbers:   []string{defaultAccountNumber},
	}

	for _, build := range cardBuilders {
		build(card)
	}

	return card
}

func WithACardNumber(cardNumber string) func(u *Card) {
	return func(u *Card) {
		u.CardNumber = cardNumber
	}
}

func WithAToken(token string) func(u *Card) {
	return func(u *Card) {
		u.Token = token
	}
}

func WithStatus(status ctm.Status) func(u *Card) {
	return func(u *Card) {
		u.Status = status
	}
}

func WithStatusReason(reason ctm.StatusReason) func(u *Card) {
	return func(u *Card) {
		u.Reason = &reason
	}
}

func WithControls(controlsPresetType CardControlsPresetType) func(u *Card) {
	return func(u *Card) {
		u.CardControls = controlsPresetType
	}
}

func WithNewCardNumber(cardNumber string) func(u *Card) {
	return func(u *Card) {
		u.NewCardNumber = util.ToStringPtr(cardNumber)
	}
}

func WithNewCard(cardNumber, tokenizedCardNumber string) func(u *Card) {
	return func(u *Card) {
		u.NewCardNumber = util.ToStringPtr(cardNumber)
		u.NewToken = util.ToStringPtr(tokenizedCardNumber)
	}
}

func WithPINChangeCount(count int64) func(u *Card) {
	return func(u *Card) {
		u.PinChangedCount = count
	}
}

func WithAccountNumbers(accountNumbers ...string) func(u *Card) {
	return func(u *Card) {
		u.AccountNumbers = accountNumbers
	}
}

func (c Card) AddAccountNumbers(accountNumbers ...string) *Card {
	c.AccountNumbers = append(c.AccountNumbers, accountNumbers...)
	return &c
}

func Active(c *Card) {
	c.ActivationStatus = true
}

func Inactive(c *Card) {
	c.ActivationStatus = false
}

func RandomCardNumber() string {
	last4digits := func(low, hi int) int {
		return low + rand.Intn(hi-low) //nolint:gosec
	}(1000, 9999)
	return fmt.Sprintf("456462102917%d", last4digits)
}

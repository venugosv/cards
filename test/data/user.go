package data

import (
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/google/uuid"
)

type User struct {
	PersonaID string
	Cards     []*Card
}

func AUser(userBuilders ...func(*User)) *User {
	user := &User{
		PersonaID: defaultPersonaID,
	}

	for _, build := range userBuilders {
		build(user)
	}

	return user
}

// Function builders
func WithAPersonaID(personaID string) func(u *User) {
	return func(u *User) {
		u.PersonaID = personaID
	}
}

func WithACard(builders ...func(*Card)) func(u *User) {
	return func(u *User) {
		u.Cards = append(u.Cards, aCard(builders...))
	}
}

func AUserWithACard(builders ...func(*Card)) *User {
	return AUser(WithACard(builders...))
}

func (u User) AddACard(builders ...func(*Card)) *User {
	u.Cards = append(u.Cards, aCard(builders...))
	return &u
}

func (u User) CardNumber() string {
	if len(u.Cards) > 0 {
		return u.Cards[0].CardNumber
	}
	return ""
}

func (u User) Token() string {
	if len(u.Cards) > 0 {
		return u.Cards[0].Token
	}
	return ""
}

func (u User) GetCard(token string) *Card {
	for _, card := range u.Cards {
		if card.Token == token {
			return card
		}
	}
	return nil
}

func (u *User) ReplaceCard(token string, status ctm.PlasticType) *Card {
	oldCard := u.GetCard(token)
	if status == ctm.SameNumber {
		return oldCard
	}

	oldCard.Status = ctm.StatusStolen
	newCard := aCard(RandomCard(ctm.StatusIssued)...)
	u.Cards = append(u.Cards, newCard)

	return newCard
}

func DefaultUser() *User {
	return AUser(
		WithAPersonaID(defaultPersonaID),
		WithACard(
			Active,
			WithACardNumber(defaultCardNumber),
			WithAToken(defaultCardToken),
			WithStatus(ctm.StatusIssued),
			WithControls(CardControlsPresetNotEnrolled),
			WithPINChangeCount(1),
			WithAccountNumbers(defaultAccountNumber),
		),
	)
}

func RandomCard(status ctm.Status) []func(*Card) {
	cardNumber := RandomCardNumber()
	token := RandomCardNumber()

	return []func(*Card){
		Active,
		WithACardNumber(cardNumber),
		WithAToken(token),
		WithStatus(status),
		WithControls(CardControlsPresetNotEnrolled),
		WithPINChangeCount(1),
		WithAccountNumbers(defaultAccountNumber),
	}
}

func RandomUser() *User {
	return AUser(
		WithAPersonaID(uuid.New().String()),
	).AddACard(RandomCard(ctm.StatusIssued)...)
}

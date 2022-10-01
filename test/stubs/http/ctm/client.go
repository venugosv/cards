package ctm

import (
	"context"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/identity"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/test/data"
)

type StubClient struct {
	testingData        *data.Data
	ActivateError      error
	InquiryError       error
	InquiryErrorFunc   func() error
	ReplaceError       error
	UpdateError        error
	SetPreferenceError error
	updateDetailsError error
	PINInfoUpdateError error
}

// NewStubClient creates a ctm client stubs
func NewStubClient(testData *data.Data) StubClient {
	return StubClient{
		testingData: testData,
	}
}

func (m StubClient) ReplaceCard(ctx context.Context, req *ctm.ReplaceCardRequest, tokenizedCardNumber string) (string, error) {
	if m.ReplaceError != nil {
		return "", m.ReplaceError
	}

	user := m.testingData.GetUserByPersonaID(GetPersonaID(ctx))
	newCard := user.ReplaceCard(tokenizedCardNumber, req.PlasticType)

	return newCard.Token, nil
}

func (m StubClient) UpdatePreferences(ctx context.Context, _ *ctm.UpdatePreferencesRequest, tokenizedCardNumber string) (bool, error) {
	if m.SetPreferenceError != nil {
		return false, m.SetPreferenceError
	}

	item := m.testingData.GetCardByTokenizedCardNumber(tokenizedCardNumber)
	if item.CardControls == data.CardControlsPresetCanNotBeEnrolled {
		return false, nil
	}

	return true, nil
}

func (m StubClient) UpdateDetails(ctx context.Context, req *ctm.UpdateDetailsRequest, tokenizedCardNumber string) (bool, error) {
	if m.updateDetailsError != nil {
		return false, m.updateDetailsError
	}

	return true, nil
}

func cardControlsPresent(item data.Card) bool {
	return item.CardControls == data.CardControlsPresetAllControls ||
		item.CardControls == data.CardControlsPresetGlobalControls ||
		item.CardControls == data.CardControlsPresetContactlessControl
}

func (m StubClient) DebitCardInquiry(ctx context.Context, req string) (*ctm.DebitCardResponse, error) {
	if m.InquiryError != nil {
		return nil, m.InquiryError
	}
	if m.InquiryErrorFunc != nil {
		if err := m.InquiryErrorFunc(); err != nil {
			return nil, err
		}
	}
	return GetCardDetails(m.testingData, req, GetPersonaID(ctx))
}

func (m StubClient) Activate(_ context.Context, tokenizedCardNumber string) (bool, error) {
	if m.ActivateError != nil {
		return false, m.ActivateError
	}

	card := m.testingData.GetCardByTokenizedCardNumber(tokenizedCardNumber)
	card.ActivationStatus = true
	return true, nil
}

func (m StubClient) UpdateStatus(_ context.Context, tokenizedCardNumber string, status ctm.Status) (bool, error) {
	if m.UpdateError != nil {
		return false, m.UpdateError
	}

	card := m.testingData.GetCardByTokenizedCardNumber(tokenizedCardNumber)
	card.Status = status
	return true, nil
}

func (m StubClient) UpdatePINInfo(_ context.Context, tokenizedCardNumber string) (bool, error) {
	if m.PINInfoUpdateError != nil {
		return false, m.PINInfoUpdateError
	}

	card := m.testingData.GetCardByTokenizedCardNumber(tokenizedCardNumber)
	card.PinChangedCount++
	return true, nil
}

func GetPersonaID(ctx context.Context) string {
	user, err := identity.Get(ctx)
	if err != nil {
		return ""
	}
	return user.PersonaID
}

func GetCardDetails(testingData *data.Data, tokenizedCardNumber string, userID string) (*ctm.DebitCardResponse, error) {
	var item *data.Card
	if userID != "" {
		item = testingData.GetUserItemByTokenizedCardNumber(tokenizedCardNumber, userID)
	} else {
		item = testingData.GetCardByTokenizedCardNumber(tokenizedCardNumber)
	}
	if item == nil {
		return nil, anzerrors.New(codes.NotFound, "ctm failed", anzerrors.NewErrorInfo(context.Background(), anzcodes.DatabaseFailure, "unable to find card"))
	}

	// Need a fresh copy of item to construct DebitCardResponses because its properties are using pointers.
	dataItem := *item

	response := &ctm.DebitCardResponse{
		AccountsLinkedCount: 2,
		CollectionBranch:    4672,
		CollectionStatus:    "Card NOT Collected",
		DetailsChangedDate:  "2015-08-05",
		DispatchedMethod:    "Sent to Branch",
		EmbossingLine1:      "MR NATHAN FUKUSHIMA",
		EmbossingLine2:      "MR NATHAN FUKUSHIMA",
		FirstName:           "NATHAN",
		LastName:            "FUKUSHIMA",
		ExpiryDate:          "1705",
		IssueBranch:         4672,
		IssueReason:         "New",
		IssueDate:           "2015-08-05",
		Limits: []ctm.NewLimits{
			{
				DailyLimit:          1000,
				DailyLimitAvailable: 1000,
				Type:                ctm.LimitTypeAPO,
			},
			{
				DailyLimit:          2500,
				DailyLimitAvailable: 2347,
				LastTransaction:     "2015-08-05",
				Type:                ctm.LimitTypeATMEFTPOS,
			},
		},
		MerchantUpdatePreference: true,
		PinChangeDate:            "2015-08-05",
		PinFailedCount:           0,
		ProductCode:              "PDV",
		ReplacementCount:         0,
		StatusChangedDate:        "2015-08-05",
		StatusChangedUserID:      "AVRSR",
		SubProductCode:           "001",
		Title:                    "MR",
		TotalCards:               1,
		Wallets: ctm.Wallet{
			ApplePay:   2,
			GooglePay:  1,
			SamsungPay: 0,
			Fitness:    0,
			ECommerce:  0,
			Other:      0,
		},
	}

	response.CardNumber = ctm.Card{
		Token:       dataItem.Token,
		Last4Digits: dataItem.CardNumber[len(dataItem.CardNumber)-4:],
	}

	if dataItem.NewCardNumber != nil && dataItem.NewToken != nil {
		newToken := *dataItem.NewToken
		newCardNumber := *dataItem.NewCardNumber

		var last4Digits string
		if len(newCardNumber) > 4 {
			last4Digits = newCardNumber[len(newCardNumber):]
		} else {
			last4Digits = newCardNumber
		}

		response.NewCardNumber = &ctm.Card{
			Token:       newToken,
			Last4Digits: last4Digits,
		}
	}

	response.ActivationStatus = dataItem.ActivationStatus
	response.CardControlPreference = cardControlsPresent(dataItem)
	response.PinChangedCount = dataItem.PinChangedCount
	response.Status = dataItem.Status
	if dataItem.Reason != nil {
		response.StatusReason = *dataItem.Reason
	}

	return response, nil
}

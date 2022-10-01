package ctm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/util/apic"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"
)

const (
	cardPreference    = "SetPreference"
	cardUpdateDetails = "UpdateDetails"
	cardReplace       = "Replacement"
)

// First option will have a new plastic dispatched with new card details (in the case of a Lost or Stolen card), and the
// second will have a new plastic dispatched with the same details as the previous card (in the case of a damaged card).
type PlasticType string

const (
	NewNumber  PlasticType = "New Number"
	SameNumber PlasticType = "Same Number"
)

type DispatchedMethod string

const (
	DispatchedMethodMail   DispatchedMethod = "Mail"
	DispatchedMethodBranch DispatchedMethod = "Sent to Branch"
)

type ReplaceCardRequest struct {
	PlasticType              PlasticType      `json:"plasticType"`
	FirstName                string           `json:"firstName"`
	LastName                 string           `json:"lastName"`
	EmbossingLine1           string           `json:"embossingLine1"`
	EmbossingLine2           string           `json:"embossingLine2"`
	DispatchedMethod         DispatchedMethod `json:"despatchMethod"`
	CollectionBranch         string           `json:"collectionBranch"`
	DesignCode               string           `json:"designCode"`
	MerchantUpdatePreference bool             `json:"merchantUpdatePreference"`
	MailingAddress           MailingAddress   `json:"mailingAddress"`
	ContactInformation       ContactInfo      `json:"contactInformation"`
}

type MailingAddress struct {
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	AddressLine3 string `json:"addressLine3"`
	PostCode     string `json:"postCode"`
	CountryCode  string `json:"countryCode,omitempty"`
	CountryName  string `json:"countryName,omitempty"`
}

type ContactInfo struct {
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
}

type ReplaceCardResponse struct {
	CardNumber CardNumber `json:"cardNumber"`
}

type CardNumber struct {
	Token       string `json:"token"`
	Last4Digits string `json:"last4digits"`
}

// CardControlPreference - When set to true, it indicates customer has enabled visa card control to better manage their
// finance. This field will also be included in CTM Non-Monetary events published to CIM, so that it can be passed on to
// CAM and Falcon. Updates to the indicator will also be logged by CTM to be passed to Base24 every 15 minutes in the
// ‘trickle feed’ process."
// MerchantUpdatePreference - When this field is set to true (i.e. Opted-In), it means the customer is allowing Visa to
// pass on their updated card number and expiry date information to participating merchants, so any payments set up from
// this card at registered Merchants are not interrupted. Recurring bills will continue, without the need for the customer
// to advise them. The customer however, has the option to ‘Opt-out’, by contacting ANZ, where-by this new indicator will
// need to be updated to false.
type UpdatePreferencesRequest struct {
	CardControlPreference    *bool `json:"cardControlPreference,omitempty"`
	MerchantUpdatePreference *bool `json:"merchantUpdatePreference,omitempty"`
}

// **Replace a debit card with new or old card details**
// 1. Ordering a replacement card with old details, i.e. same PAN as the current card, can only be done if the card is
// in a Issued status. In this case, the plasticType should be set to Same Number, and you are able to update any of the
// other details on the card before it is despatched. A possible use-case for this is if a card is physically damaged,
// and a customer would like to request a replacement.
// 2. Ordering a replacement card with new details, i.e. different PAN from the current card, can only be done if a card
// is in a Lost or Stolen status. In this case, the plasticType should be set to New Number, and you are not able to
// update any other details on the card before it is despatched. The only use-case for this is when a card has either
// been Lost, Stolen, or the card details compromised, allowing a customer to have a replacement card with new details sent.
func (c client) ReplaceCard(ctx context.Context, req *ReplaceCardRequest, tokenizedCardNumber string) (string, error) {
	replaceCardURL := fmt.Sprintf(replaceAPIUrlTemplate, c.baseURL, tokenizedCardNumber)

	body, _ := json.Marshal(req)

	response, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, replaceCardURL, body), fmt.Sprintf("ctm:%s", cardReplace))
	if err != nil {
		return "", err
	}

	var out ReplaceCardResponse
	if err := json.Unmarshal(response, &out); err != nil {
		logf.Error(ctx, err, "debit card maintenance failed unexpected response from downstream")
		return "", anzerrors.Wrap(err, codes.Internal, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}

	if req.PlasticType == SameNumber {
		out.CardNumber.Token = tokenizedCardNumber
	}

	return out.CardNumber.Token, nil
}

// Update preferences on a debit card such as auto merchant updates or card control.
//
// - The merchant update flag is used to help maintain the continuity of the payment experience in a variety of
// reissuance situations. Currently, all debit and credit cards are ‘Opted-Out’ for VAU (Visa Account Updater) updates.
// Scheme debit cards however, are “Opted In”. To ‘Opt-In’ for VAU updates, means the customer allows Visa to pass on
// their updated card number and expiry date information to participating merchants, so any payments set up from this
// card at registered Merchants are not interrupted. Recurring bills will continue, without the need for the customer to
// advise them. The customer however, has the option to ‘Opt-out’, by contacting ANZ, where-by this flag will need to be
// updated to false.
// - The card controls indicator can be updated for Visa Debit Cards and Access Debit Cards. Updates to this flag will
// be logged by CTM to be passed to Base24 every 15 minutes in the ‘trickle feed’ process. Base24 will maintain this on
// their CRDD database, which will be used to decide if a card transactions needs to be authorised at Visa as well as CTM.
// - Mailing Address fields & Postcode fields are now Mandatory fields for a Card Replacement from this version onwards.
// - New field Card Withdrawal limit, can be updated using this version, however Plastic type must equal Spaces.
// You cannot update the card withdrawal limit and order a card replacement at the same time.
func (c client) UpdatePreferences(ctx context.Context, req *UpdatePreferencesRequest, tokenizedCardNumber string) (bool, error) {
	setPreferenceURL := fmt.Sprintf(preferenceAPIUrlTemplate, c.baseURL, tokenizedCardNumber)

	body, _ := json.Marshal(req)

	resp, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPatch, setPreferenceURL, body), fmt.Sprintf("ctm:%s", cardPreference))
	// In case the card was already enrolled, return true
	if bytes.Contains(resp, []byte("NO CHANGES MADE OR DETECTED")) {
		err = nil
	}

	return err == nil, err
}

type UpdateDetailsRequest struct {
	Title           string `json:"title"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	EmbossingLine1  string `json:"embossingLine1"`
	EmbossingLine2  string `json:"embossingLine2"`
	WithdrawalLimit string `json:"withdrawalLimit"`
}

// This operation is used to only update (maintain) the personal details surrounding a card in CTM. These details can
// only be updated if the card is in Issued status. Personal details on the card cannot be edited once a card status is
// set to Lost or Stolen status. To update preferences, or to request a replacement card, refer to the other operations
// in this API.
func (c client) UpdateDetails(ctx context.Context, req *UpdateDetailsRequest, tokenizedCardNumber string) (bool, error) {
	updateDetailsURL := fmt.Sprintf(updateDetailsAPIUrlTemplate, c.baseURL, tokenizedCardNumber)

	body, _ := json.Marshal(req)

	_, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPatch, updateDetailsURL, body), fmt.Sprintf("ctm:%s", cardUpdateDetails))

	return err == nil, err
}

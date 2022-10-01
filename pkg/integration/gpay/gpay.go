package gpay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anzx/fabric-cards/pkg/date"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
)

const (
	clientAppID = "lotus"
	intent      = "PUSH_PROV_MOBILE"
	provider    = "40010075001"
	iDnV        = "true"
)

type Payload struct {
	// PAN of the card to be enrolled and provisioned. Required Size: 13-19
	AccountNumber string `json:"accountNumber"`
	// CVV2 value associated with the PAN on the card. Optional Size: 3-4
	CVV2 string `json:"cvv2,omitempty"`
	// The full name on the Visa card associated with the enrolled payment instrument. Optional Size: 0-256
	Name string `json:"name"`
	// Payment instrument's expiration date. See section below. Required
	ExpirationDate ExpirationDate `json:"expirationDate"`
	// Billing address associated with the payment instrument. Optional
	BillingAddress BillingAddress `json:"billingAddress"`
	// Described in-depth in the Payment Instrument Provider. Required
	Provider Provider `json:"provider"`
}

type ExpirationDate struct {
	// The month that the Visa card is set to expire. Required Size: 2
	Month string `json:"month"`
	// The year that the Visa card is set to expire. Required Size: 4
	Year string `json:"year"`
}

type BillingAddress struct {
	// If the Issuer Country = US, UK or Canada, this field is "Conditional".
	// ROW (Rest of world) Issuer Country, this field is “Optional”.
	// Street 1 on billing address for the payment instrument.
	// Permitted characters: Whitespace, a-z, A-Z, 0-9, Symbols: .,'-_#:/
	// Conditional / Optional Size: 1-140
	Line1 string `json:"line1,omitempty"`
	Line2 string `json:"line2,omitempty"`
	Line3 string `json:"line3,omitempty"`
	// If the Issuer Country = US, UK or Canada, this field is "Conditional".
	// ROW (Rest of world) Issuer Country, this field is “Optional”.
	// The city associated with the enrolled payment instrument.
	// Permitted characters: Whitespace, a-z, A-Z, 0-9, Symbols: .'-
	// Conditional / Optional Size: 1-100
	City string `json:"city,omitempty"`
	// If the Issuer Country = US, UK or Canada, this field is “Conditional”.
	// ROW (Rest of world) Issuer Country, this field is “Optional”.
	// State or province code.
	// State or province code in ISO 3166-2 format, eg "NY".
	// Conditional / Optional Size: 3
	State string `json:"state,omitempty"`
	// If the Issuer Country = US, UK or Canada, this field is “Conditional”.
	// ROW (Rest of world) Issuer Country, this field is “Optional”.
	// Country code (e.g. “US”).
	// Country in ISO 3166-1 alpha-2 format, eg "US".
	// Conditional / Optional Size: 2
	Country string `json:"country,omitempty"`
	// If the Issuer Country = US, UK or Canada, this field is “Conditional”.
	// ROW (Rest of world) Issuer Country, this field is “Optional”.
	// The postal code associated with the enrolled payment instrument.
	// Permitted characters:  • A-Z • a-z • 0-9
	// Conditional / Optional Size: 3-16
	PostalCode string `json:"postalCode,omitempty"`
}

type Provider struct {
	// The intent of the encryptor; what is the encryptor of the data trying to do?
	// **PUSH_PROV_MOBILE** - The value PUSH_PROV_MOBILE means the issuer is providing the PAN for the purpose of
	// provisioning a token for the consumer on a particular device, for a particular wallet/account.
	// **PUSH_PROV_ONFILE** - The value PUSH_PROV_ONFILE means the issuer is providing the PAN for the purpose of
	// provisioning a token to be stored on file(cloud bound and not device bound) for ecommerce transactions,
	// for a particular wallet/account.
	Intent string `json:"intent"`
	// Client Wallet Provider is the token requestor’s ID (TRID), which is returned to the WP as part of onboarding.
	ClientWalletProvider string `json:"clientWalletProvider"`
	// Client-provided consumer ID that identifies the Wallet Account Holder entity.
	// It must match the value TWP will send in the token provision request.
	ClientWalletAccountID string `json:"clientWalletAccountID"`
	// Stable device identification set by Wallet Provider. Could be computer identifier or ID
	// tied to hardware such as TEE_ID or SE_ID.
	// − This field must match the clientDeviceID TWP will send in token provision request
	// − Required if intent is “PUSH_PROV_MOBILE”.
	ClientDeviceID string `json:"clientDeviceID"`
	// Unique identifier for the client application, used to provide some of the encrypted values.
	// Required if intent is “PUSH_PROV_MOBILE”.
	ClientAppID string `json:"clientAppID"`
	// String field to specify if the Issuer wants ID&V to be performed.
	// If the value is “false” or missing then Issuer will not receive 0100 TAR or 0100 AV, and
	// no step up will be triggered during provision. Permitted values - “true” or “false”.
	IsIDnV string `json:"isIDnV"`
}

func NewPayload(ctx context.Context, card *ctm.DebitCardResponse, cardNumber string, address *sspb.Address, stableHardwareID string, walletID string) ([]byte, error) {
	if card == nil || card.ExpiryDate == "" {
		return nil, errors.New("no expiry provided")
	}
	expiry := date.GetDate(ctx, date.YYMM, card.ExpiryDate)

	if cardNumber == "" {
		return nil, errors.New("plaintext card number not provided")
	}

	if address == nil {
		return nil, errors.New("no address provided")
	}

	if stableHardwareID == "" {
		return nil, errors.New("no device hardware ID provided")
	}

	payload := &Payload{
		AccountNumber: cardNumber,
		Name:          card.EmbossingLine1,
		ExpirationDate: ExpirationDate{
			Month: fmt.Sprintf("%02d", expiry.Month.Value),
			Year:  fmt.Sprintf("%d", expiry.Year.Value),
		},
		BillingAddress: BillingAddress{
			Line1:      address.GetLineOne(),
			Line2:      address.GetLineTwo(),
			Line3:      address.GetLineThree(),
			City:       address.GetCity(),
			State:      address.GetState(),
			Country:    convertCountry(ctx, address.GetCountry()),
			PostalCode: address.GetPostalCode(),
		},
		Provider: Provider{
			Intent:                intent,
			ClientWalletProvider:  provider,
			ClientWalletAccountID: walletID,
			ClientDeviceID:        stableHardwareID,
			ClientAppID:           clientAppID,
			IsIDnV:                iDnV,
		},
	}
	return json.Marshal(payload)
}

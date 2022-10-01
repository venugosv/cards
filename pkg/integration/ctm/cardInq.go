package ctm

import (
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

type Limits struct {
	DailyLimit          *int64  `json:"dailyLimit,omitempty"`
	DailyLimitAvailable *int64  `json:"dailyLimitAvailable,omitempty"`
	LastTransaction     *string `json:"lastTransaction,omitempty"`
	Type                *string `json:"type,omitempty"`
}

type Status string

const (
	StatusClosed                Status = "Closed"
	StatusDelinquentReturn      Status = "Delinquent (Return Card)"
	StatusDelinquentRetain      Status = "Delinquent (Retain Card)"
	StatusIssued                Status = "Issued"
	StatusLost                  Status = "Lost"
	StatusStolen                Status = "Stolen"
	StatusUnissuedNdIciCards    Status = "Unissued (N&D ICI Cards)"
	StatusTemporaryBlock        Status = "Temporary Block"
	StatusBlockAtm              Status = "Block ATM"
	StatusBlockAtmPosExcludeCnp Status = "Block ATM & POS (Exclude CNP)"
	StatusBlockAtmPosCnpBch     Status = "Block ATM, POS, CNP & BCH"
	StatusBlockAtmPosCnp        Status = "Block ATM, POS & CNP"
	StatusBlockCnp              Status = "Block CNP"
	StatusBlockPosExcludeCnp    Status = "Block POS (exclude CNP)"
)

func (s Status) String() string {
	return string(s)
}

type StatusReason string

const (
	StatusReasonWithPinOrAccountRelated      StatusReason = "With PIN or Account Related"
	StatusReasonWithoutPin                   StatusReason = "Without PIN"
	StatusReasonDamaged                      StatusReason = "Damaged"
	StatusReasonLastPrimeDebitLinkageDeleted StatusReason = "Last Prime debit linkage deleted"
	StatusReasonClosed                       StatusReason = "Closed"
	StatusReasonFraud                        StatusReason = "Fraud"
)

func (d StatusReason) Pointer() *StatusReason { return &d }

type StatusCode string

const (
	StatusCodeClosed                StatusCode = "C"
	StatusCodeDelinquentReturn      StatusCode = "D"
	StatusCodeDelinquentRetain      StatusCode = "E"
	StatusCodeIssued                StatusCode = "I"
	StatusCodeLost                  StatusCode = "L"
	StatusCodeStolen                StatusCode = "S"
	StatusCodeUnissuedNdIciCards    StatusCode = "U"
	StatusCodeTemporaryBlock        StatusCode = "T"
	StatusCodeBlockAtm              StatusCode = "A"
	StatusCodeBlockAtmPosExcludeCnp StatusCode = "B"
	StatusCodeBlockAtmPosCnpBch     StatusCode = "F"
	StatusCodeBlockAtmPosCnp        StatusCode = "G"
	StatusCodeBlockCnp              StatusCode = "H"
	StatusCodeBlockPosExcludeCnp    StatusCode = "P"
)

type CollectionStatus string

const (
	CollectionStatusNotCollected CollectionStatus = "Card Not Collected"
	CollectionStatusCollected    CollectionStatus = "Card Collected"
)

type LimitType string

const (
	LimitTypeATMEFTPOS LimitType = "ATMEFTPOS"
	LimitTypeAPO       LimitType = "APO"
)

func (s LimitType) String() string {
	return string(s)
}

type NewLimits struct {
	// The types of limits set on the card and their values
	Type                LimitType `json:"type"`
	DailyLimit          int64     `json:"dailyLimit"`
	DailyLimitAvailable int64     `json:"dailyLimitAvailable"`
	// Date of the last transaction in YYYY-MM-DD format
	LastTransaction string `json:"lastTransaction"`
}

type Card struct {
	Token       string `json:"token"`
	Last4Digits string `json:"last4digits"`
}

type Wallet struct {
	ApplePay   uint32 `json:"applePay"`
	GooglePay  uint32 `json:"googlePay"`
	SamsungPay uint32 `json:"samsungPay"`
	Fitness    uint32 `json:"fitness"`
	ECommerce  uint32 `json:"eCommerce"`
	Other      uint32 `json:"other"`
}

type DebitCardResponse struct {
	// Title of the cardholder
	Title string `json:"title"`
	// FirstName of the cardholder
	FirstName string `json:"firstName"`
	// LastName of the cardholder
	LastName string `json:"lastName"`
	// The card's ProductCode
	ProductCode string `json:"productCode"`
	// The card's SubProductCode
	SubProductCode string     `json:"subProductCode"`
	StatusCode     StatusCode `json:"statusCode"`
	Status         Status     `json:"status"`
	// Number of Accounts linked to the card
	AccountsLinkedCount int64 `json:"accountsLinkedCount"`
	// The card's ExpiryDate in YYMM format
	ExpiryDate string `json:"expiryDate"`
	// The IssueReason for the debit card. Populated for Scheme Debit Card only, for all other cards will contain Spaces
	// IssueReason includes New, Reissue, Lost or Stolen, Replacement
	IssueReason string `json:"issueReason"`
	// ActivationStatus of the Card
	ActivationStatus bool        `json:"activationStatus"`
	Limits           []NewLimits `json:"limits"`
	// The first embossed line on the physical card
	EmbossingLine1 string `json:"embossingLine1"`
	// The second embossed line on the physical card
	EmbossingLine2 string `json:"embossingLine2"`
	// The closing date of debit card in YYYY-MM-DD format (if applicable)
	ClosedDate   *string      `json:"closedDate,omitempty"`
	StatusReason StatusReason `json:"statusReason"`
	// Total number of cards
	TotalCards       int64            `json:"totalCards"`
	DispatchedMethod DispatchedMethod `json:"dispatchedMethod"`
	// Number of times the card has been replaced
	ReplacementCount int64 `json:"replacementCount"`
	// Branch ID from where the card was issued
	IssueBranch int `json:"issueBranch"`
	// Branch ID from where the card was collected
	CollectionBranch int              `json:"collectionBranch"`
	CollectionStatus CollectionStatus `json:"collectionStatus"`
	// Card ReplacedDate in YYYY-MM-DD format
	ReplacedDate string `json:"replacedDate"`
	// ReissueDate of the card in YYYY-MM-DD format
	ReissueDate string `json:"reissueDate"`
	// Card IssueDate of the in YYYY-MM-DD format
	IssueDate     string `json:"issueDate"`
	CardNumber    Card   `json:"cardNumber"`
	OldCardNumber *Card  `json:"oldCardNumber,omitempty"`
	NewCardNumber *Card  `json:"newCardNumber,omitempty"`
	// Previous card's expiry date in YYYYMM format
	PrevExpiryDate string `json:"prevExpiryDate"`
	// Last PIN change date in YYYYMM format
	PinChangeDate string `json:"pinChangeDate"`
	// Number of times the PIN has been changed
	PinChangedCount int64 `json:"pinChangedCount"`
	// Last PIN failure date in YYYYMM format
	LastPinFailed string `json:"lastPinFailed"`
	// Number of times the PIN has failed
	PinFailedCount int64 `json:"pinFailedCount"`
	// Card status changed date in YYYYMM format
	StatusChangedDate string `json:"statusChangedDate"`
	// Card details changed date in YYYYMM format
	DetailsChangedDate string `json:"detailsChangedDate"`
	// Card status changed User ID
	StatusChangedUserID string `json:"statusChangedUserID"`
	// Card details changed User ID
	DetailsChangedUserID string `json:"detailsChangedUserID"`
	// Fee Waived Reason
	FeeWaivedReason string `json:"feeWaivedReason"`
	// The design code determines the card appearance as understood by the embosser
	DesignCode  string `json:"designCode"`
	DesignColor string `json:"designColor"`
	// When this field is set to true (i.e. Opted-In), it means the customer is allowing Visa to pass on their updated
	// card number and expiry date information to participating merchants, so any payments set up from this card at
	// registered Merchants are not interrupted. Recurring bills will continue, without the need for the customer to
	// advise them. The customer however, has the option to ‘Opt-out’, by contacting ANZ, where-by this new indicator
	// will need to be updated to false.
	MerchantUpdatePreference bool `json:"merchantUpdatePreference"`
	// When set to true, it indicates customer has enabled visa card control to better manage their finance. This field
	// will also be included in CTM Non-Monetary events published to CIM, so that it can be passed on to CAM and Falcon.
	// Updates to the indicator will also be logged by CTM to be passed to Base24 every 15 minutes in the ‘trickle feed’
	// process.
	CardControlPreference bool `json:"cardControlPreference"`
	// Shows the number of times this card has been tokenised into each type of digital Wallet
	Wallets Wallet `json:"wallets"`
}

// Use this Debit Card Inquiry API to retrieve Card details using the CTM Card Inquiry service. For a given debit card
// number, information such as status, product codes, customer linkage, daily limits of the FSO and ATM transactions,
// history, etc. will be fetched for the consumer. This API interacts with of the CTM message PCTM-PCM-CARD-ENQ v08.
func (c client) DebitCardInquiry(ctx context.Context, tokenizedCardNumber string) (*DebitCardResponse, error) {
	debitCardInquiryURL := fmt.Sprintf(debitCardInqUrlTemplate, c.baseURL, tokenizedCardNumber)

	resp, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodGet, debitCardInquiryURL, nil), "ctm:DebitCardInquiry")
	if err != nil {
		return nil, err
	}

	var response DebitCardResponse
	if err = json.Unmarshal(resp, &response); err != nil {
		logf.Error(ctx, err, "debit card inquiry failed unexpected response from downstream")
		return nil, anzerrors.Wrap(err, codes.Internal, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}

	return &response, nil
}

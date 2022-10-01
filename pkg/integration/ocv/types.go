//nolint:deadcode,varcheck
package ocv

import (
	"strings"
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/pkg/accountformatter"
	"github.com/pkg/errors"
)

type LinkedParty struct {
	// Ownership to the Contract (Package / LinkedAccount / Service).
	// Type of ownership on the contract. Possible values - SOL, JNT (not in beta phase). If not present SOL will be considered.
	RelationshipType *string `json:"relationshipType,omitempty"`
	// Party-to-Contract relationship start date. If not present, Current_Timestamp will be used instead.
	StartDate *string `json:"startDate,omitempty"`
	// Party-to-Contract relationship end date
	EndDate *string `json:"endDate,omitempty"`
	Party   Party   `json:"party"`
}

type Party struct {
	OcvID string `json:"ocvId"`
}

type Account struct {
	// DDA, Savings LinkedAccount or VISA Debit Card number. Mandatory for the package-to-account & account-to-card link scenario.
	AccountNumber *string `json:"accountNumber,omitempty"`
	// Type of the contract. Must be present for add and update package request. Possible values - PACKAGE, ACCOUNT, SERVICE.
	AccountType accountType `json:"accountType"`
	// The name of the agreement or contract
	AccountNameOne *string `json:"accountNameOne,omitempty"`
	// In CAP the account name can span across 3 separate fields. These additional custom fields are used to ensure the full account name is captured in OCV
	AccountNameTwo *string `json:"accountNameTwo,omitempty"`
	// In CAP the account name can span across 3 separate fields. These additional custom fields are used to ensure the full account name is captured in OCV
	AccountNameThree *string `json:"accountNameThree,omitempty"`
	// The alternate name of the contract
	AccountShortName *string `json:"accountShortName,omitempty"`
	// Identifies the type of product associated with the account or card. Mandatory for the CAP LinkedAccount & Service creation scenario. Possible values- DDA, PDV
	ProductCode string `json:"productCode"`
	// Sub product code for the LinkedAccount or Service. Possible values - XD, 101
	AccountSubProduct *string `json:"accountSubProduct,omitempty"`
	// LinkedAccount or Service start date. If not present, Current_Timestamp will be used instead.
	AccountOpenedDate *string `json:"accountOpenedDate,omitempty"`
	// LinkedAccount end date.
	AccountClosedDate *string `json:"accountClosedDate,omitempty"`
	// Identifies the accountRelationshipStatus of the LinkedAccount or Service (Card). Values should be on the EDG standard - Open, Active, Suspended, Dormant and Closed.
	AccountStatus *ctm.StatusCode `json:"accountStatus,omitempty"`
	// LinkedAccount company id, signifies CAP country code. Mandatory for CAP LinkedAccount creation.
	CompanyID *string `json:"companyId,omitempty"`
	// ANZx product code. Possible marketing code values are yet to be confirmed.
	MarketingCode *string `json:"marketingCode,omitempty"`
	// DDA or Savings LinkedAccount branch number.
	AccountBranchNumber *string `json:"accountBranchNumber,omitempty"`
	// Originating Source system for the LinkedAccount & Service. Use - CAP-CIS for LinkedAccount & CTM for Service (Card)
	AccountSource string `json:"accountSource"`
	// Signifies Party-to-Account or Party-to-Card relationship. Mandatory for the account & service creation scenario.
	LinkedParties []LinkedParty `json:"linkedParties"`
}

type LinkedAccount struct {
	Account              Account                `json:"account"`
	AccountRelationships []AccountRelationships `json:"accountRelationships"`
}

type accountRelationshipStatus string

const (
	accountRelationshipStatusActive    accountRelationshipStatus = "Active"
	accountRelationshipStatusInactive  accountRelationshipStatus = "Inactive"
	accountRelationshipStatusCancelled accountRelationshipStatus = "Cancelled"
	accountRelationshipStatusPending   accountRelationshipStatus = "Pending"
	accountRelationshipStatusSuspended accountRelationshipStatus = "Suspended"
	accountRelationshipStatusReplaced  accountRelationshipStatus = "Replaced"
)

type relationshipValue string

const (
	relationshipValueComponentOf              relationshipValue = "Component of"
	relationshipValueFeeDebitLinkedAccountFor relationshipValue = "Fee Debit LinkedAccount For"
)

type AccountRelationships struct {
	// Package-to-LinkedAccount or LinkedAccount-to-Card relationship start date. If not present, Current_Timestamp will be used instead.
	StartDate string `json:"startDate,omitempty"`
	// Package-to-LinkedAccount or LinkedAccount-to-Card relationship end date
	EndDate string `json:"endDate,omitempty"`
	// Status of the relationship.Possible values - In Progress,Active,Pending
	Status accountRelationshipStatus `json:"status,omitempty"`
	// Possible values - Component of, Fee Debit LinkedAccount For
	RelationshipValue relationshipValue `json:"relationshipValue,omitempty"`
}

type accountType string

const (
	accountTypePackage accountType = "PACKAGE"
	accountTypeAccount accountType = "ACCOUNT"
	accountTypeService accountType = "SERVICE"
)

type productCode string

const (
	productCodeTS1 productCode = "TRANSACT&SAVING01"
	productCodeDDA productCode = "DDA"
)

type MaintainContractRequest struct {
	// Signifies the contract number, e.g. Package number or LinkedAccounts number. Must not be present in case of Create package payload.
	AccountNumber *string `json:"accountNumber,omitempty"`
	// Type of the contract. Must be present for add and update package request. Possible values - PACKAGE, ACCOUNT, SERVICE - recommended
	AccountType accountType `json:"accountType"`
	// Product marketing code for the Package or LinkedAccounts. Mandatory for the add package, delta package update, or other linking flow such as P2A and A2C. Possible values - TRANSACT&SAVING01 and DDA.
	ProductCode productCode `json:"productCode"`
	// Agreement name. It is an optional field. May be present for the add package flow.
	AccountNameOne *string `json:"accountNameOne,omitempty"`
	// In CAP the account name can span across 3 separate fields. These additional custom fields are used to ensure the full account name is captured in OCV
	AccountNameTwo *string `json:"accountNameTwo,omitempty"`
	// In CAP the account name can span across 3 separate fields. These additional custom fields are used to ensure the full account name is captured in OCV
	AccountNameThree *string `json:"accountNameThree,omitempty"`
	// The alternate name of the contract
	AccountShortName *string `json:"accountShortName,omitempty"`
	// Package currency. Must be present for add package scenario. Possible values - AUD, USD and so on.
	AccountCurrency *string `json:"accountCurrency,omitempty"`
	// Date on which the package is opened. If not present, Current_Timestamp will be used instead.
	AccountOpenedDate *string `json:"accountOpenedDate,omitempty"`
	// Date on which the package is closed.
	AccountClosedDate *string `json:"accountClosedDate,omitempty"`
	// Identifies the accountRelationshipStatus of the Package. Values should be on the EDG standard - Open, Active, Suspended, Dormant and Closed.
	AccountStatus *string `json:"accountStatus,omitempty"`
	// Originating Source system for the Package - for create package it's ZAFIN, whereas for linking card to an account, it's CAP-CIS
	AccountSource string `json:"accountSource"`
	// Package code represents the package that the customer subscribes to (or cancels package).
	MarketingCode *string `json:"marketingCode,omitempty"`
	// Signifies Party-to-Package relationship. Mandatory for add package scenario.
	LinkedParties []LinkedParty `json:"linkedParties,omitempty"`
	// LinkedAccounts(s) or Service (Card) to be mapped to a package or account respectively. Mandatory for linking Package-to-LinkedAccount or LinkedAccounts-to-Card scenario. Must not be present for add package flow.
	LinkedAccounts []LinkedAccount `json:"linkedAccounts,omitempty"`
}

type LinkedPartyResponse struct {
	Party Party `json:"party"`
}

type AccountResponse struct {
	// Account or Card number. Mandatory for the package-to-account & account-to-card link scenario.
	AccountNumber string `json:"accountNumber"`
	// Account or Card or Service. Mandatory for the package-to-account & account-to-card link scenario.Possible values - ACCOUNT,SERVICE,CARD
	AccountType string `json:"accountType"`
	// Signifies contract key which is combination of raw account number_product code_company id.
	AccountKey string `json:"accountKey"`
	// Signifies Party-to-Package relationship. Mandatory for add package scenario.
	LinkedParties []LinkedPartyResponse `json:"linkedParties"`
}

type LinkedAccountResponse struct {
	Account AccountResponse `json:"account"`
}

type Response struct {
	// Signifies the contract number, e.g. Package number or LinkedAccounts number. Must not be present in case of Create package payload.
	AccountNumber string `json:"accountNumber,omitempty"`
	// Signifies contract key combination of raw account number_product code_company id
	AccountKey string `json:"accountKey"`
	// Type of the contract. Must be present for add and update package request. Possible values - PACKAGE,ACCOUNT,SERVICE,CARD
	AccountType string `json:"accountType"`
	// Signifies Party-to-Package relationship. Mandatory for add package scenario.
	LinkedParties []LinkedPartyResponse `json:"linkedParties"`
	// LinkedAccounts(s) or Service (Card) to be mapped to a package or account respectively. Mandatory for linking Package-to-LinkedAccount or LinkedAccounts-to-Card scenario. Must not be present for add package flow.
	LinkedAccounts []LinkedAccountResponse `json:"linkedAccounts"`
}

type GetPartyByIDRequest struct {
	Identifiers []IdentifierInfo `json:"identifiers"`
}

// IdentifierInfo holds the indentifers needed for the request.
type IdentifierInfo struct {
	IdentifierUsageType string `json:"identifierUsageType"`
	Identifier          string `json:"identifier"`
}

// RetrievePartyRs retrieve party rs
type RetrievePartyRs struct {
	// accounts
	Accounts []*RetrievePartyRsAccount `json:"accounts"`
	// identifiers
	Identifiers []*Identifier `json:"identifiers"`
	// source
	// Example: CAP-CIS
	Source string `json:"source,omitempty"`
	// source systems
	SourceSystems []*SourceSystem `json:"sourceSystems"`
}

// SourceSystem source system
type SourceSystem struct {
	// ID of Source System
	// Example: 1212201901
	// Required: true
	SourceSystemID string `json:"sourceSystemId"`

	// Source System Name
	// Example: CAP-CIS
	// Required: true
	SourceSystemName string `json:"sourceSystemName"`
}

func (r RetrievePartyRs) GetAccount(accountNumber string) *RetrievePartyRsAccount {
	formattedAccount := accountformatter.NewFormattedAccount(accountNumber).String()
	var out *RetrievePartyRsAccount
	for _, account := range r.Accounts {
		a := accountformatter.NewFormattedAccount(account.AccountNumber).String()
		if formattedAccount == a {
			out = account
			break
		}
	}

	return out
}

func (r RetrievePartyRs) GetCAPCSID() (string, error) {
	var (
		capID   = "CAP ID"
		capcsID = "CAP-CIS"
		out     string
	)

	for _, identifier := range r.Identifiers {
		if capID == identifier.IdentifierUsageType && capcsID == identifier.Source {
			out = identifier.Identifier
			break
		}
	}

	if out == "" {
		return "", errors.New("unable to find CAP-CIS ID in party response")
	}

	return strings.TrimLeft(out, "0"), nil
}

// RetrievePartyRsAccount account
type RetrievePartyRsAccount struct {
	// RetrievePartyRsAccount Branch Number
	// Example: 01
	AccountBranchNumber string `json:"accountBranchNumber,omitempty"`

	// RetrievePartyRsAccount Closed Date
	// Example: 2010-05-14
	// Format: date
	AccountClosedDate string `json:"accountClosedDate,omitempty"`

	// Agreement Name
	// Example: GAURA ENTERPRISES PTY LTD
	AccountNameOne string `json:"accountNameOne,omitempty"`

	// Agreement Name Three
	// Example: KATHA TRUST
	AccountNameThree string `json:"accountNameThree,omitempty"`

	// Agreement Name Two
	// Example: ACN 123 301 790 IIOC \u0026 ATF DEVA
	AccountNameTwo string `json:"accountNameTwo,omitempty"`

	// RetrievePartyRsAccount Number
	// Example: 708405438
	AccountNumber string `json:"accountNumber,omitempty"`

	// RetrievePartyRsAccount Open Date
	// Example: 2005-05-14
	// Format: date
	AccountOpenedDate string `json:"accountOpenedDate,omitempty"`

	// Agreement Nickname
	// Example: DEVA KATHA TRU
	AccountShortName string `json:"accountShortName,omitempty"`

	// RetrievePartyRsAccount Status
	// Example: Active
	AccountStatus string `json:"accountStatus,omitempty"`

	// RetrievePartyRsAccount raw status
	// Example: Validated
	AccountStatusRaw string `json:"accountStatusRaw,omitempty"`

	// RetrievePartyRsAccount Sub Product Code for the Source System. Sub Product Code is separated by : from the SourceSystem. In the eg. CAP-CIS is the source system and ddabc is sub product code
	// Example: CAP-CIS:DDABC
	AccountSubProduct string `json:"accountSubProduct,omitempty"`

	// Type of the contract. Must be present for add and update package request. Possible values - PACKAGE, ACCOUNT, SERVICE.
	// Example: PACKAGE
	AccountType string `json:"accountType,omitempty"`

	// companyId
	// Example: 01
	CompanyID string `json:"companyId,omitempty"`

	// Marketing code represents the product code for the package, account or service (card) that the customer subscribes to. Possible values are - TRANSACT&SAVING01 for package, TRANSACT01 for DDA account, SAVING01 for Savings account, CVDC for VISA debit card.
	// Example: TRANSACT\u0026SAVING01
	MarketingCode string `json:"marketingCode,omitempty"`

	// Number Of Signatures
	// Example: 2
	NumberOfSignatures string `json:"numberOfSignatures,omitempty"`

	// RetrievePartyRsAccount Product Code for the Source System. Product Code is separated by : from the SourceSystem. In the eg. CAP-CIS is the source system and dda is product code
	// Example: CAP-CIS:DDA
	ProductCode string `json:"productCode,omitempty"`

	// Regulated Indicator
	// Example: Y
	RegulatedIndicator string `json:"regulatedIndicator,omitempty"`

	// RelationshipType
	// Example: GTR
	RelationshipType string `json:"relationshipType,omitempty"`
}

// RetrievePartyRq retrieve party rq
//
type RetrievePartyRq struct {
	// Retrieve using identifiers. While doing identifier retrieve sourceSystems should not be provided.
	// Max Items: 1
	// Min Items: 0
	Identifiers []*IdentifierRq `json:"identifiers"`
}

// IdentifierRq identifier rq
//
type IdentifierRq struct {
	// Identifier
	// Example: 12345679801
	// Required: true
	Identifier *string `json:"identifier"`

	// Identifier Usage Type
	// Example: ABN
	// Required: true
	IdentifierUsageType *string `json:"identifierUsageType"`
}

// Identifier identifier
//
// swagger:model Identifier
type Identifier struct {
	// Identifier
	// Example: 12345679801
	Identifier string `json:"identifier,omitempty"`

	// Identifier Usage Type
	// Example: ABN
	IdentifierUsageType string `json:"identifierUsageType,omitempty"`

	// Source System Name
	// Example: CAP-CIS
	Source string `json:"source,omitempty"`
}

func (a *RetrievePartyRsAccount) IssuedToday() bool {
	return a.AccountOpenedDate == time.Now().Format("2006-01-02")
}

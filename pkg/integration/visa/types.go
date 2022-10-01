package visa

import (
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/protobuf/types/known/durationpb"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

type TransactionControlDocument struct {
	// Required - The time the request is received. Value is in UTC time
	ReceivedTimestamp string `json:"receivedTimestamp"`

	// Required - The processing time in milliseconds
	ProcessingTimeInMS int64 `json:"processingTimeinMs"`

	// Required
	Resource Resource `json:"resource"`
}

type Request struct {
	// Optional - Used to set rules that apply to all transactions performed with this account.
	GlobalControls []*GlobalControl `json:"globalControls"`

	// Optional - These control-types are used to target different Merchant Groups based on their Merchant Category Code (MCC).
	MerchantControls []*MerchantControl `json:"merchantControls"`

	// Optional - These control-types target specific types of transactions such as ATM withdrawals or E-commerce.
	TransactionControls []*TransactionControl `json:"transactionControls"`
}

type Resource struct {
	// Optional - Used to liberate tokens from any card control settings placed on the primaryAccountNumber.
	// If set to 'true' on an enrolled paymentToken then the token will operate completely independent
	// of the PAN's card controls. If set to 'true' on a primaryAccountNumber then all related token
	// transactions will be ignored by the PAN's card controls.
	ShouldDecouple bool `json:"shouldDecouple"`

	// Optional - Provides the timestamp, in UTC, of when the resource was last updated.
	LastUpdateTimeStamp string `json:"lastUpdateTimeStamp"`

	// Optional - System generated ID for the Control Document that is bound to the Account Identifier
	// (primaryAccountNumber/paymentToken)
	DocumentID string `json:"documentID"`

	// Optional - Used to set rules that apply to all transactions performed with this account.
	GlobalControls []*GlobalControl `json:"globalControls"`

	// Optional - These control-types are used to target different Merchant Groups based on their Merchant Category Code (MCC).
	MerchantControls []*MerchantControl `json:"merchantControls"`

	// Optional - These control-types target specific types of transactions such as ATM withdrawals or E-commerce.
	TransactionControls []*TransactionControl `json:"transactionControls"`
}

// GlobalControl Used to set rules that apply to all transactions performed with this account.
type GlobalControl struct {
	// Required - If true, VTC will trigger a decline for all transactions related to this control type.
	// If false, other attributes like DeclineThreshold will be checked.
	ShouldDeclineAll bool `json:"shouldDeclineAll"`

	// Optional - Will trigger a decline for all transactions with amounts that equal or exceed this threshold
	// for this control type. During authorization processing the cardholderBillAmount is then used
	// for comparison to identify if a decline should be triggered.
	DeclineThreshold *float64 `json:"declineThreshold"`

	// Required - If true, this control type will be applied during transaction processing.
	// If false, this control type will not be applied during transaction processing.
	ControlEnabled bool `json:"isControlEnabled"`

	// Optional - Uniquely identifies the cardholder who is to receive the alert message. The UserIdentifier should be a
	// GUID but at a minimum it must be unique per enrolling application and must not contain any PII data.
	// It is mandatory for all VTC notifications. The maximum number of characters allowed is 72.
	UserIdentifier string `json:"userIdentifier"`

	// Optional - Will trigger an alert for all approved transactions with amounts that equal or exceed the threshold
	// setting for this control type. During authorization processing the cardholderBillAmount is used for
	// comparison to identify if an alert should be sent.
	AlertThreshold *float64 `json:"alertThreshold"`

	// Optional - Used when the physical card is lost or misplaced but the cardholder wants to continue transacting
	// via their mobile phone or wearable. If set to 'true' any transactions performed with primaryAccountNumber
	// will be declined; however, token-based transactions will continue to process as normal.
	// **Only applicable with global controls.**
	DeclineAllNonTokenizeTransactions bool `json:"declineAllNonTokenizeTransactions"`

	// Optional - If true, VTC will trigger a decline notification for all transactions matching the associated
	// control type. If false, no alerts will be sent for declined transactions related to this control type.
	ShouldAlertOnDecline bool             `json:"shouldAlertOnDecline"`
	UserInformation      *userInformation `json:"userInformation"`
	SpendLimit           *spendLimit      `json:"spendLimit"`
	FilterByCountry      *filterByCountry `json:"filterByCountry"`
	TimeRange            *timeRange       `json:"timeRange"`
}

var ControlType_value = map[string]ccpb.ControlType{
	"UNKNOWN_UNSPECIFIED":         0,
	"TCT_ATM_WITHDRAW":            1,
	"TCT_AUTO_PAY":                2,
	"TCT_BRICK_AND_MORTAR":        3,
	"TCT_CROSS_BORDER":            4,
	"TCT_E_COMMERCE":              5,
	"TCT_CONTACTLESS":             6,
	"MCT_ADULT_ENTERTAINMENT":     7,
	"MCT_AIRFARE":                 8,
	"MCT_ALCOHOL":                 9,
	"MCT_APPAREL_AND_ACCESSORIES": 10,
	"MCT_AUTOMOTIVE":              11,
	"MCT_CAR_RENTAL":              12,
	"MCT_ELECTRONICS":             13,
	"MCT_SPORT_AND_RECREATION":    14,
	"MCT_GAMBLING":                15,
	"MCT_GAS_AND_PETROLEUM":       16,
	"MCT_GROCERY":                 17,
	"MCT_HOTEL_AND_LODGING":       18,
	"MCT_HOUSEHOLD":               19,
	"MCT_PERSONAL_CARE":           20,
	"MCT_SMOKE_AND_TOBACCO":       21,
	"GCT_GLOBAL":                  22,
}

// MerchantControl These control-types are used to target different Merchant Groups based on their Merchant Category Code (MCC).
type MerchantControl struct {
	// Required - If true, VTC will trigger a decline for all transactions related to this control type. If false,
	// other attributes like DeclineThreshold will be checked.
	ShouldDeclineAll bool `json:"shouldDeclineAll"`

	// Optional - The point in time (GMT) at which the ImpulseDelayPeriod will expire. This value can also be
	// calculated by adding the ImpulseDelayPeriod value to the ImpulseDelayStart value.
	ImpulseDelayEnd *string `json:"impulseDelayEnd"`

	// Optional - The point in time (GMT) at which the initial ImpulseDelayPeriod was received on a merchant control.
	ImpulseDelayStart *string `json:"impulseDelayStart"`

	// Optional - When specified as part of a merchant control within a request with a value in the format (HH:MM),
	// that control will remain in a blocked status until that time period is elapsed. If a request is sent with the
	// ImpulseDelayPeriod element missing on a merchant control then the block is removed immediately and all
	// previous controls are reinstated.
	ImpulseDelayPeriod *string `json:"impulseDelayPeriod"`

	// Optional - If true, VTC will trigger a decline notification for all transactions matching the associated
	// control type. If false, no alerts will be sent for declined transactions related to this control type.
	ShouldAlertOnDecline bool `json:"shouldAlertOnDecline"`

	// Optional - Will trigger a decline for all transactions with amounts that equal or exceed this threshold for
	// this control type. During authorization processing the cardholderBillAmount is then used for comparison to
	// identify if a decline should be triggered.
	DeclineThreshold *float64 `json:"declineThreshold"`

	// Required - If true, this control type will be applied during transaction processing. If false, this control type
	// will not be applied during transaction processing.
	ControlEnabled bool `json:"isControlEnabled"`

	// Required - Indicates the specific type of the control
	ControlType string `json:"controlType"`

	// Optional - Uniquely identifies the cardholder who is to receive the alert message. The UserIdentifier should
	// be a GUID but at a minimum it must be unique per enrolling application and must not contain any PII data.
	// It is mandatory for all VTC notifications. The maximum number of characters allowed is 72.
	UserIdentifier string `json:"userIdentifier"`

	// Optional - Will trigger an alert for all approved transactions with amounts that equal or exceed the
	// threshold setting for this control type. During authorization processing the cardholderBillAmount is used
	// for comparison to identify if an alert should be sent.
	AlertThreshold *float64 `json:"alertThreshold"`

	// Optional - The amount of time remaining before the impulseDelayPeriod expires. This value can also be
	// calculated by subtracting the impulseDelayStart time from the impulseDelayEnd time.
	ImpulseDelayRemaining *string `json:"impulseDelayRemaining"`

	UserInformation *userInformation `json:"userInformation"`

	SpendLimit *spendLimit `json:"spendLimit"`

	FilterByCountry *filterByCountry `json:"filterByCountry"`

	TimeRange *timeRange `json:"timeRange"`
}

func (c MerchantControl) GetImpulseDelayStartTimestamp() *timestamppb.Timestamp {
	if c.ImpulseDelayStart == nil {
		log.Printf("impulse delay start time retrieved when nil")
		return nil
	}
	impulseDelayStart := strings.ReplaceAll(*c.ImpulseDelayStart, "-", "/")
	delayStartTime, err := time.Parse("2006/01/02 15:04:05", impulseDelayStart)
	if err != nil {
		log.Printf("unable to parse impulse delay start time %v", err)
		return nil
	}
	return timestamppb.New(delayStartTime)
}

func (c MerchantControl) GetImpulseDelayPeriodProto() *durationpb.Duration {
	if c.ImpulseDelayPeriod == nil {
		log.Printf("impulse delay period retrieved when nil")
		return nil
	}
	impulseDelayPeriod := strings.ReplaceAll(*c.ImpulseDelayPeriod, ":", "h")
	delayPeriodTime, err := time.ParseDuration(fmt.Sprintf("%sm", impulseDelayPeriod))
	if err != nil {
		log.Printf("unable to parse impulse delay period %v", err)
		return nil
	}
	return durationpb.New(delayPeriodTime)
}

// TransactionControl These control-types target specific types of transactions such as ATM withdrawals or E-commerce.
type TransactionControl struct {
	// Required - If true, VTC will trigger a decline for all transactions related to this control type. If false,
	// other attributes like DeclineThreshold will be checked.
	ShouldDeclineAll bool `json:"shouldDeclineAll"`

	// Optional - Will trigger a decline for all transactions with amounts that equal or exceed this threshold for
	// this control type. During authorization processing the cardholderBillAmount is then used for comparison to
	// identify if a decline should be triggered.
	DeclineThreshold *float64 `json:"declineThreshold"`

	// Required - If true, this control type will be applied during transaction processing. If false, this control
	// type will not be applied during transaction processing.
	ControlEnabled bool `json:"isControlEnabled"`

	// Required - Indicates the specific type of the control
	ControlType string `json:"controlType"`

	// Optional - Uniquely identifies the cardholder who is to receive the alert message. The UserIdentifier should be
	// a GUID but at a minimum it must be unique per enrolling application and must not contain any PII data. It is
	// mandatory for all VTC notifications. The maximum number of characters allowed is 72.
	UserIdentifier string `json:"userIdentifier"`

	// Optional - Will trigger an alert for all approved transactions with amounts that equal or exceed the threshold
	// setting for this control type. During authorization processing the cardholderBillAmount is used for comparison
	// to identify if an alert should be sent.
	AlertThreshold *float64 `json:"alertThreshold"`

	// Optional - If true, VTC will trigger a decline notification for all transactions matching the associated control
	// type. If false, no alerts will be sent for declined transactions related to this control type.
	ShouldAlertOnDecline bool `json:"shouldAlertOnDecline"`

	UserInformation *userInformation `json:"userInformation"`
	SpendLimit      *spendLimit      `json:"spendLimit"`
	FilterByCountry *filterByCountry `json:"filterByCountry"`
	TimeRange       *timeRange       `json:"timeRange"`
}

type userInformation struct {
	// Optional - ApplicationDefinedAttributes by the issuer. These can be used to enrich the VTC notification alerts.
	ApplicationDefinedAttributes []applicationDefinedAttributes `json:"applicationDefinedAttributes"`

	// Optional - Identifier for the issuer to map the user Name
	BankingIdentifier *string `json:"bankingIdentifier"`

	// Required - Name of the user who configured the control.Name is mandatory if UserInformation exists.
	Name string `json:"name"`
}

type applicationDefinedAttributes struct { // TODO: work out what the hell this is
}

type SpendLimitType string

const (
	SpendLimitTypeMonth     SpendLimitType = "LMT_MONTH"
	SpendLimitTypeWeek      SpendLimitType = "LMT_WEEK"
	SpendLimitTypeDay       SpendLimitType = "LMT_DAY"
	SpendLimitTypeDateRange SpendLimitType = "LMT_DATE_RANGE"
	SpendLimitTypeRecurring SpendLimitType = "LMT_RECURRING"
)

type spendLimit struct {
	// Optional - The starting date and time of a date and time range bounding the control. Only transactions
	// attempted during the specified date and time range will trigger this control. The StartDateTime cannot be
	// in the past. It can be today or some date and time in the future. The value is in the time zone set by
	// the TimeZoneID field and format is "CCYY-MM-DD HH:MM". The hour and minute portion is optional, if not
	// provided, then these values will be set to "00:00". This field is required only when type = LMT_DATE_RANGE.
	// If it set when the type = LMT_RECURRING then the control will become active at that time, and the behavior
	// will be as if the user set the recurring functionality at that point in time. In all other cases the
	// field is optional and even if present, it will be ignored.
	StartDateTime *string `json:"startDateTime"`

	// Required - The maximum accumulated spend for the time period at which VTC will then trigger declines.
	// Once met or exceeded, all subsequent purchases related to **this control-type will be declined** until
	// the new time period begins (e.g. a new month). If 'alertOnDecline' is true then the cardholder will be
	// notified of these transactions.
	DeclineThreshold float64 `json:"declineThreshold"`

	// Optional - The TimeZoneID is used to determine what time zone should be applied to this object. It
	// should be in Area/Location format. For example: 'America/Denver' is in United States MST.
	// If not provided, it is assumed to be UTC
	TimeZoneID *string `json:"timeZoneID"`

	// Optional - The maximum value of total approved purchases within the time period before triggering an
	// alert. Once met or exceeded, any further purchases related to **this** control-type will **not** trigger
	// another spendLimit alert until the next time period begins (and everything is reset). **However, alerts
	// will still be triggered for any transactions that meet/exceed the per transaction AlertThreshold**.
	AlertThreshold *float64 `json:"alertThreshold"`

	// Required - Provides the time period that the control spans.
	SpendLimitType SpendLimitType `json:"type"`

	// Optional - The ending date and time of a date and time range bounding the control. Only transactions attempted
	// during the specified date and time range will trigger this control. The EndDateTime cannot be before the
	// startDateTime. The value is in the time zone set by the timeZoneID field and the format is "CCYY-MM-DD HH:MM".
	// The hour and minute portion is optional, if not provided, then these values will be set to "23:59".
	// This field is required only when type = LMT_DATE_RANGE. In all other cases, the field is optional and even
	// if present, it will be ignored.
	EndDateTime string `json:"endDateTime"`

	// Required - The total amount of all approved transactions performed within the time period for this control
	// type. It is reset by the amount of the first transaction for this control type once a new time period begins.
	CurrentPeriodSpend float64 `json:"currentPeriodSpend"`

	// Optional - The number of days after which the current period spend is reset and a new cycle begins. This
	// field is required only when type = LMT_RECURRING; in all other cases, the field is optional and even if
	// present, it will be ignored. Allowed minimum value is 1 and maximum value is 366.
	RecurringCycleTime int32 `json:"recurringCycleTime"`
}

type filterByCountry struct {
	// Transactions from this list of countries will be reviewed by this control type. Should be an upper
	// case ISO 3166 3-letter country code.
	ControlDisabledCountryList []string `json:"controlDisabledCountryList"`

	// Transactions from this list of countries **will be ignored** for this control type. Should be an upper
	// case ISO 3166 3-letter country code.
	ControlEnabledCountryList []string `json:"controlEnabledCountryList"`
}

type timeRange struct {
	// Required - The TimeZoneID is used to determine what time zone should be applied to this object. It should
	// be in Area/Location format. For example: 'America/Denver' is in United States MST. If not provided, it is
	// assumed to be UTC
	TimeZoneID string `json:"timeZoneID"`

	// Required - The start time in UTC of a timeRange that bounds a control. Only transactions attempted during the
	// specified timeRange will trigger this control. The format of the string is in 24hr time using HH:MM format.
	StartTime string `json:"startTime"`

	// Required - The end time in UTC of a timeRange that bounds a control. Only transactions attempted during the
	// specified timeRange will trigger this control. The format of the string is in 24hr time using HH:MM format.
	EndTime string `json:"endTime"`
}

func GetCategory(c ccpb.ControlType) Category {
	return controlTypes[c]
}

type Category int32

func (s Category) String() string {
	return [...]string{"TRANSACTION", "MERCHANT", "GLOBAL"}[s]
}

const (
	TRANSACTION Category = iota
	MERCHANT
	GLOBAL
)

var controlTypes = map[ccpb.ControlType]Category{
	ccpb.ControlType_TCT_ATM_WITHDRAW:            TRANSACTION,
	ccpb.ControlType_TCT_AUTO_PAY:                TRANSACTION,
	ccpb.ControlType_TCT_BRICK_AND_MORTAR:        TRANSACTION,
	ccpb.ControlType_TCT_CROSS_BORDER:            TRANSACTION,
	ccpb.ControlType_TCT_E_COMMERCE:              TRANSACTION,
	ccpb.ControlType_TCT_CONTACTLESS:             TRANSACTION,
	ccpb.ControlType_MCT_ADULT_ENTERTAINMENT:     MERCHANT,
	ccpb.ControlType_MCT_AIRFARE:                 MERCHANT,
	ccpb.ControlType_MCT_ALCOHOL:                 MERCHANT,
	ccpb.ControlType_MCT_APPAREL_AND_ACCESSORIES: MERCHANT,
	ccpb.ControlType_MCT_AUTOMOTIVE:              MERCHANT,
	ccpb.ControlType_MCT_CAR_RENTAL:              MERCHANT,
	ccpb.ControlType_MCT_ELECTRONICS:             MERCHANT,
	ccpb.ControlType_MCT_SPORT_AND_RECREATION:    MERCHANT,
	ccpb.ControlType_MCT_GAMBLING:                MERCHANT,
	ccpb.ControlType_MCT_GAS_AND_PETROLEUM:       MERCHANT,
	ccpb.ControlType_MCT_GROCERY:                 MERCHANT,
	ccpb.ControlType_MCT_HOTEL_AND_LODGING:       MERCHANT,
	ccpb.ControlType_MCT_HOUSEHOLD:               MERCHANT,
	ccpb.ControlType_MCT_PERSONAL_CARE:           MERCHANT,
	ccpb.ControlType_MCT_SMOKE_AND_TOBACCO:       MERCHANT,
	ccpb.ControlType_GCT_GLOBAL:                  GLOBAL,
}

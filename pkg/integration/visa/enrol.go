package visa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anzx/fabric-cards/pkg/util/apic"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"google.golang.org/grpc/codes"
)

const registerFailed = "failed to register card number"

type EnrolByPanRequest struct {
	PaymentToken            string                   `json:"paymentToken,omitempty"`
	PrimaryAccountNumber    string                   `json:"primaryAccountNumber,omitempty"`
	AlertsEnrollmentDetails *AlertsEnrollmentDetails `json:"alertsEnrollmentDetails,omitempty"`
}

type AlertsEnrollmentDetails struct {
	CountryCode              string                     `json:"countryCode"`
	FirstName                string                     `json:"firstName"`
	LastName                 string                     `json:"lastName"`
	PortfolioID              string                     `json:"portfolioID,omitempty"`
	PreferredLanguage        string                     `json:"preferredLanguage"`
	UserIdentifier           string                     `json:"userIdentifier"`
	DefaultAlertsPreferences []DefaultAlertsPreferences `json:"defaultAlertsPreferences"`
}

type DefaultAlertsPreferences struct {
	CallingCode          string `json:"callingCode,omitempty"`
	ContactType          string `json:"contactType"`
	ContactValue         string `json:"contactValue"`
	IsVerified           bool   `json:"isVerified,omitempty"`
	PreferredEmailFormat string `json:"preferredEmailFormat,omitempty"`
	Status               string `json:"status,omitempty"`
}

// Register a primaryAccountNumber or paymentToken in Visa Transaction Controls. If the account is already registered
// then VTC will return the existing documentID and its contents.
func (c client) Register(ctx context.Context, primaryAccountNumber string) (string, error) {
	if !checkPrimaryAccountNumber(primaryAccountNumber) {
		return "", anzerrors.New(codes.InvalidArgument, registerFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "cannot parse requested card number"))
	}

	destination := fmt.Sprintf("%s/%s", c.baseURL, enrolByPanEndpoint)

	request := EnrolByPanRequest{
		PrimaryAccountNumber: primaryAccountNumber,
	}

	requestBody, _ := json.Marshal(request)

	response, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, destination, requestBody), "visa:Register")
	if err != nil {
		return "", err
	}

	controlDocument, err := handleTransactionControlDocument(ctx, response)
	if err != nil {
		return "", err
	}

	return controlDocument.DocumentID, nil
}

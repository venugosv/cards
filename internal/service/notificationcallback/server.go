package notificationcallback

import (
	"context"
	"fmt"
	"strconv"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/notification"
	types "github.com/anzx/fabricapis/pkg/fabric/type"
	ncpb "github.com/anzx/fabricapis/pkg/visa/service/notificationcallback"
	log "github.com/anzx/pkg/log"
	"github.com/google/uuid"
)

const (
	transactionDeclined = "DECLINED"
	AUD                 = "036"
)

type server struct {
	ncpb.UnimplementedNotificationCallbackAPIServer
	CommandCentre sdk.Publisher
}

// NewServer constructs a new CustomerRulesAPI from configured clients
func NewServer(cmdcntr sdk.Publisher) ncpb.NotificationCallbackAPIServer {
	return &server{
		CommandCentre: cmdcntr,
	}
}

func (s server) Alert(ctx context.Context, request *ncpb.Request) (*ncpb.Response, error) {
	if !feature.FeatureGate.Enabled(feature.NotificationCallbackDeclinedEvent) {
		logf.Info(ctx, "notification callback declined events are disabled by feature flag")
		return &ncpb.Response{}, nil
	}

	details := request.GetTransactionDetails()
	personaId := details.GetUserIdentifier()
	if personaId == "" {
		return nil, errors.New("no user identifier present visa notification callback request")
	}

	// Early out if the transaction was not declined
	if request.GetTransactionOutcome().GetTransactionApproved() != transactionDeclined {
		logf.Info(ctx, "notification callback: transaction was approved, nothing to do")
		return &ncpb.Response{}, nil
	}

	maskedCardNumber := details.GetPrimaryAccountNumber()
	if maskedCardNumber == "" || len(maskedCardNumber) < 4 {
		return nil, errors.New("this transaction was not associated with a valid card")
	}

	currencyCodeName, err := getCurrencyCode(details.GetBillerCurrencyCode())
	if err != nil {
		return nil, err
	}

	var merchantName string
	if details.GetMerchantInfo() != nil {
		merchantName = details.GetMerchantInfo().GetName()
	}

	moneyValue := details.GetCardholderBillAmount()

	ccreq := transactionDeclinedNotification(ctx, personaId, currencyCodeName, moneyValue, maskedCardNumber, merchantName)

	log.Info(ctx, "Publishing controls declined notification", log.Str("personaID", personaId), log.Str("title", ccreq.Preview.Title), log.Str("body", ccreq.Preview.Body))

	resp, err := s.CommandCentre.Publish(ctx, ccreq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to publish to pubsub")
	}

	if resp.Status != sdk.PublishResponsePublished {
		return nil, errors.New("failed to publish controls declined alert")
	}

	log.Info(ctx, "Notification sent", log.Str("personaID", personaId))

	return &ncpb.Response{}, nil
}

func getCurrencyCode(currencyCodeDigits string) (string, error) {
	if currencyCodeDigits == AUD {
		return "$", nil
	}
	// We get an ISO-4217 3-digit code as a string, so parse it to an integer then use our map to get the name
	// eg. "036" -> int32(36) -> "AUD"
	currencyCodeInt, err := strconv.ParseInt(currencyCodeDigits, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid currency code: %s", currencyCodeDigits)
	}

	currencyCodeName, ok := types.Currency_name[int32(currencyCodeInt)]
	if !ok {
		return "", fmt.Errorf("unknown ISO 4217 country code: %d", currencyCodeInt)
	}

	return currencyCodeName, nil
}

func transactionDeclinedNotification(ctx context.Context, persona string, currency string, value float32, maskedCardNumber string, merchantName string) *sdk.NotificationForPersona {
	valueString := fmt.Sprintf("%0.2f", value)
	last4digits := maskedCardNumber[len(maskedCardNumber)-4:]
	var body string
	if merchantName == "" {
		body = fmt.Sprintf("A transaction of %s%s was declined because of a control you placed on your card ending in %s", currency, valueString, last4digits)
	} else {
		body = fmt.Sprintf("A transaction of %s%s (%s) was declined because of a control you placed on your card ending in %s", currency, valueString, merchantName, last4digits)
	}

	log.Debug(ctx, "Notification Composed", log.Str("currency", currency), log.Str("valueString", valueString), log.Str("merchantName", merchantName), log.Str("last4digits", last4digits))

	return &sdk.NotificationForPersona{
		PersonaID: persona,
		Notification: notification.Simple{
			ActionURL: "https://plus.anz/cards",
		},
		Preview: notification.Preview{
			Title: "Transaction Blocked",
			Body:  body,
		},
		IdempotencyKey: uuid.NewString(),
	}
}

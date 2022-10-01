package callback

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v6"

	ecpb "github.com/anzx/fabricapis/pkg/visa/service/enrollmentcallback"
	ncpb "github.com/anzx/fabricapis/pkg/visa/service/notificationcallback"
	"google.golang.org/grpc"
)

const (
	TransactionCurrencyCode     = "036"
	TransactionCurrencyCodeName = "$"
	TransactionValue            = 1444.89
)

type Callback struct {
	enrollmentCallbackAPIClient   ecpb.EnrollmentCallbackAPIClient
	notificationCallbackAPIClient ncpb.NotificationCallbackAPIClient
	ctx                           context.Context
	currentCard                   string
}

func NewCallback(ctx context.Context, callback *grpc.ClientConn, currentCard string) *Callback {
	return &Callback{
		enrollmentCallbackAPIClient:   ecpb.NewEnrollmentCallbackAPIClient(callback),
		notificationCallbackAPIClient: ncpb.NewNotificationCallbackAPIClient(callback),
		ctx:                           ctx,
		currentCard:                   currentCard,
	}
}

func (c *Callback) GetLast4Digits() string {
	return c.currentCard[len(c.currentCard)-4:]
}

func (c *Callback) Enroll() (*ecpb.Response, error) {
	return c.enrollmentCallbackAPIClient.Enroll(c.ctx, c.enrolmentRequest())
}

func (c *Callback) Disenroll() (*ecpb.Response, error) {
	return c.enrollmentCallbackAPIClient.Disenroll(c.ctx, c.enrolmentRequest())
}

func (c *Callback) enrolmentRequest() *ecpb.Request {
	return &ecpb.Request{
		BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
			{
				PrimaryAccountNumber: c.currentCard,
				ServiceTypes:         []string{"CTC"},
			},
		},
	}
}

func (c *Callback) Alert() (*ncpb.Response, error) {
	return c.notificationCallbackAPIClient.Alert(c.ctx, &ncpb.Request{
		TransactionTypes: []string{},
		AppId:            gofakeit.UUID(),
		SponsorId:        gofakeit.UUID(),
		TransactionDetails: &ncpb.TransactionDetails{
			PaymentToken:             c.currentCard,
			CardholderBillAmount:     TransactionValue,
			NameOnCard:               gofakeit.Name(),
			UserIdentifier:           gofakeit.UUID(),
			BillerCurrencyCode:       TransactionCurrencyCode,
			RequestReceivedTimeStamp: time.Now().UTC().String(),
			PrimaryAccountNumber:     c.currentCard,
			RetrievalReferenceNumber: gofakeit.UUID(),
			TransactionId:            gofakeit.UUID(),
			ExchangeRateDetails:      nil,
			MerchantInfo: &ncpb.MerchantInfo{
				Name:                 "Grill'd Healthy Burgers",
				CountryCode:          "AUS",
				CurrencyCode:         TransactionCurrencyCode,
				TransactionAmount:    TransactionValue,
				MerchantCategoryCode: "abcd",
				City:                 "Melbourne",
				PostalCode:           "3000",
			},
		},
		TransactionOutcome: &ncpb.TransactionOutcome{
			CtcDocumentId:             gofakeit.UUID(),
			DecisionResponseTimeStamp: "1",
			TransactionApproved:       "DECLINED",
			NotificationId:            gofakeit.UUID(),
			DecisionId:                gofakeit.UUID(),
			AlertDetails: []*ncpb.AlertDetails{
				{
					TriggeringAppId: gofakeit.UUID(),
					RuleCategory:    "PCT_MERCHANT",
					RuleType:        "MCT_GAMBLING",
				},
			},
		},
	})
}

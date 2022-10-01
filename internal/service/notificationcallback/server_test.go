package notificationcallback

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	cc "github.com/anzx/fabric-cards/test/stubs/grpc/commandcentre"

	"github.com/anzx/fabric-cards/pkg/feature"
	cardcontrols "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	ncpb "github.com/anzx/fabricapis/pkg/visa/service/notificationcallback"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/fixtures"
	"github.com/stretchr/testify/assert"
)

const aPersonaID = "d91acf54-4c87-48aa-85a9-dd41c72c54d6"

func TestNewService(t *testing.T) {
	c := fixtures.AServer().WithData(data.AUserWithACard())
	got := NewServer(c.CommandCentreEnv)
	assert.NotNil(t, got)
	assert.IsType(t, &server{}, got)
}

func TestServer_Alert_WhenFeatureDisabled_ReturnsOk(t *testing.T) {
	disableNotificationMap := map[feature.Feature]bool{
		feature.NotificationCallbackDeclinedEvent: false,
	}
	err := feature.FeatureGate.Set(disableNotificationMap)
	require.NoError(t, err)

	req := &ncpb.Request{
		TransactionOutcome: &ncpb.TransactionOutcome{
			TransactionApproved: "DECLINED",
		},
	}

	fakeCc := cc.NewFakePublisher()
	s := &server{
		CommandCentre: &commandcentre.Client{
			Publisher: &fakeCc,
		},
	}

	_, err = s.Alert(context.Background(), req)

	require.Equal(t, 0, fakeCc.Count, "nothing should be published if the feature is disabled")
	require.NoError(t, err)
}

func TestServer_Alert(t *testing.T) {
	tests := []struct {
		name               string
		request            *ncpb.Request
		expectedError      string
		expectedPubSubSend int
		want               string
	}{
		{
			name:    "Skip invalid request",
			request: &ncpb.Request{},
		},
		{
			name: "happy path",
			request: &ncpb.Request{
				TransactionDetails: &ncpb.TransactionDetails{
					UserIdentifier:           aPersonaID,
					BillerCurrencyCode:       "036",
					RequestReceivedTimeStamp: "12345",
					PrimaryAccountNumber:     "1234123412341234",
					MerchantInfo: &ncpb.MerchantInfo{
						Name:                 "Lime",
						CountryCode:          "AUD",
						MerchantCategoryCode: "1234",
						CurrencyCode:         "AUD",
					},
				},
				TransactionOutcome: &ncpb.TransactionOutcome{
					DecisionId:                "123",
					NotificationId:            "abc123",
					TransactionApproved:       "DECLINED",
					DecisionResponseTimeStamp: "1234",
					AlertDetails: []*ncpb.AlertDetails{
						{
							TriggeringAppId: gofakeit.UUID(),
							RuleCategory:    "PCT_MERCHANT",
							RuleType:        cardcontrols.ControlType_MCT_ALCOHOL.String(),
						},
					},
				},
			},
			want:               "A transaction of $0.00 (Lime) was declined because of a control you placed on your card ending in 1234",
			expectedPubSubSend: 1,
		},
		{
			name: "happy path not aud",
			request: &ncpb.Request{
				TransactionDetails: &ncpb.TransactionDetails{
					UserIdentifier:           aPersonaID,
					BillerCurrencyCode:       "840",
					RequestReceivedTimeStamp: "12345",
					PrimaryAccountNumber:     "1234123412341234",
					MerchantInfo: &ncpb.MerchantInfo{
						Name:                 "Lime",
						CountryCode:          "USA",
						MerchantCategoryCode: "1234",
						CurrencyCode:         "USD",
					},
				},
				TransactionOutcome: &ncpb.TransactionOutcome{
					DecisionId:                "123",
					NotificationId:            "abc123",
					TransactionApproved:       "DECLINED",
					DecisionResponseTimeStamp: "1234",
					AlertDetails: []*ncpb.AlertDetails{
						{
							TriggeringAppId: gofakeit.UUID(),
							RuleCategory:    "PCT_MERCHANT",
							RuleType:        cardcontrols.ControlType_MCT_ALCOHOL.String(),
						},
					},
				},
			},
			want:               "A transaction of USD0.00 (Lime) was declined because of a control you placed on your card ending in 1234",
			expectedPubSubSend: 1,
		},
		{
			name: "happy path without merchant info",
			request: &ncpb.Request{
				TransactionDetails: &ncpb.TransactionDetails{
					UserIdentifier:           aPersonaID,
					BillerCurrencyCode:       "840",
					RequestReceivedTimeStamp: "12345",
					PrimaryAccountNumber:     "1234123412341234",
				},
				TransactionOutcome: &ncpb.TransactionOutcome{
					DecisionId:                "123",
					NotificationId:            "abc123",
					TransactionApproved:       "DECLINED",
					DecisionResponseTimeStamp: "1234",
					AlertDetails: []*ncpb.AlertDetails{
						{
							TriggeringAppId: gofakeit.UUID(),
							RuleCategory:    "PCT_MERCHANT",
							RuleType:        cardcontrols.ControlType_MCT_ALCOHOL.String(),
						},
					},
				},
			},
			want:               "A transaction of USD0.00 was declined because of a control you placed on your card ending in 1234",
			expectedPubSubSend: 1,
		},
		{
			name: "fails with invalid country code",
			request: &ncpb.Request{
				TransactionDetails: &ncpb.TransactionDetails{
					UserIdentifier:           aPersonaID,
					BillerCurrencyCode:       "FOO",
					RequestReceivedTimeStamp: "12345",
					PrimaryAccountNumber:     "1234123412341234",
					MerchantInfo: &ncpb.MerchantInfo{
						Name:                 "Lime",
						CountryCode:          "AUD",
						MerchantCategoryCode: "1234",
						CurrencyCode:         "AUD",
					},
				},
				TransactionOutcome: &ncpb.TransactionOutcome{
					DecisionId:                "123",
					NotificationId:            "abc123",
					TransactionApproved:       "DECLINED",
					DecisionResponseTimeStamp: "1234",
					AlertDetails: []*ncpb.AlertDetails{
						{
							TriggeringAppId: gofakeit.UUID(),
							RuleCategory:    "PCT_MERCHANT",
							RuleType:        cardcontrols.ControlType_MCT_ALCOHOL.String(),
						},
					},
				},
			},
			expectedError: "invalid currency code: FOO",
		},
		{
			name: "fails when user ID not provided",
			request: &ncpb.Request{
				TransactionDetails: &ncpb.TransactionDetails{
					BillerCurrencyCode:       "975",
					RequestReceivedTimeStamp: "12345",
					PrimaryAccountNumber:     "1234123412341234",
					MerchantInfo: &ncpb.MerchantInfo{
						Name:                 "Lime",
						CountryCode:          "AUD",
						MerchantCategoryCode: "1234",
						CurrencyCode:         "AUD",
					},
				},
				TransactionOutcome: &ncpb.TransactionOutcome{
					DecisionId:                "123",
					NotificationId:            "abc123",
					TransactionApproved:       "DECLINED",
					DecisionResponseTimeStamp: "1234",
					AlertDetails: []*ncpb.AlertDetails{
						{
							TriggeringAppId: gofakeit.UUID(),
							RuleCategory:    "PCT_MERCHANT",
							RuleType:        cardcontrols.ControlType_MCT_ALCOHOL.String(),
						},
					},
				},
			},
			expectedError: "no user identifier present visa notification callback request",
		},
		{
			name: "nothing done when transaction approved",
			request: &ncpb.Request{
				TransactionDetails: &ncpb.TransactionDetails{
					UserIdentifier:           aPersonaID,
					BillerCurrencyCode:       "840",
					RequestReceivedTimeStamp: "12345",
					MerchantInfo: &ncpb.MerchantInfo{
						Name:                 "Lime",
						CountryCode:          "AUD",
						MerchantCategoryCode: "1234",
						CurrencyCode:         "AUD",
					},
				},
				TransactionOutcome: &ncpb.TransactionOutcome{
					DecisionId:                "123",
					NotificationId:            "abc123",
					TransactionApproved:       "APPROVED",
					DecisionResponseTimeStamp: "1234",
					AlertDetails: []*ncpb.AlertDetails{
						{
							TriggeringAppId: gofakeit.UUID(),
							RuleCategory:    "PCT_MERCHANT",
							RuleType:        cardcontrols.ControlType_MCT_ALCOHOL.String(),
						},
					},
				},
			},
		},
		{
			name: "rejects transaction with no payment token",
			request: &ncpb.Request{
				TransactionDetails: &ncpb.TransactionDetails{
					UserIdentifier:           aPersonaID,
					BillerCurrencyCode:       "036",
					RequestReceivedTimeStamp: "12345",
					MerchantInfo: &ncpb.MerchantInfo{
						Name:                 "Lime",
						CountryCode:          "AUD",
						MerchantCategoryCode: "1234",
						CurrencyCode:         "AUD",
					},
				},
				TransactionOutcome: &ncpb.TransactionOutcome{
					DecisionId:                "123",
					NotificationId:            "abc123",
					TransactionApproved:       "DECLINED",
					DecisionResponseTimeStamp: "1234",
					AlertDetails: []*ncpb.AlertDetails{
						{
							TriggeringAppId: gofakeit.UUID(),
							RuleCategory:    "PCT_MERCHANT",
							RuleType:        cardcontrols.ControlType_MCT_ALCOHOL.String(),
						},
					},
				},
			},
			expectedError: "this transaction was not associated with a valid card",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if isValid := test.request.Validate(); isValid != nil {
				t.Skipf("request was invalid: %s", isValid.Error())
			}

			require.NoError(t, feature.FeatureGate.Set(map[feature.Feature]bool{
				feature.NotificationCallbackDeclinedEvent: true,
			}))

			cc := cc.NewFakePublisher()
			s := &server{
				CommandCentre: &cc,
			}

			_, err := s.Alert(context.Background(), test.request)
			if test.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, cc.GetLastMessage())
			}

			require.Equal(t, test.expectedPubSubSend, cc.Count)
		})
	}
}

func TestTransactionDeclinedNotification(t *testing.T) {
	tests := []struct {
		name             string
		value            float32
		maskedCardNumber string
		expected         string
		merchantName     string
		controlType      cardcontrols.ControlType
	}{
		{
			name:             "nice values",
			value:            12.34,
			maskedCardNumber: "************1234",
			expected:         "A transaction of AUD12.34 was declined because of a control you placed on your card ending in 1234",
		},
		{
			name:             "long decimal tail truncated",
			value:            4.567891011,
			maskedCardNumber: "************6789",
			expected:         "A transaction of AUD4.57 was declined because of a control you placed on your card ending in 6789",
		},
		{
			name:             "big values still have correct decimal place",
			value:            17000.87,
			maskedCardNumber: "************3333",
			expected:         "A transaction of AUD17000.87 was declined because of a control you placed on your card ending in 3333",
		},
		{
			name:             "with merchant name has different message",
			value:            1200,
			maskedCardNumber: "************5569",
			merchantName:     "Generic Shop",
			expected:         "A transaction of AUD1200.00 (Generic Shop) was declined because of a control you placed on your card ending in 5569",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := transactionDeclinedNotification(context.Background(), "1233", "AUD", test.value, test.maskedCardNumber, test.merchantName)
			longTitle := res.Preview.Body
			require.Equal(t, test.expected, longTitle)
		})
	}
}

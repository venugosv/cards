package ctm

import (
	"context"
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	testUtil "github.com/anzx/fabric-cards/test/util"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/util/testutil"

	"github.com/anzx/fabric-cards/pkg/integration/util"

	"github.com/stretchr/testify/assert"
)

const (
	cardNumber          = "4514170000000001"
	tokenizedCardNumber = "123456789009876"
	key                 = "ClientIDKey"
)

func TestCardStatus_String(t *testing.T) {
	tests := []struct {
		name       string
		cardStatus Status
		want       string
	}{
		{
			name:       "Status Closed",
			cardStatus: StatusClosed,
			want:       "Closed",
		}, {
			name:       "Status Delinquent (Return)",
			cardStatus: StatusDelinquentReturn,
			want:       "Delinquent (Return Card)",
		}, {
			name:       "Status Delinquent (Retain)",
			cardStatus: StatusDelinquentRetain,
			want:       "Delinquent (Retain Card)",
		}, {
			name:       "Status Issued",
			cardStatus: StatusIssued,
			want:       "Issued",
		}, {
			name:       "Status Lost",
			cardStatus: StatusLost,
			want:       "Lost",
		}, {
			name:       "Status Stolen",
			cardStatus: StatusStolen,
			want:       "Stolen",
		}, {
			name:       "Status Unissued",
			cardStatus: StatusUnissuedNdIciCards,
			want:       "Unissued (N&D ICI Cards)",
		}, {
			name:       "Status Temporary",
			cardStatus: StatusTemporaryBlock,
			want:       "Temporary Block",
		}, {
			name:       "Status ATM Block",
			cardStatus: StatusBlockAtm,
			want:       "Block ATM",
		}, {
			name:       "Status Block ATM & POS (Exclude CNP)",
			cardStatus: StatusBlockAtmPosExcludeCnp,
			want:       "Block ATM & POS (Exclude CNP)",
		}, {
			name:       "Status Block ATM, POS, CNP & BCH",
			cardStatus: StatusBlockAtmPosCnpBch,
			want:       "Block ATM, POS, CNP & BCH",
		}, {
			name:       "Status Block ATM, POS & CNP",
			cardStatus: StatusBlockAtmPosCnp,
			want:       "Block ATM, POS & CNP",
		}, {
			name:       "Status Block CNP",
			cardStatus: StatusBlockCnp,
			want:       "Block CNP",
		}, {
			name:       "Status Block POS (exclude CNP)",
			cardStatus: StatusBlockPosExcludeCnp,
			want:       "Block POS (exclude CNP)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cardStatus.String())
		})
	}
}

func TestPointer(t *testing.T) {
	t.Run("", func(t *testing.T) {
		want := func(reason StatusReason) *StatusReason { return &reason }(StatusReasonClosed)
		got := StatusReasonClosed.Pointer()
		assert.Equal(t, want, got)
	})
}

func TestClient_DebitCardInquiry(t *testing.T) {
	tests := []struct {
		name                string
		tokenizedCardNumber string
		want                *DebitCardResponse
		mockAPIc            apic.Clienter
		wantErr             error
		wantCause           string
	}{
		{
			name:                "successfully make card inquiry call",
			tokenizedCardNumber: tokenizedCardNumber,
			want:                &DebitCardResponse{Title: "MS", FirstName: "FFFF", LastName: "FFFF", ProductCode: "PDV", SubProductCode: "031", StatusCode: "L", Status: "Lost", AccountsLinkedCount: 2, ExpiryDate: "1707", IssueReason: "03", ActivationStatus: true, Limits: []NewLimits{{Type: "ATMEFTPOS", DailyLimit: 1000, DailyLimitAvailable: 887, LastTransaction: "2015-08-25"}}, EmbossingLine1: "MS FFFFF", EmbossingLine2: "FFFF", ClosedDate: util.ToStringPtr("2014-05-25"), StatusReason: "With PIN or Account Related", TotalCards: 4, DispatchedMethod: "Mail", ReplacementCount: 2, IssueBranch: 3011, CollectionBranch: 3352, CollectionStatus: "Card Collected", ReplacedDate: "2010-11-12", ReissueDate: "2014-06-06", IssueDate: "2004-05-25", CardNumber: Card{Token: "4512222222222222", Last4Digits: "0986"}, OldCardNumber: &Card{Token: "3333333333333333", Last4Digits: "6548"}, NewCardNumber: &Card{Token: "2222222222222222", Last4Digits: "4466"}, PrevExpiryDate: "201407", PinChangeDate: "2012-05-27", PinChangedCount: 3, LastPinFailed: "2014-09-02", PinFailedCount: 8, StatusChangedDate: "2018-09-02", DetailsChangedDate: "2016-10-20", StatusChangedUserID: "X794139", DetailsChangedUserID: "X794139", FeeWaivedReason: "", DesignCode: "905", DesignColor: "blue", MerchantUpdatePreference: true, CardControlPreference: true, Wallets: Wallet{ApplePay: 2, GooglePay: 1, SamsungPay: 0, Fitness: 0, ECommerce: 0, Other: 0}},
			mockAPIc:            testUtil.MockAPIcer{Response: mockSuccessfulNewDebitCardInquiryResponse},
		},
		{
			name:                "handle unexpected body from downstream",
			tokenizedCardNumber: tokenizedCardNumber,
			wantErr:             errors.New("fabric error: status_code=Internal, error_code=2, message=failed request, reason=unexpected response from downstream"),
			mockAPIc:            testUtil.MockAPIcer{Response: []byte(`%%`)},
		},
		{
			name:                "handle 400",
			tokenizedCardNumber: tokenizedCardNumber,
			wantErr:             errors.New("failed request"),
			mockAPIc:            testUtil.MockAPIcer{ResponseErr: errors.New("failed request")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := c.DebitCardInquiry(testutil.GetContext(true), tt.tokenizedCardNumber)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_NewDebitCardInquiryError(t *testing.T) {
	t.Run("unable to access server", func(t *testing.T) {
		c := &client{
			apicClient: testUtil.MockAPIcer{ResponseErr: errors.New("failed request")},
		}

		ctx := testutil.GetContext(true)
		got, err := c.DebitCardInquiry(ctx, tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.Nil(t, got)
	})
	t.Run("no url provided in config", func(t *testing.T) {
		config := &Config{
			BaseURL:        "%",
			ClientIDEnvKey: key,
		}

		got, err := ClientFromConfig(context.Background(), nil, config, nil)
		require.Error(t, err)
		assert.Nil(t, got)
	})
	t.Run("unable to unmarshal response body", func(t *testing.T) {
		c := &client{
			apicClient: testUtil.MockAPIcer{Response: []byte(`%%`)},
		}

		want := errors.New("fabric error: status_code=Internal, error_code=2, message=failed request, reason=unexpected response from downstream")
		got, err := c.DebitCardInquiry(testutil.GetContext(true), tokenizedCardNumber)
		assert.NotNil(t, err)
		assert.Equal(t, want.Error(), err.Error())
		assert.Nil(t, got)
	})
}

var mockSuccessfulNewDebitCardInquiryResponse = []byte(`{
  "title": "MS",
  "firstName": "FFFF",
  "lastName": "FFFF",
  "productCode": "PDV",
  "subProductCode": "031",
  "statusCode": "L",
  "status": "Lost",
  "accountsLinkedCount": 2,
  "expiryDate": "1707",
  "issueReason": "03",
  "activationStatus": true,
  "limits": [
    {
      "type": "ATMEFTPOS",
      "dailyLimit": 1000,
      "dailyLimitAvailable": 887,
      "lastTransaction": "2015-08-25"
    }
  ],
  "embossingLine1": "MS FFFFF",
  "embossingLine2": "FFFF",
  "closedDate": "2014-05-25",
  "statusReason": "With PIN or Account Related",
  "totalCards": 4,
  "dispatchedMethod": "Mail",
  "replacementCount": 2,
  "issueBranch": 3011,
  "collectionBranch": 3352,
  "collectionStatus": "Card Collected",
  "replacedDate": "2010-11-12",
  "reissueDate": "2014-06-06",
  "issueDate": "2004-05-25",
  "cardNumber": {
    "token": "4512222222222222",
    "last4digits": "0986"
  },
  "oldCardNumber": {
    "token": "3333333333333333",
    "last4digits": "6548"
  },
  "newCardNumber": {
    "token": "2222222222222222",
    "last4digits": "4466"
  },
  "prevExpiryDate": "201407",
  "pinChangeDate": "2012-05-27",
  "pinChangedCount": 3,
  "lastPinFailed": "2014-09-02",
  "pinFailedCount": 8,
  "statusChangedDate": "2018-09-02",
  "detailsChangedDate": "2016-10-20",
  "statusChangedUserID": "X794139",
  "detailsChangedUserID": "X794139",
  "feeWaivedReason": "",
  "designCode": "905",
  "designColor": "blue",
  "merchantUpdatePreference": true,
  "cardControlPreference": true,
  "wallets": {
      "applePay": 2,
      "googlePay": 1,
      "samsungPay": 0,
	  "fitness": 0,
	  "eCommerce": 0,
      "other": 0
    }
}`)

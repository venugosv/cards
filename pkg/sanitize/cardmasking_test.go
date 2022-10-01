package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskCardNumbersInString(t *testing.T) {
	t.Run("successfully mask card number in body", func(t *testing.T) {
		want := "{\"primaryAccountNumber\":\"451417******0001\"}"
		got := MaskCardNumbersInString("{\"primaryAccountNumber\":\"4514170000000001\"}")
		assert.Equal(t, want, got)
	})
	t.Run("successfully mask multiple card numbers in body", func(t *testing.T) {
		want := "{\"primaryAccountNumber\":\"451417******0001\",\"primaryAccountNumber\":\"451417******0001\"}"
		got := MaskCardNumbersInString("{\"primaryAccountNumber\":\"4514170000000001\",\"primaryAccountNumber\":\"4514170000000001\"}")
		assert.Equal(t, want, got)
	})
	t.Run("successfully mask card number mid string", func(t *testing.T) {
		want := "qffd123456******3456dfghk123456786543sdfg34567dfgh456"
		got := MaskCardNumbersInString("qffd1234567890123456dfghk123456786543sdfg34567dfgh456")
		assert.Equal(t, want, got)
	})
	t.Run("Test Hashed Card Number", func(t *testing.T) {
		want := "oJVoTHSm5rn/Ru/tYUZL93767XLr6Ag1v1R6vwdti/c="
		got := MaskCardNumbersInString("oJVoTHSm5rn/Ru/tYUZL93767XLr6Ag1v1R6vwdti/c=")
		assert.Equal(t, want, got)
	})
	t.Run("successfully mask alone card number", func(t *testing.T) {
		want := "123456******3456"
		got := MaskCardNumbersInString("1234567890123456")
		assert.Equal(t, want, got)
	})
	t.Run("successfully handle nil card number", func(t *testing.T) {
		want := ""
		got := MaskCardNumbersInString("")
		assert.Equal(t, want, got)
	})
	t.Run("successfully handle nil card number", func(t *testing.T) {
		want := `{"accountsLinkedCount": 2,"activationStatus": true,"collectionBranch": 4672,"collectionStatus": "Card NOT Collected","dispatchedMethod": "Sent to Branch","expiryDate": "1705","issueBranch": 4672,"issueDate": "2015-08-05","replacementCount": 0,"status": "%s","detailsChangedDate": "2015-08-05","embossedName": "MR NATHAN FUKUSHIMA","firstName": "NATHAN","lastName": "FUKUSHIMA","limits": [	{	"dailyLimit": 1000,	"dailyLimitAvailable": 1000,	"type": "POS"	},	{	"dailyLimit": 1000,	"dailyLimitAvailable": 1000,	"type": "ATM"	}]}`
		got := MaskCardNumbersInString(want)
		assert.Equal(t, want, got)
	})
	t.Run("less than 16 chars", func(t *testing.T) {
		want := "170000000001"
		got := MaskCardNumbersInString(want)
		assert.Equal(t, want, got)
	})
}

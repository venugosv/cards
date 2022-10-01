package config

import (
	"testing"
)

type TestName string

const (
	V1beta1CardAPIReplaceDamaged             TestName = "v1beta1.CardAPI/ReplaceDamaged"
	V1beta1CardAPIList                       TestName = "v1beta1.CardAPI/List"
	V1beta1CardAPIReplaceLost                TestName = "v1beta1.CardAPI/ReplaceLost"
	V1beta1CardAPIActivate                   TestName = "v1beta1.CardAPI/Activate"
	V1beta1CardAPIGetWrappingKey             TestName = "v1beta1.CardAPI/GetWrappingKey"
	V1beta1CardAPISetPIN                     TestName = "v1beta1.CardAPI/SetPIN"
	V1beta1CardAPIChangePIN                  TestName = "v1beta1.CardAPI/ChangePIN"
	V1beta1CardAPIGetDetails                 TestName = "v1beta1.CardAPI/GetDetails"
	V1beta1CardAPIAuditTrail                 TestName = "v1beta1.CardAPI/AuditTrail"
	V1beta1WalletAPICreateApplePaymentToken  TestName = "v1beta1.WalletAPI/CreateApplePaymentToken"  //nolint:gosec
	V1beta1WalletAPICreateGooglePaymentToken TestName = "v1beta1.WalletAPI/CreateGooglePaymentToken" //nolint:gosec

	V1beta2CardControlsAPIListControls   TestName = "v1beta2.CardControlsAPI/ListControls"
	V1beta2CardControlsAPIQueryControls  TestName = "v1beta2.CardControlsAPI/QueryControls"
	V1beta2CardControlsAPISetControls    TestName = "v1beta2.CardControlsAPI/SetControls"
	V1beta2CardControlsAPIRemoveControls TestName = "v1beta2.CardControlsAPI/RemoveControls"
	V1beta2CardControlsAPIBlockCard      TestName = "v1beta2.CardControlsAPI/BlockCard"
	V1beta2CardControlsAPIUnBlockCard    TestName = "v1beta2.CardControlsAPI/UnBlockCard"
)

type Toggle struct {
	rules map[TestName]bool
}

func NewToggle(rules map[TestName]bool) *Toggle {
	return &Toggle{rules: rules}
}

func (c *Toggle) Skip(t *testing.T, name TestName) {
	if !c.rules[name] {
		t.Skip(name, "test disabled by config")
	}
}

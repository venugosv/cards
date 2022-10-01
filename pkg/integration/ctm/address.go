package ctm

import (
	"context"
	"strings"

	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"
)

const (
	maxCardMailingAddressLength = 36
	aus                         = "AUS"
)

func translateAUISOState(in string) string {
	out, ok := ocvToCapStateTranslation[in]
	if !ok {
		return in
	}
	return out
}

var ocvToCapStateTranslation = map[string]string{
	"AU-CT": "ACT",
	"AU-NS": "NSW",
	"AU-NT": "NT",
	"AU-QL": "QLD",
	"AU-SA": "SA",
	"AU-TS": "TAS",
	"AU-VI": "VIC",
	"AU-WA": "WA",
}

// GetAddress rules are documented in https://confluence.service.anz/display/ABT/JOIN%3A+Account+Origination+Rules
func GetAddress(ctx context.Context, in *sspb.Address) (MailingAddress, error) {
	if isAustralianAddress(in) {
		return MailingAddress{}, anzerrors.New(codes.FailedPrecondition, "Invalid Address",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardInvalidAddress, "address on customer profile is international"))
	}

	out := MailingAddress{
		AddressLine1: in.LineOne,
		AddressLine2: in.LineTwo,
		AddressLine3: in.LineThree,
		PostCode:     in.PostalCode,
	}

	switch {
	case in.LineTwo == "":
		out.AddressLine2 = toCityAndState(in)
	case in.LineThree == "":
		out.AddressLine3 = toCityAndState(in)
	default:
		out.AddressLine2 = out.AddressLine2 + " " + out.AddressLine3
		out.AddressLine3 = toCityAndState(in)
	}

	out.truncate(maxCardMailingAddressLength)
	out.trimWhitespace()

	return out, nil
}

func isAustralianAddress(in *sspb.Address) bool {
	return in.GetCountry() != aus
}

func toCityAndState(in *sspb.Address) string {
	return JoinNonBlank([]string{" "}, in.City, translateAUISOState(in.State), in.PostalCode)
}

func (m *MailingAddress) truncate(max int) {
	m.AddressLine1 = truncate(m.AddressLine1, max)
	m.AddressLine2 = truncate(m.AddressLine2, max)
	m.AddressLine3 = truncate(m.AddressLine3, max)
}

func (m *MailingAddress) trimWhitespace() {
	m.AddressLine1 = strings.TrimSpace(m.AddressLine1)
	m.AddressLine2 = strings.TrimSpace(m.AddressLine2)
	m.AddressLine3 = strings.TrimSpace(m.AddressLine3)
}

func truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length]
}

// JoinNonBlank
// Blank strings are ignored when concatenating
// Separators apply to the term that follows the separator, and is only used if that term and the previous term are non-blank
// Separator usage wraps around if quantity of gaps exceeds separators
func JoinNonBlank(separators []string, strs ...string) string {
	result := ""
	for i, str := range strs {
		if len(str) == 0 {
			continue
		}
		if len(result) > 0 {
			result += separators[(i-1)%len(separators)]
		}
		result += str
	}
	return result
}

package sanitize

import (
	"regexp"
)

// MaskCardNumbersInString checks presence of numbers more than 9 digit and masks all character except last 4 digits of number.
// cardNumbersString String to be identified for PCI compliance and masked accordingly.
// returns Masked sting with masking character '*'.
func MaskCardNumbersInString(in string) string {
	if in == "" {
		return ""
	}

	return string(MaskCardNumbersInByte([]byte(in)))
}

func MaskCardNumbersInByte(in []byte) []byte {
	return regexp.MustCompile(`(\d{6})(\d{6})(\d{4})`).ReplaceAll(in, []byte("$1******$3"))
}

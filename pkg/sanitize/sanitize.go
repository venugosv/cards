package sanitize

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const exp = ".*(?i)(token|email|phone|address|jwt|pin|name|auth|cvv|client-id|ccv|emboss|accountNumber|balance|password|secret).*"

var sensitiveData = regexp.MustCompile(exp)

// ConvertToLoggableFieldValue will convert the data into a map[string]interface if possible, otherwise will convert to a string
func ConvertToLoggableFieldValue(data []byte) interface{} {
	return ConvertToLoggableFieldValueWithType(data, "")
}

func ConvertToLoggableFieldValueWithType(data []byte, contentType string) interface{} {
	maskedData := MaskCardNumbersInByte(data)

	// This is likely a urlencoded request or response body
	if strings.Contains(contentType, "x-www-form-urlencoded") {
		return parseAndRedactUrlEncoded(string(maskedData))
	}

	result, err := parseAndRedactJson(maskedData)
	// Unable to convert to JSON
	if err != nil {
		return string(maskedData)
	}
	return result
}

// parseAndRedactUrlEncoded takes a x-www-form-urlencoded string,
// parses it and sets all values with sensitive keys to *
func parseAndRedactUrlEncoded(input string) string {
	result, err := url.ParseQuery(input)
	// If parsing failed, return the original (masked) string
	if err != nil {
		return input
	}
	for key := range result {
		if sensitiveData.MatchString(key) {
			result.Set(key, "*")
		}
	}
	// Convert back to urlencoded format
	output := result.Encode()
	// This is necessary as the builtin url encode function will escape * to %2A
	return strings.ReplaceAll(output, "%2A", "*")
}

func parseAndRedactJson(data []byte) (interface{}, error) {
	// Try parsing as JSON array
	arrayResult := make([]interface{}, 0)
	err := json.Unmarshal(data, &arrayResult)
	if err == nil {
		redactInterface(arrayResult)
		return arrayResult, nil
	}
	// Try parsing as JSON map
	mapResult := make(map[string]interface{})
	err = json.Unmarshal(data, &mapResult)
	if err == nil {
		redactMap(mapResult)
		return mapResult, nil
	}
	return nil, err
}

// redactMap recursively replace values with `*` if the key is in the blockList
func redactMap(data map[string]interface{}) {
	for key := range data {
		if sensitiveData.MatchString(key) {
			data[key] = "*"
		} else if m, ok := data[key].(map[string]interface{}); ok {
			redactMap(m)
		} else if i, ok := data[key].([]interface{}); ok {
			redactInterface(i)
		}
	}
}

// redactInterface provides a method of recursively stepping through
// objects and identifying key, value pairs for sanitizing
func redactInterface(data []interface{}) {
	for key := range data {
		if m, ok := data[key].(map[string]interface{}); ok {
			redactMap(m)
		} else if i, ok := data[key].([]interface{}); ok {
			redactInterface(i)
		}
	}
}

// GetHeaders will convert the http.Header into a map[string]interface and
// redact data if the key is found in the blockList
func GetHeaders(headers http.Header) map[string]interface{} {
	requestHeaders := make(map[string]interface{})
	for key, val := range headers {
		if sensitiveData.MatchString(key) {
			requestHeaders[key] = "*"
			continue
		}
		requestHeaders[key] = val
	}
	return requestHeaders
}

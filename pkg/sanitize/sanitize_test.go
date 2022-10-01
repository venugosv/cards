package sanitize

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const UrlEncoding = "application/x-www-form-urlencoded"

func TestGetHeaders(t *testing.T) {
	t.Run("jwt header masked successfully", func(t *testing.T) {
		in := http.Header{
			"jwt": []string{"Bearer", "qwertyuiop"},
		}
		want := map[string]interface{}{
			"jwt": "*",
		}
		got := GetHeaders(in)

		assert.Equal(t, want, got)
	})
	t.Run("no masking required", func(t *testing.T) {
		in := http.Header{
			"notmasked": []string{"qwertyuiop"},
		}
		want := map[string]interface{}{
			"notmasked": []string{"qwertyuiop"},
		}
		got := GetHeaders(in)

		assert.Equal(t, want, got)
	})
}

func TestRedactMap(t *testing.T) {
	test := struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		name: "test",
		in: map[string]interface{}{
			"firstname": "scary sensitive data",
			"firstName": "scary sensitive data",
			"lastname":  "scary sensitive data",
			"lastName":  "scary sensitive data",
			"person": map[string]interface{}{
				"firstname": "scary sensitive data",
				"firstName": "scary sensitive data",
				"lastname":  "scary sensitive data",
				"lastName":  "scary sensitive data",
			},
			"embossedname": "scary sensitive data",
			"daisys": map[string]interface{}{
				"token": "scary sensitive data",
				"Token": map[string]interface{}{
					"token": []interface{}{
						map[string]interface{}{
							"token": "scary sensitive data",
						},
						map[string]interface{}{
							"token": "scary sensitive data",
						},
					},
				},
			},
			"embossedName":         "scary sensitive data",
			"name":                 "scary sensitive data",
			"Name":                 "scary sensitive data",
			"token":                "scary sensitive data",
			"Token":                "scary sensitive data",
			"X-Vault-Token":        "scary sensitive data",
			"cvv":                  "scary sensitive data",
			"CVV":                  "scary sensitive data",
			"Cvv":                  "scary sensitive data",
			"ccv":                  "scary sensitive data",
			"Ccv":                  "scary sensitive data",
			"CCV":                  "scary sensitive data",
			"jwt":                  "scary sensitive data",
			"JWT":                  "scary sensitive data",
			"jwtAuth":              "scary sensitive data",
			"authorization":        "scary sensitive data",
			"Authorization":        "scary sensitive data",
			"wrappingkey":          "scary sensitive data",
			"wrappingKey":          "scary sensitive data",
			"encryptedpinblock":    "scary sensitive data",
			"encryptedPINBlock":    "scary sensitive data",
			"encryptednewpinblock": "scary sensitive data",
			"encryptedNewPINBlock": "scary sensitive data",
			"encryptedoldpinblock": "scary sensitive data",
			"encryptedOldPINBlock": "scary sensitive data",
			"X-Ibm-Client-Id":      "scary sensitive data",
			"embossingLine1":       "scary sensitive data",
			"email":                "scary sensitive data",
			"emailAddress":         "scary sensitive data",
			"addressLine1":         "scary sensitive data",
			"addressLine2":         "scary sensitive data",
			"addressLine3":         "scary sensitive data",
			"address":              "scary sensitive data",
			"phone":                "scary sensitive data",
			"phoneNumber":          "scary sensitive data",
			"legalName":            "scary sensitive data",
		},
	}
	t.Run(test.name, func(t *testing.T) {
		redactMap(test.in)
		assert.NotContains(t, test.in, "scary")
	})
}

func TestRedactInterface(t *testing.T) {
	test := struct {
		name string
		in   []interface{}
		want []interface{}
	}{
		name: "test",
		in: []interface{}{
			map[string]interface{}{
				"token": "scary sensitive data",
			},
			map[string]interface{}{
				"token": "scary sensitive data",
			},
			[]interface{}{
				map[string]interface{}{
					"token": "scary sensitive data",
				},
				map[string]interface{}{
					"token": "scary sensitive data",
				},
			},
		},
	}
	t.Run(test.name, func(t *testing.T) {
		redactInterface(test.in)
		assert.NotContains(t, test.in, "scary")
	})
}

func TestConvertToLoggableFieldValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		contentType string
		payload     []byte
		expected    string
		notContains string
	}{
		{
			name:        "test cardnumber removed from raw bytes",
			payload:     []byte(`1234567890123456`),
			expected:    "123456******3456",
			notContains: "1234567890123456",
		},
		{
			name:        "test cardnumber removed from json",
			payload:     []byte(`{"cardNumber": "1234567890123456"}`),
			notContains: "1234567890123456",
		},
		{
			name:        "test cardnumber removed from invalid json",
			payload:     []byte(`%%1234567890123456`),
			expected:    "%%123456******3456",
			notContains: "1234567890123456",
		},
		{
			name:        "test json arrays and nested maps sanitised",
			payload:     []byte(`{"pin":"1234","pins":["1234", "1234"],"obj":{"pin1":1234,"pin2":"1234","pins":[1234, 1234]}}`),
			notContains: "1234",
		},
		{
			name:        "test top level JSON array sanitised",
			payload:     []byte(`[{"address": "1234", "timestamp": 5555}]`),
			notContains: "1234",
		},
		{
			name:        "test x-www-form-urlencoded sanitised",
			payload:     []byte(`pin=1234&test=test`),
			contentType: UrlEncoding,
			expected:    "pin=*&test=test",
			notContains: "1234",
		},
		{
			name:        "test x-www-form-urlencoded sanitised with scopes",
			payload:     []byte(`scopes=scope1+scope2+scope3&secret_key=5678`),
			contentType: UrlEncoding,
			expected:    "scopes=scope1+scope2+scope3&secret_key=*",
			notContains: "5678",
		},
		{
			name:        "test x-www-form-urlencoded sanitised with arrays",
			payload:     []byte(`scope=scope1&scope=scope2&secret_key=s1&secret_key=s2`),
			contentType: UrlEncoding,
			expected:    "scope=scope1&scope=scope2&secret_key=*",
			notContains: "s1",
		},
		{
			name:        "test x-www-form-urlencoded sanitised with *",
			payload:     []byte(`scope=*&scope=scope2&secret_key=*&secret_key=s2`),
			contentType: UrlEncoding,
			expected:    "scope=*&scope=scope2&secret_key=*",
			notContains: "s2",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := fmt.Sprint(ConvertToLoggableFieldValueWithType(test.payload, test.contentType))
			assert.NotContains(t, actual, test.notContains)
			if test.expected != "" {
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

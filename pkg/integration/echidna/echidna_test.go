package echidna

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	testUtil "github.com/anzx/fabric-cards/test/util"
	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	"github.com/anzx/fabric-cards/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
)

const (
	encryptedPINBlockOld = "h6VXDX/10a4s4EF2h947f55/xCPiADzEcOju1ZnPIrtjEIwejhEKlIfURWLE7vajffhZ/YNFdQV/fG9aQRALJQpos8K4w772PyNYumlML1vA6wtZninHJwaYJOGueX+N2gd63lsyhnwt5dezTNBhPDWcccXeRc3BK4nNWNkbm40Ng+zHCxSjn/Oqcav00kAIXBtGxfijU/8s7sAf8Ss9TwoVq1FG35ceFiWqfTotH4wD8X7BwH2N7WN3KYJMOeA9myVffdqnv9QNk/ldq/Nvou3AaH2qUilqORyzErBugSfoAeCe9u2/BAzKZeh1q7USKhO5SzfNGb+WwY8Ui1+juQ=="
	encryptedPINBlockNew = "depZGnbxrHYgnFJYv78HAOMdetVjQxf6ZRLCfUkO5D7Fn8JGCtXvaZbYwTATTM6iFrhW9gk4Z+7Ow3aPqW+SwghKLpZaLdQHs8kwGGHYNysP43Gblgt3EagrigNcytinHTVBixj4cDVKy3UzlKDuoM42Fp81Qrw/YNuI2+HAokHuC1UP2bL5+nGmQ04dsjn0QEa4cTmhsaeaZxQNvNiFKibztjIfHVWvvKsztDc2rflme6KuVVvtXaS3Psls2GGg6fc3CfV+2rXQzvk/KnyVmgqvr0+NKONJf/PqrILyqwqEhREfC8bmtzz9x0GPffoozEdQl7mG6gTgRrYF7qXhqQ=="
	issuedCardNumber     = "4622390512341000"
	key                  = "EchidnaKey"
)

type mockSecretManager struct {
	name    string
	payload string
	err     error
}

func (m mockSecretManager) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name:    m.name,
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(m.payload)},
	}, m.err
}

func gsmClient() *gsm.Client {
	return &gsm.Client{
		SM: mockSecretManager{
			name:    "testName",
			payload: "redispassword",
		},
	}
}

func getResponse(action Action, code int) *Response {
	var (
		message    string
		encodedKey *string
	)

	switch action {
	case ActionGetWrappingKey:
		message = fmt.Sprintf("Get wrapping key %s", messages[code])
		encodedKey = util.ToStringPtr("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArI6WKTJMLVfpaG+Mkaj4IVX3/2dbtHvacI9sKfutMsg5It6pEvFf9oYoIWMQkxFARf14ds0+1t83sm6foPHm4HZ0oP2GX0iiFdALEZr3C6C2FXAoQQXYMGeczoeta0IwF75B3Pr6VETQjf7niL00MF0n/McsE9tu9VTOFjq6LkvZgOnBe9wG+f0nvdx29FAPzIjdpBoZ27Ingmtnmtk2T9oadY5vXE2ruIhjU2rL/8aPPN8LtvlWrcV0y+YW2l4EMGenAFYMu4jh6R5deNfartmNotJgbzHFcD7EpXJivzYgdMvea2Dy7AjlC5cic4ijcna750HhfMoFFNqf6T7psQIDAQAB")
	case ActionSelect:
		message = fmt.Sprintf("Select PIN %s", messages[code])
	case ActionVerify:
		message = fmt.Sprintf("Verify PIN %s", messages[code])
	case ActionChange:
		message = fmt.Sprintf("Change PIN %s", messages[code])
	}

	return &Response{
		Method: action,
		Result: Result{
			Code:       code,
			Message:    message,
			EncodedKey: encodedKey,
		},
		LogMessages: LogMessages{
			WantLevel: LoglevelInfo,
		},
	}
}

var messages = map[int]string{
	0:    "operation successful.",
	55:   "Incorrect PIN.",
	75:   "Maximum PIN tries exceeded.",
	1010: "operation failed due to Tandem error response.",
	1011: "operation failed due to request-response error.",
	1012: "RemotePIN service unavailable.",
	1013: "operation has timed out.",
	1015: "operation failed due to RemotePIN service error.",
}

func TestClientFromConfig(t *testing.T) {
	clientIDEnvKey := "THIS-IS-A-CLIENT-ID"
	gsmClient := gsmClient()

	t.Run("New Echidna with httpClient supplied", func(t *testing.T) {
		server := httptest.NewServer(nil)
		got, err := ClientFromConfig(context.Background(), server.Client(), &Config{
			ClientIDEnvKey: clientIDEnvKey,
		}, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("New Echidna without httpClient supplied", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, &Config{
			ClientIDEnvKey: clientIDEnvKey,
		}, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("New Echidna with bad baseURL", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, &Config{
			BaseURL:        "%%",
			ClientIDEnvKey: clientIDEnvKey,
		}, gsmClient)
		require.Error(t, err)
		assert.Nil(t, got)
	})
	t.Run("New Echidna with no config", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, nil, gsmClient)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
	t.Run("New Echidna with no client ID found", func(t *testing.T) {
		clientNoSecret := &gsm.Client{
			SM: mockSecretManager{
				name:    "testName",
				payload: "",
			},
		}

		got, err := ClientFromConfig(context.Background(), nil, &Config{
			ClientIDEnvKey: "notakey",
		}, clientNoSecret)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestBaseURLError(t *testing.T) {
	gsmClient := gsmClient()

	const name = "Fail to parse URL caught and handled"
	c, err := ClientFromConfig(context.Background(), nil, &Config{
		ClientIDEnvKey: key,
	}, gsmClient)
	require.NoError(t, err)
	t.Run(name, func(t *testing.T) {
		got, err := c.GetWrappingKey(testutil.GetContext(true))
		assert.Empty(t, got)
		assert.NotNil(t, err)
		assert.Error(t, err, "mandatory request fields not supplied")
	})
	t.Run(name, func(t *testing.T) {
		err := c.SelectPIN(testutil.GetContext(true), IncomingRequest{
			PlainPAN:          "1234567890123456",
			EncryptedPINBlock: "1234567890",
		})
		assert.NotNil(t, err)
		assert.Error(t, err, "mandatory request fields not supplied")
	})
	t.Run(name, func(t *testing.T) {
		err := c.VerifyPIN(testutil.GetContext(true), IncomingRequest{
			PlainPAN:          "1234567890123456",
			EncryptedPINBlock: "1234567890",
		})
		assert.NotNil(t, err)
		assert.Error(t, err, "mandatory request fields not supplied")
	})
	t.Run(name, func(t *testing.T) {
		err := c.ChangePIN(testutil.GetContext(true), IncomingChangePINRequest{
			PlainPAN:             "1234567890123456",
			EncryptedPINBlockNew: "1234567890",
			EncryptedPINBlockOld: "fghjk",
		})
		assert.NotNil(t, err)
		assert.Error(t, err, "mandatory request fields not supplied")
	})
}

func getBytesData(resp interface{}) []byte {
	data, _ := json.Marshal(resp)
	return data
}

func TestClient_SetPIN(t *testing.T) {
	tests := []struct {
		name           string
		request        IncomingRequest
		want           *Response
		mockAPIc       apic.Clienter
		wantErr        string
		requestHandler http.HandlerFunc
	}{
		{
			name: "successfully set pin",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			want:     getResponse(ActionSelect, 0),
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(SetPINResponse{Response: *getResponse(ActionSelect, 0)})},
		},
		{
			name: "1010 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed request, reason=Operation failed due to internal error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(SetPINResponse{Response: *getResponse(ActionSelect, 1010)})},
		},
		{
			name: "1011 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed request, reason=Operation failed due to internal error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(SetPINResponse{Response: *getResponse(ActionSelect, 1011)})},
		},
		{
			name: "1012 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Service unavailable",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(SetPINResponse{Response: *getResponse(ActionSelect, 1012)})},
		},
		{
			name: "1013 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=DeadlineExceeded, error_code=2, message=failed request, reason=Operation has timed out",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(SetPINResponse{Response: *getResponse(ActionSelect, 1013)})},
		},
		{
			name: "1015 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Operation failed due to service error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(SetPINResponse{Response: *getResponse(ActionSelect, 1015)})},
		},
		{
			name:    "Handle mandatory fields not supplied",
			request: IncomingRequest{},
			want:    nil,
			wantErr: "mandatory request fields not supplied",
		},
		{
			name: "Handle 500 error from downstream",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			want:     nil,
			wantErr:  "unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name: "fail to unmarshal body",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			mockAPIc: testUtil.MockAPIcer{Response: []byte{32}},
			wantErr:  "unexpected response from downstream",
		},
	}
	for _, test := range tests {
		echidna := &client{
			apicClient: test.mockAPIc,
		}

		t.Run(test.name, func(t *testing.T) {
			err := echidna.SelectPIN(testutil.GetContext(true), test.request)
			if test.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_VerifyPIN(t *testing.T) {
	tests := []struct {
		name           string
		request        IncomingRequest
		want           *Response
		mockAPIc       apic.Clienter
		wantErr        string
		requestHandler http.HandlerFunc
	}{
		{
			name: "successfully verify pin",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			want:     getResponse(ActionVerify, 0),
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(VerifyPINResponse{Response: *getResponse(ActionVerify, 0)})},
		},
		{
			name: "1010 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed request, reason=Operation failed due to internal error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(VerifyPINResponse{Response: *getResponse(ActionVerify, 1010)})},
		},
		{
			name: "1011 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed request, reason=Operation failed due to internal error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(VerifyPINResponse{Response: *getResponse(ActionVerify, 1011)})},
		},
		{
			name: "1012 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Service unavailable",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(VerifyPINResponse{Response: *getResponse(ActionVerify, 1012)})},
		},
		{
			name: "1013 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=DeadlineExceeded, error_code=2, message=failed request, reason=Operation has timed out",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(VerifyPINResponse{Response: *getResponse(ActionVerify, 1013)})},
		},
		{
			name: "1015 err code response",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: encryptedPINBlockOld,
			},
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Operation failed due to service error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(VerifyPINResponse{Response: *getResponse(ActionVerify, 1015)})},
		},
		{
			name:    "Handle mandatory fields not supplied",
			request: IncomingRequest{},
			want:    nil,
			wantErr: "mandatory request fields not supplied",
		},
		{
			name: "Handle 500 error from downstream",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: "depZGnbxrHYgnFJYv78HAOMdetVjQxf6ZRLCfUkO5D7Fn8JGCtXvaZbYwTATTM6iFrhW9gk4Z+7Ow3aPqW+SwghKLpZaLdQHs8kwGGHYNysP43Gblgt3EagrigNcytinHTVBixj4cDVKy3UzlKDuoM42Fp81Qrw/YNuI2+HAokHuC1UP2bL5+nGmQ04dsjn0QEa4cTmhsaeaZxQNvNiFKibztjIfHVWvvKsztDc2rflme6KuVVvtXaS3Psls2GGg6fc3CfV+2rXQzvk/KnyVmgqvr0+NKONJf/PqrILyqwqEhREfC8bmtzz9x0GPffoozEdQl7mG6gTgRrYF7qXhqQ==",
			},
			want:     nil,
			wantErr:  "unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name: "fail to unmarshal body",
			request: IncomingRequest{
				PlainPAN:          issuedCardNumber,
				EncryptedPINBlock: "depZGnbxrHYgnFJYv78HAOMdetVjQxf6ZRLCfUkO5D7Fn8JGCtXvaZbYwTATTM6iFrhW9gk4Z+7Ow3aPqW+SwghKLpZaLdQHs8kwGGHYNysP43Gblgt3EagrigNcytinHTVBixj4cDVKy3UzlKDuoM42Fp81Qrw/YNuI2+HAokHuC1UP2bL5+nGmQ04dsjn0QEa4cTmhsaeaZxQNvNiFKibztjIfHVWvvKsztDc2rflme6KuVVvtXaS3Psls2GGg6fc3CfV+2rXQzvk/KnyVmgqvr0+NKONJf/PqrILyqwqEhREfC8bmtzz9x0GPffoozEdQl7mG6gTgRrYF7qXhqQ==",
			},
			mockAPIc: testUtil.MockAPIcer{Response: []byte{32}},
			wantErr:  "unexpected response from downstream",
		},
	}
	for _, test := range tests {
		echidna := &client{
			apicClient: test.mockAPIc,
		}

		t.Run(test.name, func(t *testing.T) {
			err := echidna.VerifyPIN(testutil.GetContext(true), test.request)
			if test.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_ChangePIN(t *testing.T) {
	tests := []struct {
		name           string
		request        IncomingChangePINRequest
		want           *Response
		mockAPIc       apic.Clienter
		wantErr        string
		requestHandler http.HandlerFunc
	}{
		{
			name: "successfully change pin",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			want:     getResponse(ActionChange, 0),
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(ChangePINResponse{Response: *getResponse(ActionChange, 0)})},
		},
		{
			name: "fail to unmarshal body",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			mockAPIc: testUtil.MockAPIcer{Response: []byte{32}},
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte{32})
			},
			wantErr: "unexpected response from downstream",
		},
		{
			name: "1010 err code response",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed request, reason=Operation failed due to internal error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(ChangePINResponse{Response: *getResponse(ActionChange, 1010)})},
		},
		{
			name: "1011 err code response",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			wantErr:  "fabric error: status_code=Internal, error_code=2, message=failed request, reason=Operation failed due to internal error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(ChangePINResponse{Response: *getResponse(ActionChange, 1011)})},
		},
		{
			name: "1012 err code response",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Service unavailable",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(ChangePINResponse{Response: *getResponse(ActionChange, 1012)})},
		},
		{
			name: "1013 err code response",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			wantErr:  "fabric error: status_code=DeadlineExceeded, error_code=2, message=failed request, reason=Operation has timed out",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(ChangePINResponse{Response: *getResponse(ActionChange, 1013)})},
		},
		{
			name: "1015 err code response",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Operation failed due to service error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(ChangePINResponse{Response: *getResponse(ActionChange, 1015)})},
		},
		{
			name:    "Handle mandatory fields not supplied",
			request: IncomingChangePINRequest{},
			want:    nil,
			wantErr: "mandatory request fields not supplied",
		},
		{
			name: "Handle 500 error from downstream",
			request: IncomingChangePINRequest{
				PlainPAN:             issuedCardNumber,
				EncryptedPINBlockOld: encryptedPINBlockOld,
				EncryptedPINBlockNew: encryptedPINBlockNew,
			},
			want:     nil,
			wantErr:  "unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
	}
	for _, test := range tests {
		echidna := &client{
			apicClient: test.mockAPIc,
		}

		t.Run(test.name, func(t *testing.T) {
			err := echidna.ChangePIN(testutil.GetContext(true), test.request)
			if test.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_GetWrappingKey(t *testing.T) {
	tests := []struct {
		name           string
		want           string
		mockAPIc       apic.Clienter
		wantErr        string
		requestHandler http.HandlerFunc
	}{
		{
			name:     "successfully get wrapping key",
			want:     *getResponse(ActionGetWrappingKey, 0).Result.EncodedKey,
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(GetWrappingKeyResponse{Response: *getResponse(ActionGetWrappingKey, 0)})},
		},
		{
			name:     "1014 err code fails",
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Information is unavailable within the PIN service",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(GetWrappingKeyResponse{Response: *getResponse(ActionGetWrappingKey, 1014)})},
		},
		{
			name:     "1014 err code fails",
			wantErr:  "fabric error: status_code=Unavailable, error_code=2, message=failed request, reason=Operation failed due to service error",
			mockAPIc: testUtil.MockAPIcer{Response: getBytesData(GetWrappingKeyResponse{Response: *getResponse(ActionGetWrappingKey, 1015)})},
		},
		{
			name:     "Handle 500 error from downstream",
			want:     "",
			wantErr:  "unexpected response from downstream",
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("unexpected response from downstream")},
		},
		{
			name:     "fail to unmarshal body",
			mockAPIc: testUtil.MockAPIcer{Response: []byte{32}},
			wantErr:  "unexpected response from downstream",
		},
	}
	for _, test := range tests {
		echidna := &client{
			apicClient: test.mockAPIc,
		}

		t.Run(test.name, func(t *testing.T) {
			got, err := echidna.GetWrappingKey(testutil.GetContext(true))
			if test.wantErr != "" {
				assert.Empty(t, got)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

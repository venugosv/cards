//nolint:deadcode
package echidna

import (
	"encoding/json"

	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"google.golang.org/grpc/codes"
)

type GetWrappingKeyRequest struct {
	// The operator (username) for the session that this call is being performed on behalf of, to be recorded in the audit logs. this should be the client identifier.
	OperatorID string     `json:"operatorId,omitempty"`
	RespFormat respFormat `json:"respFormat,omitempty"`
	LogLevel   LogLevel   `json:"LogLevel,omitempty"`
}

type Action string

const (
	ActionGetWrappingKey Action = "getWrappingKey"
	ActionSelect         Action = "selectPIN"
	ActionVerify         Action = "verifyPIN"
	ActionChange         Action = "changePIN"
)

type LogLevel string

const (
	LoglevelOff  LogLevel = "OFF"
	LoglevelInfo LogLevel = "INFO"
	LoglevelFine LogLevel = "FINE"
)

func (l LogLevel) pointer() *LogLevel {
	return &l
}

type respFormat string

//nolint:varcheck
const (
	formatSoap  respFormat = "SOAP"
	formatXml   respFormat = "XML"
	formatProps respFormat = "PROPS"
	formatJson  respFormat = "JSON"
)

func (r respFormat) pointer() *respFormat {
	return &r
}

type IncomingRequest struct {
	PlainPAN          string
	EncryptedPINBlock string
}

type Request struct {
	// (Optional) The RespFormat to use for the web service wrappingKeyResponse. Default to the same as the request RespFormat.
	RespFormat *respFormat `json:"respFormat,omitempty"`
	// (Optional) The level of trace logging to use for the call
	LogLevel *LogLevel `json:"LogLevel,omitempty"`
	// (Required) The operator (username) for the session that this call is being performed on behalf of, to be recorded in the audit logs. this should be the client identifier.
	OperatorID string `json:"operatorID,omitempty"`
	// (Required) The plain-text PAN value
	PlainPAN string `json:"plainPAN,omitempty"`
	// (Optional) The PAN-reference value used to generate the PIN block. If not given, assumed to be 0.
	RefPAN *string `json:"refPAN,omitempty"`
	// (Required) The encrypted PIN block with validation data that was obtained by the client application calling the getEncryptedIPB method of the Salt client SDK.
	EncryptedPINBlock string `json:"encryptedPINBlock,omitempty"`
	// Text Information
	TxtInfo TxtInfo `json:"txtInfo,omitempty"`
}

func newRequest(plainPan string, encryptedPINBlock string) []byte {
	req := Request{
		RespFormat:        formatJson.pointer(),
		LogLevel:          LoglevelInfo.pointer(),
		OperatorID:        "",
		PlainPAN:          plainPan,
		RefPAN:            nil, // TODO
		EncryptedPINBlock: encryptedPINBlock,
		TxtInfo:           TxtInfo{},
	}
	requestBody, _ := json.Marshal(req)
	return requestBody
}

type Response struct {
	// The name of the service method that was called.
	Method      Action      `json:"method,omitempty"`
	Result      Result      `json:"result,omitempty"`
	LogMessages LogMessages `json:"logmessages,omitempty"`
}

type GetWrappingKeyResponse struct {
	Response Response `json:"getWrappingKeyResponse"`
}

type SetPINResponse struct {
	Response Response `json:"selectPINResponse,omitempty"`
}

type VerifyPINResponse struct {
	Response Response `json:"verifyPINResponse,omitempty"`
}

type ChangePINResponse struct {
	Response Response `json:"changePINResponse,omitempty"`
}

func GetGRPCError(code int) codes.Code {
	rules := map[int]codes.Code{
		0:    codes.OK,                // PIN operation successful.
		55:   codes.InvalidArgument,   // Incorrect PIN.
		75:   codes.ResourceExhausted, // Maximum PIN tries exceeded.
		1010: codes.Internal,          // Operation failed due to Tandem error response.
		1011: codes.Internal,          // Operation failed due to request-response error.
		1012: codes.Unavailable,       // RemotePIN service unavailable.
		1013: codes.DeadlineExceeded,  // Operation has timed out.
		1014: codes.Unavailable,       // Public key information is unavailable within the RemotePIN service.
		1015: codes.Unavailable,       // Operation failed due to RemotePIN service error.
	}

	out, ok := rules[code]
	if !ok {
		return codes.Unknown
	}

	return out
}

func GetANZError(code int) anzcodes.Code {
	rules := map[int]anzcodes.Code{
		55:   anzcodes.ValidationFailure,  // Incorrect PIN.
		75:   anzcodes.RateLimitExhausted, // Maximum PIN tries exceeded.
		1010: anzcodes.DownstreamFailure,  // Operation failed due to Tandem error response.
		1011: anzcodes.DownstreamFailure,  // Operation failed due to request-response error.
		1012: anzcodes.DownstreamFailure,  // RemotePIN service unavailable.
		1013: anzcodes.DownstreamFailure,  // Operation has timed out.
		1014: anzcodes.DownstreamFailure,  // Public key information is unavailable within the RemotePIN service.
		1015: anzcodes.DownstreamFailure,  // Operation failed due to RemotePIN service error.
	}

	out, ok := rules[code]
	if !ok {
		return anzcodes.Unknown
	}

	return out
}

func GetErrorMsg(code int) string {
	rules := map[int]string{
		55:   "Incorrect PIN",
		75:   "Maximum PIN tries exceeded",
		1010: "Operation failed due to internal error",
		1011: "Operation failed due to internal error",
		1012: "Service unavailable",
		1013: "Operation has timed out",
		1014: "Information is unavailable within the PIN service.",
		1015: "Operation failed due to service error",
	}

	out, ok := rules[code]
	if !ok {
		return "Unknown"
	}

	return out
}

type Result struct {
	// The result or error code - 0 indicates success
	Code int `json:"code,omitempty"`
	// A result message providing additional detail for the result code.
	Message string `json:"message,omitempty"`
	// Base64 encoded public key information in a respFormat that can be interpreted by the Salt client side SDK.
	EncodedKey *string `json:"encodedKey,omitempty"`
}

type LogMessages struct {
	// The level below which log records were not added to the list
	WantLevel LogLevel `json:"wantLevel,omitempty"`
	// Log messages generated by the service call.
	Item []string `json:"item,omitempty"`
}

type TxtInfo struct {
	// The transaction ID for Tandem side audit purposes. Allows 1-12 chars, otherwise error 1011.
	ID *string `json:"id,omitempty"`
	// The user details, who consumes the PIN operation. This is for log/audit purposes. Allows 8 chars, otherwise error 1011.
	User *string `json:"usr,omitempty"`
	// The branch details that the user belongs to. This is for log/audit purposes. Allows 6 chars, otherwise error 1011.
	BSB *string `json:"bsb,omitempty"`
	// The workstation that the operation is being initiated from. This is for log/audit purposes. Allows 16 chars, otherwise error 1011.
	WorkStation *string `json:"wsn,omitempty"`
}

type IncomingChangePINRequest struct {
	PlainPAN             string
	EncryptedPINBlockOld string
	EncryptedPINBlockNew string
}

type changePINRequest struct {
	// The RespFormat to use for the web service wrappingKeyResponse. Default to the same as the request RespFormat.
	RespFormat *respFormat `json:"respFormat,omitempty"`
	// The level of trace logging to use for the call
	LogLevel *LogLevel `json:"LogLevel,omitempty"`
	// The operator (username) for the session that this call is being performed on behalf of, to be recorded in the audit logs. this should be the client identifier.
	OperatorID string `json:"operatorID,omitempty"`
	// The plain-text PAN value
	PlainPAN string `json:"plainPAN,omitempty"`
	// The PAN-reference value used to generate the PIN block. If not given, assumed to be 0.
	RefPANOld *string `json:"refPANOld,omitempty"`
	// The PAN-reference value used to generate the PIN block. If not given, assumed to be 0.
	RefPANNew *string `json:"refPANNew,omitempty"`
	// The encrypted PIN block with validation data that was obtained by the client application calling the getEncryptedIPB method of the Salt client SDK.
	EncryptedPINBlockOld string `json:"encryptedOldPINBlock,omitempty"`
	// The encrypted PIN block with validation data that was obtained by the client application calling the getEncryptedIPB method of the Salt client SDK.
	EncryptedPINBlockNew string `json:"encryptedNewPINBlock,omitempty"`
	// Text Information
	TxtInfo TxtInfo `json:"txtInfo,omitempty"`
}

func newChangePINRequest(plainPan string, encryptedPINBlockOld string, encryptedPINBlockNew string) []byte {
	request := changePINRequest{
		RespFormat:           formatJson.pointer(),
		LogLevel:             LoglevelInfo.pointer(),
		OperatorID:           "",
		PlainPAN:             plainPan,
		RefPANOld:            nil,
		RefPANNew:            nil,
		EncryptedPINBlockOld: encryptedPINBlockOld,
		EncryptedPINBlockNew: encryptedPINBlockNew,
		TxtInfo:              TxtInfo{},
	}
	requestBody, _ := json.Marshal(request)
	return requestBody
}

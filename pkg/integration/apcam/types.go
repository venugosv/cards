package apcam

type Request struct {
	TraceInfo TraceInfo `json:"traceInfo"`
	CardInfo  CardInfo  `json:"cardInfo"`
	Apple     Apple     `json:"apple"`
	// user id to be sent to OTP manager as part of Generate OTP API for EFTPOS scheme
	// example: null
	UUID string `json:"uuid"`
}

type TraceInfo struct {
	// Unique identifier for the request. This is the request id provided by the API consumer
	// example: 2790503e-5208-11ea-9d06-cd011c010000
	MessageID string `json:"messageID"`
	// Identifier used for correlation for messages between card schemes and CAM
	// example: 2790503e-5208-11ea-9d06-cd011c010000
	ConversationID string `json:"conversationID"`
}

type CardInfo struct {
	// Card number passed from the scheme. It's mandatory if cardData is not provided. Mandatory for in-App
	// Provisioning API example: 4645790062743033
	Fpan string `json:"fpan"`
	// Expiry date of the card in the YYYY-MM format. Not required for in-App Provisioning API.
	// example: 2025-12
	ExpiryDate string `json:"expiryDate"`
}

type Apple struct {
	// A one-time use base64-encoded nonce generated by Apple
	// example: ZWZhZHNmZjIzNDEyMzQxMjQzMTIzcg==
	Nonce string `json:"nonce"`
	// The base64-encoded device and account specific signature of the nonce
	// example: WldaaFpITm1aakl6TkRFeU16UXhNalF6TVRJemNnPT0=
	NonceSignature string `json:"nonceSignature"`
	// An array of base64-DER-encoded certificates conforming to the specifications described in Apple’s
	// Issuer Application-Based Provisioning document. The first element is the leaf certificate. The
	// second element is the sub CA certificate
	// example: MIICYDCCAgigAwIBAgIBATAJBgcqhkjOPQQBME0xCzAJBgNVBAYTAkFVMQwwCgYDVQQIDANWSUMxDDAKBgNVBAcMA01FTDEV...
	Certificates []string `json:"certificates"`
}

type Response struct {
	TraceInfo TraceInfo         `json:"traceInfo"`
	Apple     AppleResponseData `json:"apple"`
}

type AppleResponseData struct {
	EncryptedPassData  string `json:"encryptedPassData"`
	ActivationData     string `json:"activationData"`
	EphemeralPublicKey string `json:"ephemeralPublicKey"`
}

type ErrorInfo struct {
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
}

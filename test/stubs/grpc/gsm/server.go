package gsm

import (
	"context"

	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StubServer struct {
	smpb.SecretManagerServiceServer
	secrets map[string][]byte
}

// NewStubServer creates a SecretManagerServiceServer stub
func NewStubServer() smpb.SecretManagerServiceServer {
	return &StubServer{
		secrets: map[string][]byte{
			"testSecretId":                     []byte(`redispassword`),
			"apic-corp-client-id-np":           []byte(`password`),
			"apic-ecom-client-id-np":           []byte(`password`),
			"cards-forgerock-secret-np":        []byte(`password`),
			"cardcontrols-forgerock-secret-np": []byte(`password`),
			"callback-forgerock-secret-np":     []byte(`password`),
			"wallet-visa-api-key-np":           []byte(`ertyukl`),
			"wallet-visa-shared-secret-np":     []byte(`LyQnklSrxsk3Ch2+AHi9HoDW@//x1LwM123QP/ln`),
		},
	}
}

func (ss *StubServer) AccessSecretVersion(_ context.Context, req *smpb.AccessSecretVersionRequest) (*smpb.AccessSecretVersionResponse, error) {
	secret, ok := ss.secrets[req.GetName()]
	if !ok {
		return nil, status.Error(codes.NotFound, "Secret not found")
	}
	return &smpb.AccessSecretVersionResponse{
		Name: req.GetName(),
		Payload: &smpb.SecretPayload{
			Data: secret,
		},
	}, nil
}

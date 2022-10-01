package ratelimit

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/anzx/pkg/gsm"
)

type secretManager interface {
	AccessSecretVersion(ctx context.Context, in *smpb.AccessSecretVersionRequest, _ ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error)
}

type mockSecretManagerServiceClient struct {
	accessSecretVersionFunc func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error)
}

func (m mockSecretManagerServiceClient) AccessSecretVersion(ctx context.Context, req *smpb.AccessSecretVersionRequest, _ ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
	return m.accessSecretVersionFunc(ctx, req)
}

func TestRedisConfig_GetSecrets(t *testing.T) {
	privateKey := generatePrivateKey(t)
	publicBytes := encodePublicCertToPEM(t, privateKey)

	tests := []struct {
		name    string
		config  *RedisConfig
		sm      secretManager
		want    *RedisConfig
		wantErr string
	}{
		{
			name: "happy path",
			config: &RedisConfig{
				SecretID:  "SECRETID",
				TLSCertID: "TLSCertID",
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					name := req.GetName()
					var data []byte
					if name == "SECRETID" {
						data = []byte(`password`)
					} else {
						data = publicBytes
					}
					return &smpb.AccessSecretVersionResponse{
						Name: name,
						Payload: &smpb.SecretPayload{
							Data: data,
						},
					}, nil
				},
			},
			want: &RedisConfig{
				Password: "password",
			},
		}, {
			name: "unhappy path",
			config: &RedisConfig{
				SecretID:  "SECRETID",
				TLSCertID: "TLSCertID",
			},
			sm: mockSecretManagerServiceClient{
				accessSecretVersionFunc: func(ctx context.Context, req *smpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*smpb.AccessSecretVersionResponse, error) {
					return nil, errors.New("oh no")
				},
			},
			wantErr: "failed to access secret SECRETID: oh no",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			secrets := &gsm.Client{SM: test.sm}

			err := test.config.GetSecrets(ctx, secrets)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.Password, test.config.Password)
				require.NotNil(t, test.config.TlsConfig.RootCAs)
			}
		})
	}
}

// encodePublicCertToPEM take a rsa.PublicKey and return bytes suitable for writing to .pub file.
func encodePublicCertToPEM(t *testing.T, privateKey *rsa.PrivateKey) []byte {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Log("failed to create x509 cert")
		t.Fail()
	}

	// pem.Block
	caBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}

	// Private key in PEM format
	return pem.EncodeToMemory(&caBlock)
}

// generatePrivateKey creates a RSA Private Key of specified byte size.
func generatePrivateKey(t *testing.T) *rsa.PrivateKey {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Log("failed to create private key")
		t.Fail()
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		t.Log("failed to validate private key")
		t.Fail()
	}

	t.Log("Private Key generated")

	return privateKey
}

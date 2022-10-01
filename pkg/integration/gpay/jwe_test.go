package gpay

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/square/go-jose.v2"
)

const (
	keyID        = "ertyukl"
	sharedSecret = "LyQnklSrxsk3Ch2+AHi9HoDW@//x1LwM123QP/ln" //nolint:gosec
	payload      = "payload"
	key          = "test_client_secret_key"
	secret       = "test_client_secret"
	bad_key      = "fake_key"
)

type mockSecretManager struct{}

func (m mockSecretManager) AccessSecretVersion(_ context.Context, req *secretmanagerpb.AccessSecretVersionRequest, _ ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if req.Name == key {
		return &secretmanagerpb.AccessSecretVersionResponse{
			Name:    key,
			Payload: &secretmanagerpb.SecretPayload{Data: []byte(secret)},
		}, nil
	}
	return nil, fmt.Errorf("oh no")
}

func TestNewClient(t *testing.T) {
	gsm := &gsm.Client{SM: &mockSecretManager{}}
	t.Run("success", func(t *testing.T) {
		config := &Config{
			APIKeyKey:       key,
			SharedSecretKey: key,
		}
		c, err := NewClientFromConfig(context.Background(), config, gsm)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})
	t.Run("null client secret", func(t *testing.T) {
		config := &Config{
			APIKeyKey: bad_key,
		}
		c, err := NewClientFromConfig(context.Background(), config, gsm)
		require.Error(t, err)
		assert.Nil(t, c)
		assert.EqualError(t, err, "fabric error: status_code=InvalidArgument, error_code=1, message=failed to create GPay client, reason=failed to get keyID with key fake_key")
	})
	t.Run("null client secret", func(t *testing.T) {
		config := &Config{
			APIKeyKey:       key,
			SharedSecretKey: bad_key,
		}
		c, err := NewClientFromConfig(context.Background(), config, gsm)
		require.Error(t, err)
		assert.Nil(t, c)
		assert.EqualError(t, err, "fabric error: status_code=InvalidArgument, error_code=1, message=failed to create GPay client, reason=failed to get sharedSecret with key fake_key")
	})
	t.Run("nil config", func(t *testing.T) {
		c, err := NewClientFromConfig(context.Background(), nil, nil)
		assert.Nil(t, c)
		assert.Nil(t, err)
	})
}

func TestClient_CreateJWE(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		g := &client{
			recipient: jose.Recipient{
				Algorithm: jose.A256GCMKW,
				Key:       sha256Hash([]byte(sharedSecret)),
				KeyID:     keyID,
			},
		}

		c := testClient{client: g}

		ctx := context.Background()
		jwe, err := g.CreateJWE(ctx, payload)
		require.NoError(t, err)
		t.Log(jwe)

		parts := strings.Split(string(jwe), ".")
		if len(parts) != 5 {
			panic("Invalid JWE. JWE should contain 5 parts")
		}

		// JWE Header is the first part of the JWE
		jweHeader, err := base64.StdEncoding.DecodeString(parts[0])
		require.NoError(t, err)
		t.Logf("JWE Header: %s", string(jweHeader))

		decryptedPayload, err := c.decryptJWE(string(jwe))
		require.NoError(t, err)
		t.Log(decryptedPayload)

		jws, err := c.createJWS(jwe)
		require.NoError(t, err)
		t.Log(jws)

		parts = strings.Split(jws, ".")
		if len(parts) != 3 {
			panic("Invalid JWS, JWS should contain 3 parts")
		}

		// JWE Header is the first part of the JWE
		jwsHeader, _ := base64.StdEncoding.DecodeString(parts[0])
		t.Log("JWS Header: " + string(jwsHeader))

		jweFromJws, err := c.verifyJWS(jws)
		require.NoError(t, err)
		t.Log(jweFromJws)

		decryptedJWE, err := c.decryptJWE(string(jwe))
		require.NoError(t, err)
		t.Log(decryptedJWE)

		assert.Equal(t, payload, decryptedJWE)
	})
	t.Run("unsupported key type/format", func(t *testing.T) {
		g := &client{
			recipient: jose.Recipient{
				Algorithm: jose.A256GCMKW,
			},
		}

		ctx := context.Background()
		_, err := g.CreateJWE(ctx, payload)
		require.Error(t, err)
		assert.EqualError(t, err, "fabric error: status_code=Internal, error_code=4, message=failed to Create GPay JWE, reason=unable to create payload encryptor")
	})
}

type testClient struct {
	*client
}

// decryptJWE Using API Key and Shared Secret (Symmetric Encryption)
func (g *testClient) decryptJWE(encryptedPayload string) (string, error) {
	// Parse the serialized, encrypted JWE object. An error would indicate that
	// the given input did not represent a valid message.
	object, err := jose.ParseEncrypted(encryptedPayload)
	if err != nil {
		return "", err
	}

	// Now we can decrypt and get back our original payload. An error here
	// would indicate the the message failed to decrypt, e.g. because the auth
	// tag was broken or the message was tampered with.
	decrypted, err := object.Decrypt(g.recipient.Key)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// createJWS Sign a JWE and create the JWS
func (g *testClient) createJWS(jwe []byte) (string, error) {
	opts := jose.SignerOptions{}
	opts.WithHeader("kid", g.recipient.KeyID)

	// create Square.jose signing key
	key := jose.SigningKey{
		Algorithm: jose.HS256,
		Key:       g.recipient.Key,
	}

	signer, err := jose.NewSigner(key, &opts)
	if err != nil {
		return "", err
	}

	object, err := signer.Sign(jwe)
	if err != nil {
		return "", err
	}

	// Serialize the encrypted object using the compact serialization format.
	serialized, err := object.CompactSerialize()
	if err != nil {
		return "", err
	}

	return serialized, nil
}

// verifyJWS and return the JWE
func (g *testClient) verifyJWS(jws string) (string, error) {
	jsonWebSig, err := jose.ParseSigned(jws)
	if err != nil {
		return "", err
	}

	jwe, err := jsonWebSig.Verify(g.recipient.Key)
	if err != nil {
		return "", err
	}

	return string(jwe), nil
}

package gpay

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"gopkg.in/square/go-jose.v2"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/pkg/gsm"
	"github.com/anzx/pkg/log"
)

type Client interface {
	CreateJWE(context.Context, string) ([]byte, error)
}

type client struct {
	recipient jose.Recipient
}

type Config struct {
	APIKeyKey       string `json:"apiKeyKey,omitempty"`
	SharedSecretKey string `json:"sharedSecretKey,omitempty"`
}

func NewClientFromConfig(ctx context.Context, cfg *Config, gsmClient *gsm.Client) (Client, error) {
	if cfg == nil {
		logf.Debug(ctx, "GPay config not provided %v", cfg)
		return nil, nil
	}
	return NewClient(ctx, cfg.APIKeyKey, cfg.SharedSecretKey, gsmClient)
}

func NewClient(ctx context.Context, apiKeyKey string, sharedSecretKey string, gsmClient *gsm.Client) (Client, error) {
	apiKey, err := gsmClient.AccessSecret(ctx, apiKeyKey)
	if err != nil {
		log.Error(ctx, err, "failed to get keyID", log.Str("key", apiKeyKey))
		return nil, anzerrors.Wrap(err, codes.InvalidArgument, "failed to create GPay client",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, fmt.Sprintf("failed to get keyID with key %s", apiKeyKey)))
	}

	sharedSecret, err := gsmClient.AccessSecretBytes(ctx, sharedSecretKey)
	if err != nil {
		log.Error(ctx, err, "failed to get sharedSecret", log.Str("key", sharedSecretKey))
		return nil, anzerrors.Wrap(err, codes.InvalidArgument, "failed to create GPay client",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, fmt.Sprintf("failed to get sharedSecret with key %s", sharedSecretKey)))
	}

	return &client{
		recipient: jose.Recipient{
			Algorithm: jose.A256GCMKW,
			Key:       sha256Hash(sharedSecret),
			KeyID:     apiKey,
		},
	}, nil
}

func (g *client) CreateJWE(ctx context.Context, payload string) ([]byte, error) {
	opts := new(jose.EncrypterOptions)
	opts.WithHeader("kid", g.recipient.KeyID)

	// Time when JWE was issued. Expressed in UNIX epoch time (seconds since 1
	// January 1970) and issued at timestamp in UTC when the transaction was
	// created and signed.
	thirtySecondsPrior := time.Now().Add(time.Duration(-30) * time.Second)
	opts.WithHeader("iat", thirtySecondsPrior.UTC().Unix())

	encryptor, err := jose.NewEncrypter(jose.A256GCM, g.recipient, opts)
	if err != nil {
		logf.Error(ctx, err, "unable to create payload encryptor")
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to Create GPay JWE",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to create payload encryptor"))
	}

	// Encrypt a sample payload. Calling the encryptor returns an encrypted
	// JWE object, which can then be serialized for output afterwards. An error
	// would indicate a problem in an underlying cryptographic primitive.
	object, err := encryptor.Encrypt([]byte(payload))
	if err != nil {
		logf.Error(ctx, err, "unable to encrypt payload")
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to Create GPay JWE",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to encrypt payload"))
	}

	// Serialize the encrypted object using the compact serialization format.
	serialize, err := object.CompactSerialize()
	if err != nil {
		logf.Error(ctx, err, "unable to serialize encryptor object")
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to Create GPay JWE",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to serialize encryptor object"))
	}

	return []byte(serialize), nil
}

// Hash a plain text using SHA256
func sha256Hash(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	return h.Sum(nil)
}

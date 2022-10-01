package vault_external

import (
	"context"

	"github.com/googleapis/gax-go/v2"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

// JwtSigner is an interface for things that can sign JWTs. It exists to be compatible with the Google credentials client,
//  so that we can do tests without hitting Google for JWT signing.
type JwtSigner interface {
	SignJwt(ctx context.Context, req *credentialspb.SignJwtRequest, opts ...gax.CallOption) (*credentialspb.SignJwtResponse, error)
}

// FixedSignedJwt always returns a preset JWT, or an error. It is for use in testing and stubbed environments.
type FixedSignedJwt struct {
	jwt string
	key string
}

// SignJwt returns static data for testing and stubbed environments
func (f *FixedSignedJwt) SignJwt(_ context.Context, _ *credentialspb.SignJwtRequest, _ ...gax.CallOption) (*credentialspb.SignJwtResponse, error) {
	return &credentialspb.SignJwtResponse{
		SignedJwt: f.jwt,
		KeyId:     f.key,
	}, nil
}

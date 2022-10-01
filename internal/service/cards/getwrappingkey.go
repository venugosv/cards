package cards

import (
	"context"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

const failedWrappingKey = "failed to get wrapping key"

func (s server) GetWrappingKey(ctx context.Context, _ *cpb.GetWrappingKeyRequest) (*cpb.GetWrappingKeyResponse, error) {
	key, err := s.Echidna.GetWrappingKey(ctx)
	if err != nil {
		return nil, serviceErr(err, failedWrappingKey)
	}

	return &cpb.GetWrappingKeyResponse{
		EncodedKey: key,
	}, nil
}

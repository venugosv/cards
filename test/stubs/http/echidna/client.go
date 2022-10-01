package echidna

import (
	"context"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/echidna"
)

type StubClient struct {
	Err         error
	testingData *data.Data
}

// NewStubClient creates a EchidnaClient client stubs
func NewStubClient(testData *data.Data) StubClient {
	return StubClient{
		testingData: testData,
	}
}

func (e StubClient) GetWrappingKey(ctx context.Context) (string, error) {
	if e.Err != nil {
		return "", e.Err
	}
	return encodedKey, nil
}

func (e StubClient) SelectPIN(_ context.Context, r echidna.IncomingRequest) error {
	if e.Err != nil {
		return e.Err
	}

	return nil
}

func (e StubClient) VerifyPIN(ctx context.Context, _ echidna.IncomingRequest) error {
	if e.Err != nil {
		return e.Err
	}

	return nil
}

func (e StubClient) ChangePIN(ctx context.Context, _ echidna.IncomingChangePINRequest) error {
	if e.Err != nil {
		return e.Err
	}

	return nil
}

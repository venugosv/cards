package gpay

import (
	"context"
)

type StubClient struct {
	Err error
}

// NewStubClient creates a gpayClient client stubs
func NewStubClient() StubClient {
	return StubClient{}
}

func (e StubClient) CreateJWE(context.Context, string) ([]byte, error) {
	if e.Err != nil {
		return nil, e.Err
	}
	return []byte(`OPC`), nil
}

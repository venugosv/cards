package ratelimit

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/ratelimit"
)

type StubClient struct {
	Err error
}

func (r StubClient) Allow(_ context.Context, _ ratelimit.Domain) error {
	return r.Err
}

func NewStubClient() StubClient {
	return StubClient{}
}

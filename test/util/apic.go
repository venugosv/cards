package util

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/util/apic"
)

type MockAPIcer struct {
	DoCall      func(ctx context.Context, request *apic.Request, operation string) ([]byte, error)
	Response    []byte
	ResponseErr error
}

func (m MockAPIcer) Do(ctx context.Context, request *apic.Request, operation string) ([]byte, error) {
	if m.DoCall == nil {
		return m.Response, m.ResponseErr
	}
	return m.DoCall(ctx, request, operation)
}

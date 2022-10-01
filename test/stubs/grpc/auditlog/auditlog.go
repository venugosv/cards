package auditlog

import (
	"context"

	"github.com/pkg/errors"

	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"google.golang.org/protobuf/encoding/protojson"
)

type StubClient struct {
	Err  error
	Hook AuditLogHook
}

type AuditLogHook func(buf []byte)

func NewStubClient() StubClient {
	return StubClient{}
}

func (s StubClient) Send(_ context.Context, buf []byte) error {
	if s.Err != nil {
		return s.Err
	}

	if s.Hook != nil {
		s.Hook(buf)
	}
	p := &audit.AuditLog{}
	if err := protojson.Unmarshal(buf, p); err != nil {
		return err
	}

	if err := p.Validate(); err != nil {
		return err
	}

	callSuccessful := p.Status.GetCode() == "0"
	noResponseBody := p.Response.GetValue() == nil
	noErrorDetail := p.Status.Details == nil

	if callSuccessful && noResponseBody && noErrorDetail {
		return errors.New("no error sent to audit log")
	}

	if callSuccessful && noResponseBody && !noErrorDetail {
		return errors.New("no response data sent to audit log")
	}

	return nil
}

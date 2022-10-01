package fakerock

import (
	"context"

	frpb "github.com/anzx/fabricapis/pkg/fabric/service/fakerock/v1alpha1"

	"google.golang.org/genproto/googleapis/api/httpbody"

	"google.golang.org/grpc"
)

type StubClient struct {
	*StubServer
	Err error
}

func NewStubClient() StubClient {
	return StubClient{
		StubServer: NewStubServer(),
	}
}

func (s StubClient) JWKS(_ context.Context, _ *frpb.JWKSRequest, _ ...grpc.CallOption) (*httpbody.HttpBody, error) {
	panic("not implemented")
}

func (s StubClient) Login(_ context.Context, _ *frpb.LoginRequest, _ ...grpc.CallOption) (*frpb.LoginResponse, error) {
	panic("not implemented")
}

func (s StubClient) SystemLogin(ctx context.Context, in *frpb.SystemLoginRequest, _ ...grpc.CallOption) (*frpb.SystemLoginResponse, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return s.StubServer.SystemLogin(ctx, in)
}

package fakerock

import (
	"context"

	frpb "github.com/anzx/fabricapis/pkg/fabric/service/fakerock/v1alpha1"
)

type StubServer struct {
	Err error
	frpb.UnimplementedFakerockAPIServer
}

func NewStubServer() *StubServer {
	return &StubServer{}
}

func (s StubServer) SystemLogin(context.Context, *frpb.SystemLoginRequest) (*frpb.SystemLoginResponse, error) {
	return &frpb.SystemLoginResponse{Token: "token"}, nil
}

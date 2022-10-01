package dcvv2

import (
	"context"

	"github.com/anzx/fabric-cards/test/data"

	"google.golang.org/grpc"

	dcvv2pb "github.com/anzx/fabricapis/pkg/gateway/visa/service/dcvv2"
)

type StubClient struct {
	GenerateErr    error
	DCVV2APIServer StubServer
}

// NewStubClient creates a CustomerRulesAPIClient stubs
func NewStubClient(data *data.Data) StubClient {
	return StubClient{
		DCVV2APIServer: NewStubServer(data),
	}
}

func (s StubClient) Generate(ctx context.Context, in *dcvv2pb.Request, opts ...grpc.CallOption) (*dcvv2pb.Response, error) {
	if s.GenerateErr != nil {
		return nil, s.GenerateErr
	}
	return s.DCVV2APIServer.Generate(ctx, in)
}

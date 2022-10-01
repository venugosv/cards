package cardcontrols

import (
	"context"

	v1beta2pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"

	"google.golang.org/grpc"
)

type StubClient struct {
	TransferControlsErr error
}

func NewStubClient() StubClient {
	return StubClient{}
}

func (s StubClient) ListControls(_ context.Context, _ *v1beta2pb.ListControlsRequest, _ ...grpc.CallOption) (*v1beta2pb.ListControlsResponse, error) {
	return nil, nil
}

func (s StubClient) QueryControls(_ context.Context, _ *v1beta2pb.QueryControlsRequest, _ ...grpc.CallOption) (*v1beta2pb.CardControlResponse, error) {
	return nil, nil
}

func (s StubClient) SetControls(_ context.Context, _ *v1beta2pb.SetControlsRequest, _ ...grpc.CallOption) (*v1beta2pb.CardControlResponse, error) {
	return nil, nil
}

func (s StubClient) RemoveControls(_ context.Context, _ *v1beta2pb.RemoveControlsRequest, _ ...grpc.CallOption) (*v1beta2pb.CardControlResponse, error) {
	return nil, nil
}

func (s StubClient) TransferControls(_ context.Context, _ *v1beta2pb.TransferControlsRequest, _ ...grpc.CallOption) (*v1beta2pb.TransferControlsResponse, error) {
	return nil, s.TransferControlsErr
}

func (s StubClient) BlockCard(_ context.Context, _ *v1beta2pb.BlockCardRequest, _ ...grpc.CallOption) (*v1beta2pb.BlockCardResponse, error) {
	return nil, nil
}

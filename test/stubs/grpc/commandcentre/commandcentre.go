package commandcentre

import (
	"context"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
)

type StubClient struct {
	sdk.Publisher
}

func NewStubClient() *StubClient {
	return &StubClient{}
}

func (c StubClient) Publish(_ context.Context, _ sdk.PublishRequester) (*sdk.PublishResponse, error) {
	return &sdk.PublishResponse{
		Status: sdk.PublishResponsePublished,
	}, nil
}

type FakePublisher struct {
	Count    int
	Messages []string
}

func NewFakePublisher() FakePublisher {
	v := 0
	return FakePublisher{
		Count:    v,
		Messages: []string{},
	}
}

func (f *FakePublisher) Publish(_ context.Context, req sdk.PublishRequester) (*sdk.PublishResponse, error) {
	f.Count += 1
	in := req.(*sdk.NotificationForPersona)
	f.Messages = append(f.Messages, in.Preview.Body)
	return &sdk.PublishResponse{
		Status: sdk.PublishResponsePublished,
	}, nil
}

func (f FakePublisher) PublishSync(_ context.Context, _ *sdk.PublishSyncRequest) (*sdk.PublishSyncResponse, error) {
	panic("should not be called")
}

func (f FakePublisher) GetLastMessage() string {
	l := len(f.Messages)
	if l == 0 {
		return ""
	}
	return f.Messages[l-1]
}

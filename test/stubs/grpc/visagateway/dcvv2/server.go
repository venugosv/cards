package dcvv2

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/brianvoe/gofakeit/v6"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"

	"github.com/anzx/fabric-cards/test/data"

	dcvv2pb "github.com/anzx/fabricapis/pkg/gateway/visa/service/dcvv2"
)

type StubServer struct {
	data     *data.Data
	Controls []ccpb.ControlType
}

func (s StubServer) Generate(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
	count, err := strconv.Atoi(request.GetItemCount())
	if err != nil {
		return nil, err
	}
	list := make([]*dcvv2pb.Dcvv2ItemList, 0, count)
	for i := 0; i < count; i++ {
		item := &dcvv2pb.Dcvv2ItemList{
			Dcvv2Expiry: fmt.Sprintf("%d", time.Now().Add(time.Duration(i+1)*time.Hour).Unix()),
			Dcvv2Value:  fmt.Sprintf("%d", gofakeit.Number(100, 999)),
		}
		list = append(list, item)
	}
	return &dcvv2pb.Response{
		AccountInfo:           request.GetAccountInfo(),
		MessageIdentification: &dcvv2pb.MessageIdentification{TransactionId: uuid.NewString()},
		Dcvv2ItemList:         list,
		TransactionResults:    &dcvv2pb.TransactionResults{ActionCode: "00"},
	}, nil
}

// NewStubServer creates a CardEntitlementsAPIClient stubs
func NewStubServer(data *data.Data) StubServer {
	return StubServer{
		data: data,
	}
}

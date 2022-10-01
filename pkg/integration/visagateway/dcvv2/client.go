package dcvv2

import (
	"context"
	"fmt"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	dcvv2pb "github.com/anzx/fabricapis/pkg/gateway/visa/service/dcvv2"
	anzerrors "github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/errors/errcodes"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Client struct {
	dcvv2pb.DCVV2APIClient
	ClientID string
}

func NewClient(clientID string, conn *grpc.ClientConn) *Client {
	return &Client{
		DCVV2APIClient: dcvv2pb.NewDCVV2APIClient(conn),
		ClientID:       clientID,
	}
}

func (c Client) Generate(ctx context.Context, expiryDate string, cardNumber string) (*dcvv2pb.Dcvv2ItemList, error) {
	list, err := c.GenerateList(ctx, expiryDate, cardNumber, 1)
	if err != nil {
		return nil, err
	}

	return list[0], nil
}

func (c Client) GenerateList(ctx context.Context, expiryDate string, cardNumber string, count int) ([]*dcvv2pb.Dcvv2ItemList, error) {
	req := c.createRequest(expiryDate, cardNumber, count)

	if err := req.Validate(); err != nil {
		logf.Error(ctx, err, "invalid dcvv2 generate input")
		return nil, anzerrors.Wrap(err, codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "embedded message failed validation"))
	}

	resp, err := c.DCVV2APIClient.Generate(ctx, req)
	if err != nil {
		logf.Error(ctx, err, "invalid dcvv2 response")
		anzErr, _ := anzerrors.FromStatusError(err)
		return nil, anzerrors.Wrap(anzErr, anzerrors.GetStatusCode(anzErr), anzerrors.GetMessage(anzErr),
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "invalid response from visa gateway"))
	}

	actionCode := resp.GetTransactionResults().GetActionCode()
	if actionCode != "00" {
		err := fmt.Errorf("invalid action code returned %s", actionCode)
		logf.Error(ctx, err, "unexpected response from visa gateway")
		return nil, anzerrors.New(codes.NotFound, "invalid dcvv2 response", anzerrors.NewErrorInfo(ctx, errcodes.DownstreamFailure, err.Error()))
	}

	return resp.GetDcvv2ItemList(), nil
}

func (c Client) createRequest(expiryDate string, cardNumber string, count int) *dcvv2pb.Request {
	return &dcvv2pb.Request{
		AccountInfo: &dcvv2pb.AccountInfo{
			PrimaryAccountNumber: &dcvv2pb.PrimaryAccountNumber{
				ExpirationDate: expiryDate,
				Pan:            cardNumber,
			},
		},
		CardAcceptor: &dcvv2pb.CardAcceptor{
			ClientId: c.ClientID,
		},
		ItemCount: fmt.Sprintf("%d", count),
	}
}

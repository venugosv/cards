package cardonfile

import (
	"context"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/google/uuid"

	cofpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/cardonfile"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	group  = "STANDARD"
	layout = "2006-01-02 15:04:05.000"
)

type Client struct {
	cofpb.CardOnFileAPIClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		CardOnFileAPIClient: cofpb.NewCardOnFileAPIClient(conn),
	}
}

func (c Client) GetCardAcceptorID(ctx context.Context, cardNumber string) ([]string, error) {
	if cardNumber == "" {
		return nil, anzerrors.New(codes.NotFound, "card number not provided",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, "invalid card number"))
	}

	panLists, err := c.Inquiry(ctx, cardNumber)
	if err != nil {
		logf.Error(ctx, err, "failed to get response from visa gateway")
		return nil, err
	}

	var merchants []*cofpb.Merchants
	for _, panList := range panLists {
		if panList.GetPanData().GetPan() == cardNumber {
			merchants = panList.GetPanData().GetMerchants()
			break
		}
	}

	var cardAcceptorIDs []string
	for _, merchant := range merchants {
		cardAcceptorIDs = append(cardAcceptorIDs, merchant.GetCardAcceptorId())
	}

	return cardAcceptorIDs, nil
}

func (c Client) Inquiry(ctx context.Context, cardNumbers ...string) ([]*cofpb.PANList, error) {
	req := c.createRequest(cardNumbers)

	if err := req.Validate(); err != nil {
		logf.Error(ctx, err, "invalid card-on-file generate input")
		return nil, anzerrors.Wrap(err, codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "embedded message failed validation"))
	}

	resp, err := c.CardOnFileAPIClient.Inquiry(ctx, req)
	if err != nil {
		logf.Error(ctx, err, "invalid card-on-file response")
		anzErr, _ := anzerrors.FromStatusError(err)
		return nil, anzerrors.Wrap(anzErr, anzerrors.GetStatusCode(anzErr), anzerrors.GetMessage(anzErr),
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "invalid response from visa gateway"))
	}

	panList := resp.GetData().GetPanList()
	if panList == nil {
		return nil, anzerrors.New(codes.NotFound, "failed to get data from visa response",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, "data was not returned from Visa"))
	}

	return panList, nil
}

func (c Client) createRequest(pans []string) *cofpb.Request {
	// Date and time at which request is sent (up to milliseconds in UTC). Format: yyyy-MM-dd HH:mm:ss.SSS
	t := time.Now().Format(layout)
	return &cofpb.Request{
		Header: &cofpb.Request_Header{
			RequestMessageId: uuid.NewString(),
			MessageDateTime:  t,
		},
		Data: &cofpb.Request_Data{
			PrimaryAccountNumbers: pans,
			Group:                 group,
		},
	}
}

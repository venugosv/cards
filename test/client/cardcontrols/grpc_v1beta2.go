package cardcontrols

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/test/client/common"

	cards "github.com/anzx/fabric-cards/test/client/cards/v1beta1"

	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	ccpbv1beta2 "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	"google.golang.org/grpc"
)

type GRPCV1beta2Client struct {
	cardTestClient               cards.TestClient
	cardControlsV1beta2APIClient ccpbv1beta2.CardControlsAPIClient
	ctx                          context.Context
	state                        common.ConnectionState
}

func NewGRPCV1beta2Client(ctx context.Context, cards cards.TestClient, cardControls *grpc.ClientConn) *GRPCV1beta2Client {
	return &GRPCV1beta2Client{
		ctx:                          ctx,
		cardTestClient:               cards,
		cardControlsV1beta2APIClient: ccpbv1beta2.NewCardControlsAPIClient(cardControls),
	}
}

func (c *GRPCV1beta2Client) LoadCard(t *testing.T) {
	c.cardTestClient.LoadCard(t)
	c.state.CurrentCard = c.cardTestClient.GetCurrentCard()
}

func (c *GRPCV1beta2Client) Can(eligibility epbv1beta1.Eligibility) bool {
	return c.cardTestClient.Can(eligibility)
}

func (c *GRPCV1beta2Client) ListControls() (*ccpbv1beta2.ListControlsResponse, error) {
	return c.cardControlsV1beta2APIClient.ListControls(c.ctx, &ccpbv1beta2.ListControlsRequest{})
}

func (c *GRPCV1beta2Client) QueryControls() (*ccpbv1beta2.CardControlResponse, error) {
	in := c.state.CurrentCard.GetTokenizedCardNumber()
	return c.cardControlsV1beta2APIClient.QueryControls(c.ctx, &ccpbv1beta2.QueryControlsRequest{
		TokenizedCardNumber: in,
	})
}

func (c *GRPCV1beta2Client) SetControls(controls ...ccpbv1beta2.ControlType) (*ccpbv1beta2.CardControlResponse, error) {
	var controlRequest []*ccpbv1beta2.ControlRequest
	for _, control := range controls {
		controlRequest = append(controlRequest, &ccpbv1beta2.ControlRequest{
			ControlType: control,
		})
	}
	return c.cardControlsV1beta2APIClient.SetControls(c.ctx, &ccpbv1beta2.SetControlsRequest{
		TokenizedCardNumber: c.state.CurrentCard.GetTokenizedCardNumber(),
		CardControls:        controlRequest,
	})
}

func (c *GRPCV1beta2Client) RemoveControls(controls ...ccpbv1beta2.ControlType) (*ccpbv1beta2.CardControlResponse, error) {
	return c.cardControlsV1beta2APIClient.RemoveControls(c.ctx, &ccpbv1beta2.RemoveControlsRequest{
		TokenizedCardNumber: c.state.CurrentCard.GetTokenizedCardNumber(),
		ControlTypes:        controls,
	})
}

func (c *GRPCV1beta2Client) TransferControls(newTokenizedCardNumber string) (*ccpbv1beta2.TransferControlsResponse, error) {
	return c.cardControlsV1beta2APIClient.TransferControls(c.ctx, &ccpbv1beta2.TransferControlsRequest{
		CurrentTokenizedCardNumber: c.state.CurrentCard.GetTokenizedCardNumber(),
		NewTokenizedCardNumber:     newTokenizedCardNumber,
	})
}

func (c *GRPCV1beta2Client) BlockCard(action ccpbv1beta2.BlockCardRequest_Action) (*ccpbv1beta2.BlockCardResponse, error) {
	return c.cardControlsV1beta2APIClient.BlockCard(c.ctx, &ccpbv1beta2.BlockCardRequest{
		TokenizedCardNumber: c.state.CurrentCard.GetTokenizedCardNumber(),
		Action:              action,
	})
}

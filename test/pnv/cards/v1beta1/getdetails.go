//go:build pnv
// +build pnv

package v1beta1

import (
	"context"
	"testing"
	"time"

	cards "github.com/anzx/fabric-cards/test/client/cards/v1beta1"
	"github.com/anzx/fabric-cards/test/config"
	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/anzx/fabric-cards/test/common"
	"github.com/anzx/fabric-pnv/locustworker"

	"google.golang.org/protobuf/proto"
)

type GetDetails struct{}

func (GetDetails) Weight() int  { return 0 }
func (GetDetails) Name() string { return "v1beta1.CardAPI/GetDetails" }
func (GetDetails) Run(r locustworker.Runner) (int, time.Duration, error) {
	cc, err := locustworker.Setup(context.Background(), r.Config.Insecure, r.Config.Addr)
	if err != nil {
		return 0, 0, err
	}
	defer cc.Close()

	t := &testing.T{}
	user := r.GetUser()

	auth := common.GetAuthHeaders(t, user, config.AuthConfig{}, common.CardsScope...)

	ctx := auth.Context(t, context.Background())
	ctx, cancel := context.WithTimeout(ctx, r.Config.ConnectTimeout)
	defer cancel()

	c := cards.NewGRPCClient(ctx, cc, nil)

	eligible := false
	for !eligible {
		c.LoadCard(t)
		eligible = c.Can(epbv1beta1.Eligibility_ELIGIBILITY_GET_DETAILS)
	}

	start := time.Now()
	got, err := c.GetDetails()
	elapsed := time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}

	return proto.Size(got), elapsed, nil
}

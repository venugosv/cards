//go:build pnv
// +build pnv

package v1beta1

import (
	"context"
	"testing"
	"time"

	cards "github.com/anzx/fabric-cards/test/client/cards/v1beta1"
	"github.com/anzx/fabric-cards/test/common"
	"github.com/anzx/fabric-cards/test/config"
	"github.com/anzx/fabric-pnv/locustworker"
	"github.com/brianvoe/gofakeit/v6"

	"google.golang.org/protobuf/proto"
)

type ResetPIN struct{}

func (a ResetPIN) Weight() int  { return 0 }
func (a ResetPIN) Name() string { return "v1beta1.CardAPI/ResetPIN" }
func (a ResetPIN) Run(r locustworker.Runner) (int, time.Duration, error) {
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

	currentCard := user.Cards[gofakeit.Number(0, 9)]

	c.SetTokenizedCardNumber(currentCard.Token)

	start := time.Now()
	got, err := c.ResetPIN()
	elapsed := time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}

	return proto.Size(got), elapsed, nil
}

//go:build pnv
// +build pnv

package main

import (
	"log"
	"math/rand"
	"testing"

	"github.com/anzx/fabric-cards/test/config"
	"github.com/anzx/fabric-cards/test/pnv/cards/v1beta1"

	"github.com/anzx/anzdata"
	"github.com/anzx/fabric-pnv/locustworker"
)

func main() {
	cfg, err := locustworker.Env()
	if err != nil {
		log.Fatal(err)
	}

	t := &testing.T{}
	appCfg, err := config.Load(t)
	if err != nil {
		t.Fatal("pnv main failed to load config", err)
	}
	t.Logf("config loaded: \n%v", cfg.String())

	cards := appCfg.Cards
	cfg.Addr = cards.BaseURL
	cfg.Insecure = cards.Insecure
	cfg.ConnectTimeout = appCfg.Timeout
	cfg.MaxUser = appCfg.MaxUser

	auth := appCfg.Cards.Auth

	v1beta1 := []locustworker.Tester{
		v1beta1.List{AuthConfig: auth, Headers: cards.Headers},
		v1beta1.ActivateGetDetails{AuthConfig: auth, Headers: cards.Headers},
		v1beta1.GetWrappingKey{AuthConfig: auth, Headers: cards.Headers},
		v1beta1.Lost{AuthConfig: auth, Headers: cards.Headers},
		v1beta1.Damaged{AuthConfig: auth, Headers: cards.Headers},
		v1beta1.SetChangePIN{AuthConfig: auth, Headers: cards.Headers},
	}

	locustworker.Run(cfg, v1beta1,
		locustworker.WithUserFunction(
			func() anzdata.User { return anzdata.MustRandomUserFromInt(rand.Intn(cfg.MaxUser)) },
		),
	)
}

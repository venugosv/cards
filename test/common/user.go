package common

import (
	"context"
	_ "embed"
	"testing"
	"time"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/anzx/fabric-cards/test/client/cards/v1beta1"
	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/test/config"
	"github.com/brianvoe/gofakeit/v6"

	"gopkg.in/yaml.v3"

	"github.com/anzx/anzdata"
)

var CardsScope = []string{
	"https://fabric.anz.com/scopes/cards:read",
	"https://fabric.anz.com/scopes/cards:create",
	"https://fabric.anz.com/scopes/cards:update",
}

var CardControlScope = []string{
	"https://fabric.anz.com/scopes/cards:read",
	"https://fabric.anz.com/scopes/cardControls:read",
	"https://fabric.anz.com/scopes/cardControls:create",
	"https://fabric.anz.com/scopes/cardControls:update",
	"https://fabric.anz.com/scopes/cardControls:delete",
}

const (
	V1beta1CardAPI         = "v1beta1CardAPI"
	V1beta2CardControlsAPI = "V1beta2CardControlsAPI"
	CallBack               = "CallBack"
)

//go:embed users.yaml
var usersbytes []byte

func allUsers(t *testing.T) anzdata.Users {
	var out anzdata.Users
	if err := yaml.Unmarshal(usersbytes, &out); err != nil {
		t.Fatal("GetUser: unable to load in user data", err)
	}
	return out
}

func GetUser(t *testing.T, cfg config.Service, conn *grpc.ClientConn, test string) anzdata.User {
	if cfg.Auth.FromPool {
		t.Log("GetUser: Getting a user from pool")
		users, err := allUsers(t).MatchAll("Region", cfg.Auth.Region)
		if err != nil {
			t.Fatal("GetUser: unable to load in users", err)
		}

		n := 0 // the condition is to break the loop in case all users return error when list card
		for n < len(users) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()
			user := users[gofakeit.IntRange(0, len(users)-1)]
			auth := GetAuthHeaders(t, user, cfg.Auth, CardsScope...)
			if test == V1beta1CardAPI {
				ctx = auth.Context(t, ctx, cfg.Headers...)
			} else {
				ctx = auth.Context(t, ctx)
			}
			cardsClient := v1beta1.NewGRPCClient(ctx, conn, nil)
			if _, err = cardsClient.List(); err != nil {
				n++
				cancel()
				continue
			}
			if test == V1beta1CardAPI {
				cardsClient.LoadCard(t)
				if cardsClient.Can(epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST) {
					return user
				}
				n++
				cancel()
				continue
			}
			return user
		}
	}

	out, err := anzdata.AllUsers().Match("PersonaID", cfg.Auth.PersonaID)
	if err == nil {
		t.Log("getting user from anz data:\n")
		return out
	}

	t.Log("creating a random user:\n")
	return anzdata.MustRandomUserFromInt(gofakeit.Number(0, 30))
}

package main

import (
	"flag"
	"fmt"

	"github.com/anzx/anzdata"
)

var name = flag.String("name", "CAMERON V", "the name of the testing user")

func main() {
	flag.Parse()
	scopes := []string{
		"https://fabric.anz.com/scopes/cards:read",
		"https://fabric.anz.com/scopes/cards:create",
		"https://fabric.anz.com/scopes/cards:update",
		"https://fabric.anz.com/scopes/cardControls:read",
		"https://fabric.anz.com/scopes/cardControls:create",
		"https://fabric.anz.com/scopes/cardControls:update",
		"https://fabric.anz.com/scopes/cardControls:delete",
	}
	user := anzdata.AllUsers().MustMatch("Name", name)
	authOpt := anzdata.AuthJWT{
		Claims: map[string]interface{}{
			"scopes": scopes,
		},
	}
	jwt := user.MustAuth(authOpt)
	fmt.Printf("Bearer %s", jwt)
}

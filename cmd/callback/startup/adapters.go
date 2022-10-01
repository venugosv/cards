package startup

import (
	"context"
	"fmt"

	"github.com/anzx/pkg/gsm"

	"github.com/anzx/fabric-cards/pkg/integration/fakerock"
	"github.com/anzx/fabric-cards/pkg/integration/forgerock"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/vault"
	anzerrors "github.com/anzx/pkg/errors"

	"github.com/anzx/fabric-cards/cmd/callback/config/app"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
)

type Adapters struct {
	CTM           ctm.Client
	CommandCentre *commandcentre.Client
	Vault         vault.Client
	Forgerock     forgerock.Clienter
	Fakerock      *fakerock.Client
}

func NewAdapters(ctx context.Context, config app.Spec, gsmClient *gsm.Client) (*Adapters, error) {
	var adapters Adapters

	commandCentreClient, err := commandcentre.NewClient(ctx, config.CommandCentre)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure CommandCentre client with config %+v", config.CommandCentre))
	}
	adapters.CommandCentre = commandCentreClient

	vaultClient, err := vault.NewClient(ctx, nil, config.Vault)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Vault client with config %+v", config.Vault))
	}
	adapters.Vault = vaultClient

	adapters.CTM, err = ctm.ClientFromConfig(ctx, nil, config.CTM, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure CTM Client with config %+v", config.CTM))
	}

	forgerockClient, err := forgerock.ClientFromConfig(ctx, nil, config.Forgerock, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure forgerock Client with config %+v", config.Forgerock))
	}
	adapters.Forgerock = forgerockClient

	fakerockClient, err := fakerock.NewClient(ctx, config.Fakerock)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure fakerock client with config %v", config.Fakerock))
	}
	adapters.Fakerock = fakerockClient

	return &adapters, nil
}

func anzErr(err error, msg string) error {
	return anzerrors.Wrap(err, anzerrors.GetStatusCode(err), msg, anzerrors.GetErrorInfo(err))
}

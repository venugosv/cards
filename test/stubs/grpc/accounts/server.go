package accounts

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/anzdata"
	"github.com/anzx/fabric-cards/test/stubs/utils"

	anzerrors "github.com/anzx/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/identity"
	apb "github.com/anzx/fabricapis/pkg/fabric/service/accounts/v1alpha6"
	pbtype "github.com/anzx/fabricapis/pkg/fabric/type"
)

const cardControlsUser = "d91acf54-4c87-48aa-85a9-dd41c72c54e6"

type StubServer struct {
	apb.UnimplementedAccountAPIServer
	store *utils.Store
}

// NewStubServer creates a CardEntitlementsAPIClient stubs
func NewStubServer(ctx context.Context) apb.AccountAPIServer {
	return &StubServer{
		store: utils.GetStore(ctx),
	}
}

func (s *StubServer) GetAccountList(ctx context.Context, in *apb.GetAccountListRequest) (*apb.GetAccountListResponse, error) {
	user, err := s.getUserData(ctx)
	if err != nil {
		return nil, err
	}

	accountsList := make([]*apb.AccountDetails, 0, len(user.Accounts))
	for _, account := range user.Accounts {
		details := apb.AccountDetails{
			AccountNumber: account.Number,
			Bsb:           account.BSB,
			Name:          user.Name,
			Balance: &pbtype.Money{
				CurrencyCode: pbtype.Currency_AUD,
			},
			CurrentBalance: &pbtype.Money{
				CurrencyCode: pbtype.Currency_AUD,
			},
		}
		if len(account.CardTokens) == 0 {
			details.Goal = &apb.Goal{}
		}
		accountsList = append(accountsList, &details)
	}

	if user.PersonaID == cardControlsUser {
		accountsList = append(accountsList, &apb.AccountDetails{
			AccountNumber: "1234567890",
			Bsb:           "4111",
			Name:          "Card Controls User",
		})
	}

	return &apb.GetAccountListResponse{
		AccountList: accountsList,
	}, nil
}

func (s *StubServer) getUserData(ctx context.Context) (anzdata.User, error) {
	id, err := identity.Get(ctx)
	if err != nil {
		logf.Error(ctx, err, "failed to get user data")
		return anzdata.User{}, anzerrors.Wrap(err, anzerrors.GetStatusCode(err), "entitlements failed", anzerrors.GetErrorInfo(err))
	}

	return s.store.GetUser(id.PersonaID), nil
}

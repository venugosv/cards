package ocv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/anzx/pkg/log"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"

	"github.com/anzx/fabric-cards/pkg/rest"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/pkg/integration/util"
	"github.com/anzx/fabric-cards/pkg/util/apic"
)

const (
	maintainAPI  = "ocv-maintain-contract-api"
	retrieveAPI  = "ocv-retrieve-party-api"
	CTM          = "CTM"
	PDV          = "PDV"
	oneOhOne     = "101"
	CAPCIS       = "CAP-CIS"
	SOL          = "SOL"
	CVDC         = "CVDC"
	channel      = "channel"
	userID       = "userid"
	requestTime  = "RequestTime"
	timeFormat   = "2006-01-02 15:04:05.000"
	anzx         = "ANZX"
	userIDFabric = "ANZxFabric"
)

type Config struct {
	BaseURL        string `json:"baseURL"             yaml:"baseURL"             mapstructure:"baseURL"         validate:"required"`
	ClientIDEnvKey string `json:"clientIDEnvKey"      yaml:"clientIDEnvKey"      mapstructure:"clientIDEnvKey"  validate:"required"`
	MaxRetries     int    `json:"maxRetries"          yaml:"maxRetries"          mapstructure:"maxRetries"      validate:"required"`
	EnableLogging  bool   `json:"enableLogging"       yaml:"enableLogging"       mapstructure:"enableLogging"`
}

type MaintainContractAPI interface {
	AccountMaintenance(ctx context.Context, ocvID string, oldCard *ctm.DebitCardResponse, newCard *ctm.DebitCardResponse, account *RetrievePartyRsAccount) (bool, error)
}

type RetrievePartyAPI interface {
	RetrieveParty(ctx context.Context, ovcID string) ([]*RetrievePartyRs, error)
}

type Client interface {
	MaintainContractAPI
	RetrievePartyAPI
}

type client struct {
	baseURL    string
	apicClient apic.Clienter
}

func ClientFromConfig(ctx context.Context, httpClient *http.Client, config *Config, gsmClient *gsm.Client) (Client, error) {
	if config == nil {
		logf.Debug(ctx, "ocv config not provided %v", config)
		return nil, nil
	}

	return NewClient(ctx, config.BaseURL, config.ClientIDEnvKey, httpClient, config.MaxRetries, gsmClient, config.EnableLogging)
}

func NewClient(ctx context.Context, baseURL string, clientIDEnvKey string, httpClient *http.Client, maxRetries int, gsmClient *gsm.Client, requestLogging bool) (Client, error) {
	if httpClient == nil {
		httpClient = rest.NewHTTPClientWithLogAndRetry(maxRetries, func(_ *url.URL) bool { return requestLogging }, names.OCV)
	}

	apicClient, err := apic.NewAPICClient(ctx, clientIDEnvKey, httpClient, gsmClient)
	if err != nil {
		return nil, err
	}

	return &client{
		baseURL:    baseURL,
		apicClient: apicClient,
	}, nil
}

func (c client) RetrieveParty(ctx context.Context, ocvID string) ([]*RetrievePartyRs, error) {
	destination := fmt.Sprintf("%s/%s/parties/retrieve?includeAccounts=true", c.baseURL, retrieveAPI)

	parse, err := url.Parse(destination)
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "Creating retrieve party request for given OCV ID", log.Str("ocvid", ocvID))
	payload := c.getPartyRequest(ocvID)

	req := apic.NewRequest(http.MethodPost, parse.String(), payload)
	req.Headers = map[string]string{
		channel: anzx,
		userID:  userIDFabric,
	}

	resp, err := c.apicClient.Do(ctx, req, "ocv:MaintainAccount")
	if err != nil {
		return nil, err
	}

	var out []*RetrievePartyRs
	if err = json.Unmarshal(resp, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c client) getPartyRequest(ocvID string) []byte {
	payload := RetrievePartyRq{
		Identifiers: []*IdentifierRq{
			{
				IdentifierUsageType: util.ToStringPtr("One Customer ID"),
				Identifier:          util.ToStringPtr(ocvID),
			},
		},
	}
	bytes, _ := json.Marshal(payload)
	return bytes
}

func (c client) AccountMaintenance(ctx context.Context, ocvID string, oldCard *ctm.DebitCardResponse, newCard *ctm.DebitCardResponse, account *RetrievePartyRsAccount) (bool, error) {
	payload := c.getAccountMaintenanceRequest(oldCard, newCard, ocvID, account)

	destination := fmt.Sprintf("%s/%s/accounts-maintenance", c.baseURL, maintainAPI)

	request := apic.NewRequest(http.MethodPost, destination, payload)

	request.Headers = map[string]string{
		requestTime: time.Now().Format(timeFormat),
		channel:     anzx,
		userID:      userIDFabric,
	}

	_, err := c.apicClient.Do(ctx, request, "ocv:MaintainAccount")

	return err == nil, err
}

func (c client) getAccountMaintenanceRequest(oldCard *ctm.DebitCardResponse, newCard *ctm.DebitCardResponse, ocvID string, linkedAccount *RetrievePartyRsAccount) []byte {
	today := time.Now().Format("2006-01-02")

	request := MaintainContractRequest{
		AccountNumber: util.ToStringPtr(linkedAccount.AccountNumber),
		AccountType:   accountTypeAccount,
		AccountSource: CAPCIS,
		ProductCode:   productCodeDDA,
		LinkedAccounts: []LinkedAccount{
			{
				Account: Account{
					AccountNumber:     &oldCard.CardNumber.Token,
					AccountType:       accountTypeService,
					ProductCode:       oldCard.ProductCode,
					AccountSubProduct: util.ToStringPtr(oldCard.SubProductCode),
					AccountClosedDate: util.ToStringPtr(today),
					AccountStatus:     &oldCard.StatusCode, // we need to make sure the OLDCARD object is returned AFTER the replacement request
					AccountSource:     CTM,
					LinkedParties: []LinkedParty{
						{
							RelationshipType: util.ToStringPtr(SOL),
							EndDate:          util.ToStringPtr(today),
							Party: Party{
								OcvID: ocvID,
							},
						},
					},
				},
				AccountRelationships: []AccountRelationships{
					{
						EndDate:           today,
						Status:            accountRelationshipStatusCancelled,
						RelationshipValue: relationshipValueComponentOf,
					},
				},
			},
			{
				Account: Account{
					AccountNumber:       util.ToStringPtr(newCard.CardNumber.Token),
					AccountType:         accountTypeService,
					AccountNameOne:      util.ToStringPtr(linkedAccount.AccountNameOne),
					AccountNameTwo:      util.ToStringPtr(linkedAccount.AccountNameTwo),
					AccountNameThree:    util.ToStringPtr(linkedAccount.AccountNameThree),
					AccountShortName:    util.ToStringPtr(linkedAccount.AccountShortName),
					ProductCode:         PDV,
					AccountSubProduct:   util.ToStringPtr(oneOhOne),
					AccountOpenedDate:   util.ToStringPtr(today),
					AccountStatus:       &newCard.StatusCode,
					MarketingCode:       util.ToStringPtr(CVDC),
					AccountBranchNumber: util.ToStringPtr(strconv.FormatInt(int64(newCard.CollectionBranch), 10)),
					AccountSource:       CTM,
					LinkedParties: []LinkedParty{
						{
							RelationshipType: util.ToStringPtr(SOL),
							StartDate:        util.ToStringPtr(today),
							Party: Party{
								OcvID: ocvID,
							},
						},
					},
				},
				AccountRelationships: []AccountRelationships{
					{
						StartDate:         today,
						Status:            accountRelationshipStatusActive,
						RelationshipValue: relationshipValueComponentOf,
					},
				},
			},
		},
	}

	body, _ := json.Marshal(request)

	return body
}

func GetAccount(parties []*RetrievePartyRs, accountNumber string) (*RetrievePartyRsAccount, error) {
	accounts, err := GetAccounts(parties, []string{accountNumber})
	if err != nil {
		return nil, err
	}
	return accounts[0], nil
}

func GetAccounts(parties []*RetrievePartyRs, accountNumbers []string) ([]*RetrievePartyRsAccount, error) {
	var accounts []*RetrievePartyRsAccount
	for _, party := range parties {
		for _, accountNumber := range accountNumbers {
			if account := party.GetAccount(accountNumber); account != nil {
				accounts = append(accounts, account)
				break
			}
		}
	}

	if len(accounts) == 0 {
		return nil, anzerrors.New(codes.NotFound, "failed to get party and account",
			anzerrors.NewErrorInfo(context.Background(), anzcodes.Unknown, "unable to get party and account information for card"))
	}

	return accounts, nil
}

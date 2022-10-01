package ocv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/anzx/fabric-cards/test/stubs/utils"

	"github.com/anzx/fabric-cards/pkg/integration/ocv"
)

type StubServer struct {
	store *utils.Store
}

func NewStubServer(ctx context.Context) *StubServer {
	return &StubServer{
		store: utils.GetStore(ctx),
	}
}

func AppendRoutes(ctx context.Context, router *http.ServeMux) {
	router.Handle("/ocv/", handler(ctx))
}

func handler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewStubServer(ctx).ServeHTTP(w, r)
	})
}

func (s *StubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		accountMaintenanceURL = regexp.MustCompile("/accounts-maintenance")
		retrievePartyURL      = regexp.MustCompile("/ocv-retrieve-party-api/parties/retrieve")
	)

	switch {
	case accountMaintenanceURL.MatchString(r.URL.Path):
		s.accountMaintenanceHandler(w, r)
	case retrievePartyURL.MatchString(r.URL.Path):
		s.retrievePartyHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *StubServer) accountMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := utils.GetRequestBody(w, r, http.MethodPost)
	if !ok {
		return
	}
	defer r.Body.Close()

	var request ocv.MaintainContractRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	log.Printf("ocv maintain contract request: %+v\n", request)

	w.WriteHeader(http.StatusOK)
}

func (s *StubServer) retrievePartyHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := utils.GetRequestBody(w, r, http.MethodPost)
	if !ok {
		return
	}
	defer r.Body.Close()

	var request ocv.RetrievePartyRq
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	log.Printf("ocv party request: %+v\n", request)

	user := s.store.GetUserByOCV(*request.Identifiers[0].Identifier)

	response := &ocv.RetrievePartyRs{
		Identifiers: []*ocv.Identifier{
			{
				Identifier:          user.OCVID,
				IdentifierUsageType: "One Customer ID",
			}, {
				Identifier:          user.CAPCISID,
				IdentifierUsageType: "CAP ID",
				Source:              "CAP-CIS",
			},
		},
		Source: "CAP-CIS",
		SourceSystems: []*ocv.SourceSystem{
			{
				SourceSystemID:   user.CAPCISID,
				SourceSystemName: "CAP-CIS",
			},
		},
	}

	for _, account := range user.Accounts {
		outAccount := &ocv.RetrievePartyRsAccount{
			AccountBranchNumber: account.BSB,
			AccountNameOne:      user.Name,
			AccountNumber:       account.Number,
			AccountOpenedDate:   "2012-06-08",
			AccountSubProduct:   "CAP-CIS:DDASA",
			CompanyID:           "10",
			ProductCode:         "CAP-CIS:DDA",
			RelationshipType:    "SOL",
		}
		response.Accounts = append(response.Accounts, outAccount)
	}
	for _, card := range user.Cards {
		cardAccount := &ocv.RetrievePartyRsAccount{
			AccountNameOne:    user.Name,
			AccountNumber:     fmt.Sprintf("enc(%s)", card.Token),
			AccountOpenedDate: "2012-06-08",
			AccountSubProduct: "CAP-CIS:DDASA",
			CompanyID:         "10",
			ProductCode:       "CAP-CIS:DDA",
			RelationshipType:  "SOL",
		}
		response.Accounts = append(response.Accounts, cardAccount)
	}

	for _, card := range user.Cards {
		cardAccount := &ocv.RetrievePartyRsAccount{
			AccountNameOne:    user.Name,
			AccountNumber:     fmt.Sprintf("enc(%s)", card.Token),
			AccountOpenedDate: "2012-06-08",
			AccountSubProduct: "CAP-CIS:DDASA",
			CompanyID:         "10",
			ProductCode:       "CAP-CIS:DDA",
			RelationshipType:  "SOL",
		}
		response.Accounts = append(response.Accounts, cardAccount)
	}

	out, err := json.Marshal([]*ocv.RetrievePartyRs{response})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, _ = w.Write(out)
}

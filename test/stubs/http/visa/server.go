package visa

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"

	"github.com/anzx/fabric-cards/test/stubs/utils"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
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
	router.Handle("/vctc/", handler(ctx))
}

func handler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewStubServer(ctx).ServeHTTP(w, r)
	})
}

func (d *StubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		update     = regexp.MustCompile("/vctc/customerrules/consumertransactioncontrols/.*/rules")
		queryByPAN = "/vctc/customerrules/consumertransactioncontrols/inquiries/cardinquiry"
		enrolByPAN = "/vctc/customerrules/consumertransactioncontrols"
		replace    = "/vctc/customerrules/consumertransactioncontrols/accounts/accountupdate"
	)

	switch {
	case r.URL.Path == queryByPAN:
		d.query(w, r)
	case r.URL.Path == enrolByPAN:
		d.enrol(w, r)
	case r.URL.Path == replace:
		d.replace(w, r)
	case update.MatchString(r.URL.Path):
		d.update(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (d *StubServer) query(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("unable to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var request visa.QueryRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	card := d.store.GetCard(request.PrimaryAccountNumber)

	var response *visa.TransactionControlListResponses
	if card.CardControlPreference {
		merchantControls := []ccpb.ControlType{ccpb.ControlType_MCT_ALCOHOL, ccpb.ControlType_MCT_ADULT_ENTERTAINMENT, ccpb.ControlType_MCT_SMOKE_AND_TOBACCO}
		transactionControls := []ccpb.ControlType{ccpb.ControlType_TCT_E_COMMERCE, ccpb.ControlType_TCT_AUTO_PAY, ccpb.ControlType_TCT_ATM_WITHDRAW}
		response = NewTransactionControlDocumentListResponseFixture(NewResourceFixture().WithGlobalControls().WithMerchantControls(merchantControls...).WithTransactionControls(transactionControls...).Build()).Build()
	} else {
		NewTransactionControlDocumentListResponseFixture(NewResourceFixture().WithDocumentID("NOT_ENROLLED").Build()).Build()
	}

	respond(w, response)
}

func (d *StubServer) enrol(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("unable to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var request visa.EnrolByPanRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	card := d.store.GetCard(request.PrimaryAccountNumber)

	var response *visa.TransactionControlDocument
	if card.CardControlPreference {
		response = NewTransactionControlDocumentResponseFixture(NewResourceFixture().WithDocumentID(DocumentID).Build()).Build()
	} else {
		response = NewTransactionControlDocumentResponseFixture(NewResourceFixture().WithDocumentID("NOT_ENROLLED").Build()).Build()
	}

	respond(w, response)
}

func (d *StubServer) replace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("unable to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var request visa.ReplacementRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	card := d.store.GetCard(request.NewAccountID)

	var response *visa.AccountUpdateResponse
	if card.CardControlPreference {
		response = NewAccountUpdateResponseFixture().WithStatus("SUCCESS").Build()
	} else {
		response = NewAccountUpdateResponseFixture().WithStatus("FAILED").Build()
	}

	respond(w, response)
}

func (d *StubServer) update(w http.ResponseWriter, r *http.Request) {
	documentID := strings.Split(r.URL.Path, "/")[4]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("unable to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var resource visa.Resource
	if err := json.Unmarshal(body, &resource); err != nil {
		log.Printf("unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	resource.DocumentID = documentID
	resource.LastUpdateTimeStamp = time.Now().UTC().String()

	response := &visa.TransactionControlDocument{
		ReceivedTimestamp:  time.Now().UTC().String(),
		ProcessingTimeInMS: 72,
		Resource:           resource,
	}
	respond(w, response)
}

func respond(w http.ResponseWriter, documentResponse interface{}) {
	marshal, _ := json.Marshal(documentResponse)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(marshal)
}

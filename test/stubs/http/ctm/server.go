package ctm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/anzx/fabric-cards/test/stubs/utils"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
)

const (
	preferences      = "preferences"
	replace          = "replace"
	activation       = "activate"
	status           = "status"
	pinInfoUpdate    = "pin-info/update"
	debitCards       = "debit-cards"
	maintenanceAPI   = "debit-card-maintenance"
	inquiryAPI       = "debit-card-inquiry"
	statusAPI        = "debit-card-status"
	pinInfoUpdateAPI = "debit-card-pin-info-update"
)

type StubServer struct {
	ctx   context.Context
	store *utils.Store
}

func NewStubServer(ctx context.Context) *StubServer {
	return &StubServer{
		ctx:   ctx,
		store: utils.GetStore(ctx),
	}
}

func AppendRoutes(ctx context.Context, router *http.ServeMux) {
	router.Handle("/ctm/", handler(ctx))
}

func handler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewStubServer(ctx).ServeHTTP(w, r)
	})
}

func (d *StubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		debitCardInquiryURL = regexp.MustCompile(fmt.Sprintf("/%s/%s/\\d{16}", inquiryAPI, debitCards))
		activateURL         = regexp.MustCompile(fmt.Sprintf("/%s/%s/\\d{16}/%s", statusAPI, debitCards, activation))
		statusURL           = regexp.MustCompile(fmt.Sprintf("/%s/%s/\\d{16}/%s", statusAPI, debitCards, status))
		replaceCardURL      = regexp.MustCompile(fmt.Sprintf("/%s/%s/\\d{16}/%s", maintenanceAPI, debitCards, replace))
		preferencesCardURL  = regexp.MustCompile(fmt.Sprintf("/%s/%s/\\d{16}/%s", maintenanceAPI, debitCards, preferences))
		pinInfoUpdateURL    = regexp.MustCompile(fmt.Sprintf("/%s/%s/\\d{16}/%s", pinInfoUpdateAPI, debitCards, pinInfoUpdate))
	)

	switch {
	case debitCardInquiryURL.MatchString(r.URL.Path):
		d.debitCardInqHandler(w, r)
	case activateURL.MatchString(r.URL.Path):
		d.activateHandler(w, r)
	case statusURL.MatchString(r.URL.Path):
		d.updateStatusHandler(w, r)
	case replaceCardURL.MatchString(r.URL.Path):
		d.replaceHandler(w, r)
	case preferencesCardURL.MatchString(r.URL.Path):
		d.preferencesHandler(w, r)
	case pinInfoUpdateURL.MatchString(r.URL.Path):
		d.pinInfoUpdateHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (d *StubServer) preferencesHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := utils.GetRequestBody(w, r, http.MethodPatch)
	if !ok {
		return
	}
	defer r.Body.Close()

	tokenizedCardNumber := extractTokenizedCardNumber(r.URL.Path)

	var request ctm.UpdatePreferencesRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logf.Error(d.ctx, err, "unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	logf.Debug(d.ctx, "preferences request: %+v\n", request)

	card := d.store.GetCard(tokenizedCardNumber)
	if card == nil {
		w.WriteHeader(http.StatusNotFound)
	}

	if request.CardControlPreference != nil {
		card.CardControlPreference = *request.CardControlPreference
	}

	if request.MerchantUpdatePreference != nil {
		card.MerchantUpdatePreference = *request.MerchantUpdatePreference
	}

	d.store.SaveCard(tokenizedCardNumber, card)

	w.WriteHeader(http.StatusOK)
}

func (d *StubServer) replaceHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := utils.GetRequestBody(w, r, http.MethodPost)
	if !ok {
		return
	}
	defer r.Body.Close()

	oldTokenizedCardNumber := extractTokenizedCardNumber(r.URL.Path)

	var request ctm.ReplaceCardRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logf.Error(d.ctx, err, "unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	logf.Debug(d.ctx, "replace request: %+v\n", request)

	oldCard := d.store.GetCard(oldTokenizedCardNumber)

	if request.PlasticType != ctm.NewNumber {
		responseData, err := json.Marshal(ctm.ReplaceCardResponse{})
		if err != nil {
			logf.Error(d.ctx, err, "failed to marshal response body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseData)
		return
	}

	newTokenizedCardNumber := gofakeit.CreditCardNumber(&gofakeit.CreditCardOptions{Types: []string{"visa"}})

	newCard := d.store.GetCard(newTokenizedCardNumber)
	newCard.OldCardNumber = &ctm.Card{
		Token:       oldTokenizedCardNumber,
		Last4Digits: oldCard.CardNumber.Last4Digits,
	}

	oldCard.NewCardNumber = &ctm.Card{
		Token:       newTokenizedCardNumber,
		Last4Digits: newCard.CardNumber.Last4Digits,
	}

	d.store.SaveCard(oldTokenizedCardNumber, oldCard)
	d.store.SaveCard(newTokenizedCardNumber, newCard)

	out := ctm.ReplaceCardResponse{
		CardNumber: ctm.CardNumber{
			Token:       newTokenizedCardNumber,
			Last4Digits: newCard.CardNumber.Last4Digits,
		},
	}

	responseData, err := json.Marshal(out)
	if err != nil {
		logf.Error(d.ctx, err, "failed to marshal response body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseData)
}

func (d *StubServer) updateStatusHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := utils.GetRequestBody(w, r, http.MethodPatch)
	if !ok {
		return
	}
	defer r.Body.Close()

	tokenizedCardNumber := extractTokenizedCardNumber(r.URL.Path)

	var request ctm.UpdateDebitCardStatusRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logf.Error(d.ctx, err, "unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	logf.Debug(d.ctx, "update request: %+v\n", request)

	card := d.store.GetCard(tokenizedCardNumber)
	if card == nil {
		logf.Debug(d.ctx, "requested card: %s not found", tokenizedCardNumber)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	card.Status = request.Status

	d.store.SaveCard(tokenizedCardNumber, card)

	logf.Debug(d.ctx, "card new status: %s\n", card.Status)

	w.WriteHeader(http.StatusOK)
}

func (d *StubServer) activateHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := utils.GetRequestBody(w, r, http.MethodPost)
	if !ok {
		return
	}
	defer r.Body.Close()

	tokenizedCardNumber := extractTokenizedCardNumber(r.URL.Path)

	var request ctm.ActivateDebitCardRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logf.Error(d.ctx, err, "unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	logf.Debug(d.ctx, "active request: %+v\n", request)

	card := d.store.GetCard(tokenizedCardNumber)
	if card == nil {
		logf.Debug(d.ctx, "requested card: %s not found", tokenizedCardNumber)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	card.ActivationStatus = true

	d.store.SaveCard(tokenizedCardNumber, card)

	t := d.store.GetCard(tokenizedCardNumber)
	if !t.ActivationStatus {
		logf.Err(d.ctx, fmt.Errorf("activation status %v", t.ActivationStatus))
	}

	logf.Debug(d.ctx, "card new activation status: %v\n", card.ActivationStatus)

	w.WriteHeader(http.StatusOK)
}

type DebitCardInquiryRequest struct {
	CardNumber string `json:"cardNumber"`
	Detail     bool   `json:"detail"`
}

func (d *StubServer) debitCardInqHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logf.Debug(d.ctx, "method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := extractTokenizedCardNumber(r.URL.Path)
	if token == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	logf.Debug(d.ctx, "inquiry request: %+v\n", r.URL)

	card := d.store.GetCard(token)

	responseData, err := json.Marshal(card)
	if err != nil {
		logf.Error(d.ctx, err, "failed to marshal response body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseData)
}

func (d *StubServer) pinInfoUpdateHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := utils.GetRequestBody(w, r, http.MethodPatch)
	if !ok {
		return
	}
	defer r.Body.Close()

	tokenizedCardNumber := extractTokenizedCardNumber(r.URL.Path)

	var request ctm.PINInfoUpdateRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logf.Error(d.ctx, err, "unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	logf.Debug(d.ctx, "pin info update request: %+v\n", request)

	card := d.store.GetCard(tokenizedCardNumber)
	if card == nil {
		logf.Debug(d.ctx, "requested card: %s not found", tokenizedCardNumber)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	card.PinChangedCount++

	d.store.SaveCard(tokenizedCardNumber, card)

	logf.Debug(d.ctx, "new pin changed count: %v\n", card.PinChangedCount)

	w.WriteHeader(http.StatusOK)
}

func extractTokenizedCardNumber(path string) string {
	const expr = "[0-9]{16}"
	re := regexp.MustCompile(expr)

	return re.FindString(path)
}

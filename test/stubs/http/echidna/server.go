package echidna

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/anzx/fabric-cards/test/stubs/utils"

	"github.com/anzx/fabric-cards/pkg/integration/echidna"
)

const (
	encodedKey     = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArI6WKTJMLVfpaG+Mkaj4IVX3/2dbtHvacI9sKfutMsg5It6pEvFf9oYoIWMQkxFARf14ds0+1t83sm6foPHm4HZ0oP2GX0iiFdALEZr3C6C2FXAoQQXYMGeczoeta0IwF75B3Pr6VETQjf7niL00MF0n/McsE9tu9VTOFjq6LkvZgOnBe9wG+f0nvdx29FAPzIjdpBoZ27Ingmtnmtk2T9oadY5vXE2ruIhjU2rL/8aPPN8LtvlWrcV0y+YW2l4EMGenAFYMu4jh6R5deNfartmNotJgbzHFcD7EpXJivzYgdMvea2Dy7AjlC5cic4ijcna750HhfMoFFNqf6T7psQIDAQAB"
	basepath       = "ca"
	apipath        = "card-and-pin-services"
	getWrappingKey = "getWrappingKey"
	selectPIN      = "selectPIN"
	verifyPIN      = "verifyPIN"
	changePIN      = "changePIN"
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
	router.Handle("/ca/", handler(ctx))
}

func handler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewStubServer(ctx).ServeHTTP(w, r)
	})
}

func (d *StubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		getWrappingKeyURL = fmt.Sprintf("/%s/%s/%s", basepath, apipath, getWrappingKey)
		selectPINURL      = fmt.Sprintf("/%s/%s/%s", basepath, apipath, selectPIN)
		verifyPINURL      = fmt.Sprintf("/%s/%s/%s", basepath, apipath, verifyPIN)
		changePINURL      = fmt.Sprintf("/%s/%s/%s", basepath, apipath, changePIN)
	)

	if invalidRequest(w, r) {
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var request echidna.Request
	if err := json.Unmarshal(bytes, &request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	card := d.store.GetCard(request.PlainPAN)

	var body interface{}
	switch {
	case r.URL.Path == getWrappingKeyURL:
		body = echidna.GetWrappingKeyResponse{
			Response: getResponse(echidna.ActionGetWrappingKey, func(s string) *string { return &s }(encodedKey)),
		}
	case r.URL.Path == selectPINURL:
		body = echidna.SetPINResponse{
			Response: getResponse(echidna.ActionSelect, nil),
		}
		card.PinChangedCount++
	case r.URL.Path == verifyPINURL:
		body = echidna.VerifyPINResponse{
			Response: getResponse(echidna.ActionVerify, nil),
		}
	case r.URL.Path == changePINURL:
		body = echidna.ChangePINResponse{
			Response: getResponse(echidna.ActionChange, nil),
		}
		card.PinChangedCount++
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	d.store.SaveCard(request.PlainPAN, card)

	response, _ := json.Marshal(body)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func getResponse(method echidna.Action, encodedKey *string) echidna.Response {
	response := echidna.Response{
		Method: method,
		Result: echidna.Result{
			Code:    0,
			Message: fmt.Sprintf("%s operation successful.", method),
		},
		LogMessages: echidna.LogMessages{
			WantLevel: echidna.LoglevelInfo,
		},
	}
	if encodedKey != nil {
		response.Result.EncodedKey = encodedKey
	}
	return response
}

func invalidRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		log.Printf("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return true
	}

	if get := r.Header.Get("x-request-id"); get == "" {
		log.Printf("requestID not provided")
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	return false
}

package apcam

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/anzx/fabric-cards/pkg/integration/apcam"
)

const (
	pushProvisionURL   = "/apcam/internal/in-app-provisioning-data"
	encryptedPassData  = "TUJQQUQtMS1GSy0xMjM0NTYuMS0tVERFQS03QUYyOTFDOTFGM0VENEVGOTJDMUQ0NUVGRjEyN0MxRjlBQkMxMjM0N0U="                                                                                                                                                                                                                                                                                                                                                                                                                 //nolint:gosec
	activationData     = "QUJDREVGLTEtRkstMTIzNDU2LjEtLVRERUEtN0FGMjkxQzkxRjNFRDRFRjkyQzFENDVFRkYxMjdDMUY5QUJDMTIzNDdF"                                                                                                                                                                                                                                                                                                                                                                                                                 //nolint:gosec
	ephemeralPublicKey = "UVVKRFJFVkdMVEV0UmtzdE1USXpORFUyTGpFdExWUkVSVUV0TjBGR01qa3hRemt4UmpORlJEUkZSamt5UXpGRU5EVkZSa1l4TWpkRE1VWTVRVUpETVRJek5EZEZRVUpEUkVWR0xURXRSa3N0TVRJek5EVTJMakV0TFZSRVJVRXROMEZHTWpreFF6a3hSak5GUkRSRlJqa3lRekZFTkRWRlJrWXhNamRETVVZNVFVSkRNVEl6TkRkRlFVSkRSRVZHTFRFdFJrc3RNVEl6TkRVMkxqRXRMVlJFUlVFdE4wRkdNamt4UXpreFJqTkZSRFJGUmpreVF6RkVORFZGUmtZeE1qZERNVVk1UVVKRE1USXpORGRGUVVKRFJFVkdMVEV0UmtzdE1USXpORFUyTGpFdExWUkVSVUV0TjBGR01qa3hRemt4UmpORlJEUkZSamt5UXpGRU5EVkZSa1l4TWpkRE1VWTVRVUpETVRJek5EZEY=" //nolint:gosec
	YYYYMM             = "[0-9]{4}-[0-9]{2}"
)

type StubServer struct{}

func NewStubServer() *StubServer {
	return &StubServer{}
}

func AppendRoutes(ctx context.Context, router *http.ServeMux) {
	router.Handle("/apcam/", handler(ctx))
}

func handler(_ context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewStubServer().ServeHTTP(w, r)
	})
}

func (d *StubServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != pushProvisionURL {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	request, ok := parseRequest(w, r)
	if !ok {
		return
	}

	if !regexp.MustCompile(YYYYMM).MatchString(request.CardInfo.ExpiryDate) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(apcam.ErrorInfo{
			ErrorCode:        "500",
			ErrorDescription: "Invalid Expiry date, must be in YYYY-MM form",
		})
		return
	}

	_ = json.NewEncoder(w).Encode(getResponse(request))
}

func getResponse(in *apcam.Request) *apcam.Response {
	return &apcam.Response{
		TraceInfo: apcam.TraceInfo{
			MessageID:      in.TraceInfo.MessageID,
			ConversationID: in.TraceInfo.ConversationID,
		},
		Apple: apcam.AppleResponseData{
			EncryptedPassData:  encryptedPassData,
			ActivationData:     activationData,
			EphemeralPublicKey: ephemeralPublicKey,
		},
	}
}

func parseRequest(w http.ResponseWriter, r *http.Request) (*apcam.Request, bool) {
	if r.Method != http.MethodPost {
		log.Printf("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("unable to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return nil, false
	}
	defer r.Body.Close()

	var request apcam.Request
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("unexpected request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return nil, false
	}

	return &request, true
}

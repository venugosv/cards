package utils

import (
	"io"
	"log"
	"net/http"
)

func GetRequestBody(w http.ResponseWriter, r *http.Request, methods ...string) ([]byte, bool) {
	validMethod := false
	for _, method := range methods {
		if r.Method == method {
			validMethod = true
			break
		}
	}

	if !validMethod {
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
	return body, true
}

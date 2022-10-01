package vault_external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/errors/errcodes"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

const (
	authLoginPath = "/login"
)

// VaultAPIer is an interface for calling the Vault API. This interface allows us to simplify working with the
//  API in tests.
type VaultAPIer interface {
	run(ctx context.Context, method string, path string, token string, body []byte) ([]byte, error)
	vaultLogin(ctx context.Context, role string, jwt string) (*Secret, error)
}

// VaultAPI contains the data we need to make HTTP requests against the Vault API
type VaultAPI struct {
	httpClient *http.Client
	address    string
	namespace  string
	loginPath  string
}

// run makes a HTTP request against the Vault API
func (v *VaultAPI) run(ctx context.Context, method string, path string, token string, body []byte) ([]byte, error) {
	// Set up a context that we can cancel, to ensure any HTTP resources are cleaned up
	newCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	reader := bytes.NewBuffer(body)

	url := fmt.Sprintf("%s/%s", v.address, path)

	request, err := http.NewRequestWithContext(newCtx, method, url, reader)
	if err != nil {
		return nil, errors.Wrap(
			err,
			codes.Internal,
			"vault API request failed",
			errors.NewErrorInfo(ctx, errcodes.Unknown, "failed to create request"),
		)
	}

	request.Header.Set("X-Vault-Namespace", v.namespace)
	request.Header.Set("X-Vault-Token", token)
	request.Header.Set("X-Vault-Request", "true")
	request.Header.Set("x-request-id", uuid.New().String())

	response, err := v.httpClient.Do(request)
	// http.Client.Do() errors need wrapping to give more meaningful messages
	if err != nil {
		return nil, errors.Wrap(
			err,
			codes.Internal,
			"vault API request failed",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, "error making HTTP request to vault API"),
		)
	}
	// Status codes for failed requests require wrapping to give context
	statusOK := response.StatusCode >= http.StatusOK && response.StatusCode < 300
	if !statusOK {
		return nil, errors.New(
			apic.CodeFromHTTPStatus(response.StatusCode),
			"vault API request failed",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, fmt.Sprintf("vault API response status %d", response.StatusCode)),
		)
	}

	r, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(
			err,
			codes.Internal,
			"vault API request failed",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, "failed to read HTTP response body"),
		)
	}
	return r, nil
}

// vaultLogin calls the Vault API with our AuthRequest to get an AuthResponse
func (v *VaultAPI) vaultLogin(ctx context.Context, role string, jwt string) (*Secret, error) {
	authRequest := &AuthRequest{
		Role: role,
		JWT:  jwt,
	}
	requestBody, err := json.Marshal(authRequest)
	if err != nil {
		return nil, err
	}

	responseBody, err := v.run(ctx, http.MethodPost, v.loginPath, "", requestBody)
	if err != nil {
		return nil, err
	}

	var secret Secret
	unmarshalError := json.Unmarshal(responseBody, &secret)
	if unmarshalError != nil {
		return nil, errors.Wrap(
			unmarshalError,
			codes.Internal,
			"failed to get auth from vault login",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, fmt.Sprintf("failed to unmarshal JSON: %s", unmarshalError.Error())),
		)
	}
	return &secret, nil
}

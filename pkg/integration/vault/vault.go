package vault

import (
	"context"
	"fmt"
	"net/http"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/vault_external"
	"github.com/anzx/fabric-cards/pkg/rest"
	"github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/monitoring/names"
	"google.golang.org/grpc/codes"
)

// Client is the high level interface to encode and decode card numbers with Vault
type Client interface {
	EncodeCardNumber(ctx context.Context, cardNumber string) (string, error)
	DecodeCardNumber(ctx context.Context, cardNumber string) (string, error)
	EncodeCardNumbers(ctx context.Context, cardNumbers []string) (map[string]string, error)
	DecodeCardNumbers(ctx context.Context, cardNumbers []string) (map[string]string, error)
}

const transformRole = "transformrole.fabric.common"

// client is a simple wrapper over the external Vault Client interface. This allows us
//  to define a higher level interface that does the things we need.
type client struct {
	vault_external.Client
	role string
}

// NewClient creates a client based on the passed config. If httpClient is nil, a sensible default
//  is created. If config is nil, this function quietly fails and returns a nil Client
func NewClient(ctx context.Context, httpClient *http.Client, config *vault_external.Config) (Client, error) {
	if config == nil {
		logf.Debug(ctx, "vault config not provided %v", config)
		return nil, nil
	}

	if httpClient == nil {
		httpClient = rest.NewHTTPClientWithLogAndRetry(5, nil, names.HashiCorpVault)
	}

	vaultClient, err := vault_external.NewClient(ctx, httpClient, config)
	return client{
		Client: vaultClient,
		role:   config.AuthRole,
	}, err
}

// EncodeCardNumber encodes a single card number and returns the encoded value
func (c client) EncodeCardNumber(ctx context.Context, cardNumber string) (string, error) {
	encodedResponse, err := c.EncodeCardNumbers(ctx, []string{cardNumber})
	if err != nil {
		return "", err
	}
	// Get the response corresponding to the input cardNumber, only if it is present in the map
	encodedCardNumber, ok := encodedResponse[cardNumber]
	if !ok {
		return "", errors.New(
			codes.Internal,
			"failed to encode card number",
			errors.NewErrorInfo(ctx, errcodes.CardTokenizationFailed, "no card number in encode response"))
	}
	return encodedCardNumber, nil
}

// DecodeCardNumber decodes a single card number and returns the decoded value
func (c client) DecodeCardNumber(ctx context.Context, tokenizedCardNumber string) (string, error) {
	decodedCardNumbers, err := c.DecodeCardNumbers(ctx, []string{tokenizedCardNumber})
	if err != nil {
		return "", err
	}
	// Get the response corresponding to the input tokenizedCardNumber, only if it is present in the map
	if cardNumber, ok := decodedCardNumbers[tokenizedCardNumber]; ok {
		return cardNumber, nil
	} else {
		return "", errors.New(
			codes.Internal,
			"failed to decode card number",
			errors.NewErrorInfo(ctx, errcodes.CardTokenizationFailed, "no card number in decode response"),
		)
	}
}

func makeGenericTransformRequest(ctx context.Context, c client, values []string, kind vault_external.TransformKind) ([]*vault_external.TransformResult, error) {
	if len(values) == 0 {
		return nil, errors.New(
			codes.Internal,
			fmt.Sprintf("nothing to %s", kind.String()),
			errors.NewErrorInfo(ctx, errcodes.Unknown, "input list of values to transform was empty"),
		)
	}

	var batchRequest []*vault_external.TransformRequest
	for _, value := range values {
		if value == "" {
			return nil, errors.New(
				codes.Internal,
				"transform failed",
				errors.NewErrorInfo(ctx, errcodes.Unknown, "transform request contains empty string, arguments are invalid"),
			)
		}
		// The Value field is what gets encoded. The response contains a Reference field equal to whatever
		//  is in the request. We use this to create a mapping from the input to a final encoded value
		request := &vault_external.TransformRequest{
			Value:     value,
			Reference: value,
		}
		batchRequest = append(batchRequest, request)
	}
	return c.Transform(ctx, kind, transformRole, batchRequest)
}

// EncodeCardNumbers encodes a list of card numbers. It returns a map from the input cardNumbers to a
//  corresponding encoded value
func (c client) EncodeCardNumbers(ctx context.Context, cardNumbers []string) (map[string]string, error) {
	encodeResponse, err := makeGenericTransformRequest(ctx, c, cardNumbers, vault_external.TransformEncode)
	if err != nil {
		errInfo := errors.GetErrorInfo(err)
		return nil, errors.Wrap(
			err,
			errors.GetStatusCode(err),
			"failed to encode card numbers",
			errors.NewErrorInfo(ctx, errcodes.CardTokenizationFailed, errInfo.GetReason()),
		)
	}

	return mapResults(ctx, cardNumbers, encodeResponse)
}

// DecodeCardNumbers decodes a list of card numbers. It returns a map from the input numbers to a
//  corresponding decoded value
func (c client) DecodeCardNumbers(ctx context.Context, tokenizedCardNumbers []string) (map[string]string, error) {
	decodedResponse, err := makeGenericTransformRequest(ctx, c, tokenizedCardNumbers, vault_external.TransformDecode)
	if err != nil {
		errInfo := errors.GetErrorInfo(err)
		return nil, errors.Wrap(
			err,
			errors.GetStatusCode(err),
			"failed to decode card numbers",
			errors.NewErrorInfo(ctx, errcodes.CardTokenizationFailed, errInfo.GetReason()),
		)
	}

	return mapResults(ctx, tokenizedCardNumbers, decodedResponse)
}

func mapResults(ctx context.Context, request []string, results []*vault_external.TransformResult) (map[string]string, error) {
	if len(request) != len(results) {
		return nil, errors.New(
			codes.Internal,
			"failed to tokenize card numbers",
			errors.NewErrorInfo(ctx, errcodes.CardTokenizationFailed, "cannot map card numbers with transform response, lengths do not match"),
		)
	}

	response := make(map[string]string)

	for i, data := range results {
		if data.Errors != "" {
			return nil, errors.New(
				codes.Internal,
				"failed to tokenize card numbers",
				errors.NewErrorInfo(ctx, errcodes.CardTokenizationFailed, fmt.Sprintf("error in transform response: %s", data.Errors)),
			)
		}

		if data.EncodedValue != "" {
			response[request[i]] = data.EncodedValue
		} else if data.DecodedValue != "" {
			response[request[i]] = data.DecodedValue
		}
	}

	if len(response) == 0 {
		return nil, errors.New(
			codes.Internal,
			"unable to parse vault request string",
			errors.NewErrorInfo(ctx, errcodes.CardTokenizationFailed, "failed to map vault transform response values"),
		)
	}
	return response, nil
}

package vault_external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"
)

// EncodeResult is part of the return value of a BatchEncode operation
// It is a subset of TransformResult
type EncodeResult struct {
	EncodedValue string `json:"encoded_value"`
	Reference    string `json:"reference"`
}

// DecodeResult is part of the return value of a BatchDecode operation
// It is a subset of TransformResult
type DecodeResult struct {
	DecodedValue string `json:"decoded_value"`
	Reference    string `json:"reference"`
}

// TransformResult contains the data for each transformed value in a batch transformation operation
type TransformResult struct {
	EncodedValue string `json:"encoded_value"`
	DecodedValue string `json:"decoded_value"`
	Reference    string `json:"reference"`
	Errors       string `json:"Errors"`
}

// BatchRequest is the body of a batch transformation operation
type BatchRequest struct {
	Input []*TransformRequest `json:"batch_input"`
}

// BatchTransformResults is a wrapper for the results of a batch transformation
type BatchTransformResults struct {
	Results []*TransformResult `json:"batch_results"`
}

// BatchTransformResponse is a wrapper for BatchTransformResults. It is one specialized case
//  of the generic Vault API responses, and can be extended with error etc. information in the
//  future.
type BatchTransformResponse struct {
	Data *BatchTransformResults `json:"data"`
}

// TransformRequest contains data for an individual value to be transformed in a batch transformation operation.
type TransformRequest struct {
	Value          string `json:"value"`
	Transformation string `json:"transformation"`
	Ttl            int    `json:"ttl"`
	Tweak          []byte `json:"tweak"`
	Reference      string `json:"reference"`
}

// TransformKind models the possible kinds of transformation that Vault supports
type TransformKind int

const (
	TransformEncode TransformKind = iota
	TransformDecode
)

// TransformKind is ultimately used to generate a URL, so provide a nice String() instance
func (t TransformKind) String() string {
	switch t {
	case TransformEncode:
		return "encode"
	case TransformDecode:
		return "decode"
	default:
		return "invalid"
	}
}

// Transform runs a Vault batch transform operation
func (c *client) Transform(ctx context.Context, kind TransformKind, role string, request []*TransformRequest) ([]*TransformResult, error) {
	if !c.auth.isValid() {
		tokenOk := c.auth.awaitValidToken()
		if !tokenOk {
			return nil, errors.New(
				codes.Internal,
				"transform failed",
				errors.NewErrorInfo(ctx, errcodes.Unknown, "waited for valid auth token, but this operation timed out"),
			)
		}
	}

	vaultAuthToken := c.auth.getToken()

	req := &BatchRequest{
		Input: request,
	}

	logf.Debug(ctx, "running transform operation: %s", kind.String())

	path := fmt.Sprintf("v1/int/au/transform/data/%s/%s/%s", c.config.Zone, kind, role)
	requestBytes, err := json.Marshal(req)
	if err != nil {
		// json.Marshal errors can be cryptic, keep the message but add some context
		return nil, errors.Wrap(
			err,
			codes.Internal,
			"transform failed",
			errors.NewErrorInfo(ctx, errcodes.Unknown, fmt.Sprintf("failed to marshal JSON in transform request: %s", err.Error())),
		)
	}

	responseBytes, err := c.api.run(ctx, http.MethodPost, path, vaultAuthToken, requestBytes)
	if err != nil {
		// This error already has useful anzerror info, so use it all the same
		return nil, errors.Wrap(
			err,
			errors.GetStatusCode(err),
			"transform failed",
			errors.GetErrorInfo(err),
		)
	}

	var resp BatchTransformResponse
	err = json.Unmarshal(responseBytes, &resp)
	if err != nil {
		// Wrap this error as json.Unmarshal errors can be cryptic if you aren't sure where they come from
		return nil, errors.Wrap(
			err,
			codes.Internal,
			"transform failed",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, fmt.Sprintf("could not unmarshal JSON from transform response: %s", err.Error())),
		)
	}
	if resp.Data == nil {
		return nil, errors.New(
			codes.Internal,
			"transform failed",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, "transform response JSON missing result data"),
		)
	}
	return resp.Data.Results, nil
}

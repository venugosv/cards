package visa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"google.golang.org/grpc/codes"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
)

type QueryRequest struct {
	PrimaryAccountNumber string `json:"primaryAccountNumber,omitempty"`
}

type TransactionControlListResponses struct {
	// Required - The time the request is received. Value is in UTC time
	ReceivedTimestamp string `json:"receivedTimestamp"`

	// Required - The processing time in milliseconds
	ProcessingTimeInMS int64 `json:"processingTimeinMs"`

	// Required
	Resource ListResource `json:"resource"`
}

type ListResource struct {
	ControlDocuments []Resource `json:"controlDocuments"`
}

// QueryControls retrieves the existing Transaction Control Document
func (c client) QueryControls(ctx context.Context, primaryAccountNumber string) (*Resource, error) {
	if !checkPrimaryAccountNumber(primaryAccountNumber) {
		return nil, anzerrors.New(codes.InvalidArgument, "query failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "cannot parse requested card number"))
	}

	destination := fmt.Sprintf("%s/%s", c.baseURL, queryByPanEndpoint)

	request := QueryRequest{
		PrimaryAccountNumber: primaryAccountNumber,
	}

	requestBody, _ := json.Marshal(request)

	response, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, destination, requestBody), "visa:QueryControls")
	if err != nil {
		return nil, err
	}

	controlDocuments, err := handleTransactionControlListResponses(ctx, response)
	if err != nil {
		return nil, err
	}

	return getTransactionControlDocument(ctx, controlDocuments), nil
}

func getTransactionControlDocument(ctx context.Context, controlDocuments []Resource) *Resource {
	if len(controlDocuments) > 1 {
		logf.Debug(ctx, "multiple control documents returned")
	}

	// TODO: (ElliotMJackson) handle this appropriately (GH-1294)
	return &controlDocuments[0]
}

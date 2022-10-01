package visa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anzx/fabric-cards/pkg/util/apic"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"
)

type StatusResource struct {
	Status string `json:"status"`
}

type AccountUpdateResponse struct {
	// Required - The time the request is received. Value is in UTC time
	ReceivedTimestamp string `json:"receivedTimestamp"`

	// Required - The processing time in milliseconds
	ProcessingTimeinMS int64 `json:"processingTimeinMs"`

	// Required
	Resource StatusResource `json:"resource"`
}

type ReplacementRequest struct {
	CurrentAccountID string
	NewAccountID     string
}

func (c client) ReplaceCard(ctx context.Context, currentAccountID, newAccountID string) (bool, error) {
	if !checkPrimaryAccountNumber(currentAccountID) || !checkPrimaryAccountNumber(newAccountID) {
		return false, anzerrors.New(codes.InvalidArgument, "replace failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "cannot parse requested card number"))
	}

	destination := fmt.Sprintf("%s/%s", c.baseURL, cardReplacementEndpoint)

	request := &ReplacementRequest{
		CurrentAccountID: currentAccountID,
		NewAccountID:     newAccountID,
	}

	requestBody, _ := json.Marshal(request)

	response, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, destination, requestBody), "visa:ReplaceCard")
	if err != nil {
		return false, err
	}

	statusResource, err := handleAccountUpdateResponse(ctx, response)
	if err != nil {
		return false, err
	}

	if statusResource.Status != "SUCCESS" {
		return false, anzerrors.New(codes.Internal, "replace request",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlReplaceFailed, "failed to replace card number"))
	}

	return true, nil
}

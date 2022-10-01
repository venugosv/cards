package visa

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/util/apic"
)

func sendControlRequest(ctx context.Context, apicClient apic.Clienter, method string, destination string, request *Request) (*Resource, error) {
	requestBody, _ := json.Marshal(request)

	response, err := apicClient.Do(ctx, apic.NewRequest(method, destination, requestBody), fmt.Sprintf("visa:%sControls", method))
	if err != nil {
		return nil, err
	}

	return handleTransactionControlDocument(ctx, response)
}

func handleAccountUpdateResponse(ctx context.Context, responseBody []byte) (*StatusResource, error) {
	var response AccountUpdateResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		logf.Error(ctx, err, "visa failed unexpected response from downstream")
		return nil, anzerrors.Wrap(err, codes.Internal, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}

	return &response.Resource, nil
}

func handleTransactionControlDocument(ctx context.Context, responseBody []byte) (*Resource, error) {
	var response TransactionControlDocument
	if err := json.Unmarshal(responseBody, &response); err != nil {
		logf.Error(ctx, err, "visa failed unexpected response from downstream")
		return nil, anzerrors.Wrap(err, codes.Internal, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}

	return &response.Resource, nil
}

func handleTransactionControlListResponses(ctx context.Context, responseBody []byte) ([]Resource, error) {
	var response TransactionControlListResponses
	if err := json.Unmarshal(responseBody, &response); err != nil {
		logf.Error(ctx, err, "visa failed unexpected response from downstream")
		return nil, anzerrors.Wrap(err, codes.Internal, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}
	if len(response.Resource.ControlDocuments) == 0 {
		return nil, anzerrors.New(codes.Internal, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlNoDocumentFound, "no control documents found"))
	}

	return response.Resource.ControlDocuments, nil
}

func checkPrimaryAccountNumber(primaryAccountNumber string) bool {
	return regexp.MustCompile("^[0-9]{16}$").MatchString(primaryAccountNumber)
}

func checkRequest(documentID string, request *Request) bool {
	if documentID == "" || request == nil {
		return false
	}
	return request.GlobalControls != nil || request.MerchantControls != nil || request.TransactionControls != nil
}

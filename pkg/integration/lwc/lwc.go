package lwc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/pkg/errors"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"
)

const (
	failedRequest = "failed retrieve merchants request"
)

// RetrieveMerchants This service retrieves a list of enriched merchant data based on the list of CALs in the request
func (c client) RetrieveMerchants(ctx context.Context, in Request) ([]MerchantDetails, error) {
	logf.Info(ctx, "retrieve merchants request %v", in)

	if len(in.BankTransactions) == 0 {
		err := errors.New("CAL list in request is empty")
		logf.Error(ctx, err, "client:RetrieveMerchants CAL list in request is empty")
		return nil, anzerrors.Wrap(err, codes.InvalidArgument, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "CAL list in request is empty"))
	}

	requestBody, err := json.Marshal(in)
	if err != nil {
		logf.Error(ctx, err, "client:RetrieveMerchants failed to marshall request ")
		return nil, anzerrors.Wrap(err, codes.InvalidArgument, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to marshall request"))
	}

	target := fmt.Sprintf("%s%s", c.baseURL, retrieveMerchantsEndpoint)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewBuffer(requestBody))
	if err != nil {
		logf.Error(ctx, err, "client:RetrieveMerchants error creating request ")
		return nil, anzerrors.Wrap(err, codes.Internal, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "error creating http request"))
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		logf.Error(ctx, err, "client:RetrieveMerchants http request returned an error ")
		return nil, anzerrors.Wrap(err, codes.Internal, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "error making HTTP request to LWC"))
	}
	defer response.Body.Close()

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		logf.Error(ctx, err, "client:RetrieveMerchants: LWC request returned: %v", response.StatusCode)
		return nil, anzerrors.New(apic.CodeFromHTTPStatus(response.StatusCode), failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		logf.Error(ctx, err, "client:RetrieveMerchants unable to read response body")
		return nil, anzerrors.New(apic.CodeFromHTTPStatus(response.StatusCode), failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unable to read response body"))
	}

	var out Response
	if err := json.Unmarshal(responseBody, &out); err != nil {
		logf.Error(ctx, err, "client:RetrieveMerchants failed to unmarshall response")
		return nil, anzerrors.Wrap(err, codes.Internal, failedRequest,
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}

	merchantDetails := extractMerchantsList(out)
	logf.Info(ctx, "successfully retrieved merchant details")
	return merchantDetails, nil
}

func extractMerchantsList(in Response) []MerchantDetails {
	out := make([]MerchantDetails, 0, len(in.SearchResults))
	for _, result := range in.SearchResults {
		for _, merchantDetails := range result.MerchantSearchResults {
			out = append(out, merchantDetails.MerchantDetails)
		}
	}

	return out
}

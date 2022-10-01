package apic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"

	"github.com/anzx/fabric-cards/pkg/rest"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/anzx/pkg/jwtauth/jwtgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"

	"github.com/anzx/fabric-cards/pkg/middleware/requestid"
	"github.com/anzx/pkg/log"
)

const (
	applicationjson = "application/json"
	fabric          = "Fabric"
	ax2             = "AX2"
	branch          = "4111"
)

type Client struct {
	clientID   string
	httpClient *http.Client
}

type Request struct {
	Method      string
	Destination string
	Body        []byte
	Headers     map[string]string
}

func NewRequest(method string, destination string, body []byte) *Request {
	return &Request{Method: method, Destination: destination, Body: body}
}

func (r *Request) addCustomHeaders(request *http.Request) {
	if request.Header == nil {
		request.Header = http.Header{}
	}
	for key, value := range r.Headers {
		request.Header.Set(key, value)
	}
}

type Clienter interface {
	Do(context.Context, *Request, string) ([]byte, error)
}

func NewAPICClient(ctx context.Context, clientIDEnvKey string, httpClient *http.Client, gsmClient *gsm.Client) (*Client, error) {
	if clientIDEnvKey == "" {
		err := anzerrors.New(codes.Internal, "failed to create APIc adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "APIc secret key was empty"))
		logf.Error(ctx, err, "failed to create APIc client")
		return nil, err
	}

	clientID, err := gsmClient.AccessSecret(ctx, clientIDEnvKey)
	if err != nil || clientID == "" {
		log.Error(ctx, err, "unable to load APIc clientID from gsm", log.Str("key", clientIDEnvKey))
		return nil, anzerrors.Wrap(err, codes.Internal, "failed to create APIc adapter",
			anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "unable to find clientID"))
	}

	return &Client{
		clientID:   clientID,
		httpClient: httpClient,
	}, nil
}

func (c *Client) Do(ctx context.Context, r *Request, operation string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, r.Method, r.Destination, bytes.NewBuffer(r.Body))
	if err != nil {
		logf.Error(ctx, err, "APIc client failed to create request")
		// TODO: (ElliotMJackson) evaluate appropriate err code
		return nil, anzerrors.Wrap(err, codes.Internal, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.ContextInvalid, "unable to create request"))
	}

	request = tagClientOperation(request, operation)
	addStandardHeaders(request, c.clientID)
	addTracingHeaders(ctx, request)
	r.addCustomHeaders(request)

	if err := propagateAuthorisationHeader(ctx, request); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		logf.Error(ctx, err, "APIc client failed service unavailable")
		return nil, anzerrors.Wrap(err, codes.Unavailable, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "service unavailable"))
	}
	defer resp.Body.Close()

	var responseBody []byte
	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		logf.Error(ctx, err, "APIc client unable to read body: %v", resp.Body)
		return nil, anzerrors.Wrap(err, codes.DataLoss, "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "failed to read response body"))
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		logf.Error(ctx, err, "APIc client request returned: %v from %v", resp.StatusCode, operation)
		return responseBody, anzerrors.New(CodeFromHTTPStatus(resp.StatusCode), "failed request",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "unexpected response from downstream"))
	}

	return responseBody, nil
}

func propagateAuthorisationHeader(ctx context.Context, request *http.Request) error {
	jwt, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
	if err != nil {
		return anzerrors.New(codes.Unauthenticated, "service unauthenticated",
			anzerrors.NewErrorInfo(ctx, anzcodes.ContextInvalid, "failed to extract auth token"))
	}
	request.Header.Set("authorization", fmt.Sprintf("Bearer %s", jwt))
	return nil
}

func addStandardHeaders(request *http.Request, clientID string) {
	if request.Method != http.MethodGet {
		request.Header.Set("content-Type", applicationjson)
	}
	request.Header.Set("Accept", applicationjson)
	request.Header.Set("anz-application-id", fabric)
	request.Header.Set("X-Originating-App", ax2)
	request.Header.Set("x-ibm-client-id", clientID)
	request.Header.Set("X-Operator-Branch", branch)
}

func addTracingHeaders(ctx context.Context, request *http.Request) {
	request.Header.Set("x-request-id", requestid.FromContext(ctx))
}

func tagClientOperation(request *http.Request, operation string) *http.Request {
	ctx := context.WithValue(request.Context(), rest.ClientTagExtKey{}, rest.ClientTagExt{
		Operation: operation,
	})

	return request.Clone(ctx)
}

func CodeFromHTTPStatus(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusOK:
		return codes.OK
	case http.StatusRequestTimeout:
		return codes.Canceled
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	}

	grpclog.Infof("Unknown httpStatus error gRPC: %v", httpStatus)
	return codes.Internal
}

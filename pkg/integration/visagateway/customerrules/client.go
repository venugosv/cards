package customerrules

import (
	"context"
	"fmt"
	"regexp"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/sanitize"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Client struct {
	CustomerRulesAPIClient crpb.CustomerRulesAPIClient
}

type Config struct {
	BaseURL string `yaml:"baseURL" validate:"required"`
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		CustomerRulesAPIClient: crpb.NewCustomerRulesAPIClient(conn),
	}
}

// Registration of a args or paymentToken in Visa Transaction Controls. If the account is already
// registered then VTC will return the existing documentID and its contents.
func (c Client) Registration(ctx context.Context, primaryAccountNumber string) (string, error) {
	if !checkPrimaryAccountNumber(primaryAccountNumber) {
		logf.Debug(ctx, "invalid VTC registration input")
		return "", anzerrors.New(codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "embedded message failed validation"))
	}

	in := &crpb.RegisterRequest{
		PrimaryAccountNumber: &primaryAccountNumber,
	}

	response, err := c.CustomerRulesAPIClient.Register(ctx, in)
	if err != nil {
		logf.Error(ctx, err, "invalid VTC registration response")
		anzErr, _ := anzerrors.FromStatusError(err)
		return "", anzerrors.Wrap(err, anzerrors.GetStatusCode(anzErr), anzerrors.GetMessage(anzErr),
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "invalid response from visa gateway"))
	}

	if response.GetResource() == nil {
		return "", anzerrors.New(codes.InvalidArgument, "failed to register card number",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "documentID was not returned from Visa"))
	}

	return response.GetResource().DocumentId, nil
}

// ListControlDocuments Lists Transaction Control Documents for a given primary account number (PAN)
func (c Client) ListControlDocuments(ctx context.Context, primaryAccountNumber string) (*crpb.Resource, error) {
	if !checkPrimaryAccountNumber(primaryAccountNumber) {
		return nil, anzerrors.New(codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "cannot parse requested card number"))
	}

	req := &crpb.ListControlDocumentsRequest{
		PrimaryAccountNumber: primaryAccountNumber,
	}

	response, err := c.CustomerRulesAPIClient.ListControlDocuments(ctx, req)
	if err != nil {
		logf.Error(ctx, err, "invalid customer rules response")
		anzErr, _ := anzerrors.FromStatusError(err)
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(anzErr), anzerrors.GetMessage(anzErr),
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "invalid response from visa gateway"))
	}

	resource := response.GetResource()
	if resource == nil {
		return nil, anzerrors.New(codes.NotFound, "failed args",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlNoDocumentFound, "no resource found"))
	}

	controlDocuments := resource.GetControlDocuments()
	if len(controlDocuments) == 0 {
		return nil, anzerrors.New(codes.Internal, "failed args",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlNoDocumentFound, "no control documents found"))
	}

	if len(controlDocuments) > 1 {
		logf.Debug(ctx, "multiple control documents returned")
	}

	return controlDocuments[0], nil
}

// Create Add or update control(s) within the Transaction Control Document
func (c Client) Create(ctx context.Context, documentID string, in *crpb.ControlRequest) (*crpb.Resource, error) {
	request, err := getTransactionControlDocumentRequest(ctx, documentID, in)
	if err != nil {
		return nil, err
	}

	response, err := c.CustomerRulesAPIClient.CreateControls(ctx, request)
	return handleResourceResponse(ctx, err, response)
}

// Update Replace all details of the existing Transaction Control Document
func (c Client) Update(ctx context.Context, documentID string, request *crpb.ControlRequest) (*crpb.Resource, error) {
	in, err := getTransactionControlDocumentRequest(ctx, documentID, request)
	if err != nil {
		return nil, err
	}

	response, err := c.CustomerRulesAPIClient.UpdateControls(ctx, in)
	return handleResourceResponse(ctx, err, response)
}

// Delete existing Control rules within the Transaction Control Document.
func (c Client) Delete(ctx context.Context, documentID string, request *crpb.ControlRequest) (*crpb.Resource, error) {
	in, err := getTransactionControlDocumentRequest(ctx, documentID, request)
	if err != nil {
		return nil, err
	}

	response, err := c.CustomerRulesAPIClient.DeleteControls(ctx, in)
	return handleResourceResponse(ctx, err, response)
}

const (
	success         = "SUCCESS"
	cardNotEnrolled = "CARD_NOT_ENROLLED"
)

// Replace ..
func (c Client) Replace(ctx context.Context, currentAccountID string, newAccountID string) (bool, error) {
	if !checkPrimaryAccountNumber(currentAccountID) || !checkPrimaryAccountNumber(newAccountID) {
		return false, anzerrors.New(codes.InvalidArgument, "replace failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "cannot parse requested card number"))
	}

	request := &crpb.UpdateAccountRequest{
		CurrentAccountId: currentAccountID,
		NewAccountId:     newAccountID,
	}

	response, err := c.CustomerRulesAPIClient.UpdateAccount(ctx, request)
	if err != nil {
		logf.Error(ctx, err, "invalid customer rules response")
		anzErr, _ := anzerrors.FromStatusError(err)
		return false, anzerrors.Wrap(err, anzerrors.GetStatusCode(anzErr), anzerrors.GetMessage(anzErr),
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "invalid response from visa gateway"))
	}

	resource := response.GetResource()
	if resource == nil {
		return false, anzerrors.New(codes.NotFound, "failed args",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlNoDocumentFound, "no resource found"))
	}

	switch resource.GetStatus() {
	case success, cardNotEnrolled:
		message := fmt.Sprintf("controls transfer from %s to %s: %s", currentAccountID, newAccountID, resource.GetStatus())
		logf.Debug(ctx, sanitize.MaskCardNumbersInString(message))
	default:
		return false, anzerrors.New(codes.Internal, "replace args",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlReplaceFailed, "failed to replace card number"))
	}

	return true, nil
}

func getTransactionControlDocumentRequest(ctx context.Context, documentID string, request *crpb.ControlRequest) (*crpb.TransactionControlDocumentRequest, error) {
	if ok := checkRequest(documentID, request); !ok {
		return nil, anzerrors.New(codes.InvalidArgument, "invalid argument",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to make args with provided value"))
	}

	return &crpb.TransactionControlDocumentRequest{
		DocumentId: documentID,
		Request:    request,
	}, nil
}

func checkPrimaryAccountNumber(primaryAccountNumber string) bool {
	return regexp.MustCompile("^[0-9]{16}$").MatchString(primaryAccountNumber)
}

func checkRequest(documentID string, request *crpb.ControlRequest) bool {
	if documentID == "" || request == nil {
		return false
	}
	return request.GlobalControls != nil || request.MerchantControls != nil || request.TransactionControls != nil
}

func handleResourceResponse(ctx context.Context, err error, response *crpb.TransactionControlDocumentResponse) (*crpb.Resource, error) {
	if err != nil {
		logf.Error(ctx, err, "invalid customer rules response")
		anzErr, _ := anzerrors.FromStatusError(err)
		return nil, anzerrors.Wrap(err, anzerrors.GetStatusCode(anzErr), anzerrors.GetMessage(anzErr),
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "invalid response from visa gateway"))
	}

	out := response.GetResource()
	if out == nil {
		return nil, anzerrors.New(codes.NotFound, "failed args",
			anzerrors.NewErrorInfo(ctx, anzcodes.CardControlNoDocumentFound, "no resource found"))
	}

	return out, nil
}

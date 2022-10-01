package customerrules

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/util"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anz-bank/equals"
	"github.com/anzx/fabric-cards/test/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/test/util/bufconn"
	"google.golang.org/grpc"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
)

type MockVTCServer struct {
	crpb.UnimplementedCustomerRulesAPIServer
	registrationFunc func(context.Context, *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error)
	listFunc         func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error)
	getFunc          func(context.Context, *crpb.GetControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error)
	createFunc       func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error)
	updateFunc       func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error)
	deleteFunc       func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error)
	replaceFunc      func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error)
}

func (m MockVTCServer) Register(ctx context.Context, in *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error) {
	return m.registrationFunc(ctx, in)
}

func (m MockVTCServer) ListControlDocuments(ctx context.Context, in *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
	return m.listFunc(ctx, in)
}

func (m MockVTCServer) GetControlDocument(ctx context.Context, in *crpb.GetControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
	return m.getFunc(ctx, in)
}

func (m MockVTCServer) CreateControls(ctx context.Context, in *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
	return m.createFunc(ctx, in)
}

func (m MockVTCServer) UpdateControls(ctx context.Context, in *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
	return m.updateFunc(ctx, in)
}

func (m MockVTCServer) DeleteControls(ctx context.Context, in *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
	return m.deleteFunc(ctx, in)
}

func (m MockVTCServer) UpdateAccount(ctx context.Context, in *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
	return m.replaceFunc(ctx, in)
}

func TestClient_Registration(t *testing.T) {
	tests := []struct {
		name                 string
		primaryAccountNumber string
		registrationFunc     func(context.Context, *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error)
		want                 string
		wantErr              string
	}{
		{
			name:                 "happy path",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			registrationFunc: func(context.Context, *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{
					ReceivedTimestamp:  "",
					ProcessingTimeInMs: 0,
					Resource: &crpb.Resource{
						DocumentId: documentID,
					},
				}, nil
			},
			want: documentID,
		},
		{
			name:                 "unhappy path",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			registrationFunc: func(context.Context, *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{
					ReceivedTimestamp:  "",
					ProcessingTimeInMs: 0,
				}, nil
			},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=failed to register card number, reason=documentID was not returned from Visa",
		},
		{
			name:                 "invalid PAN input",
			primaryAccountNumber: "not a card number",
			wantErr:              "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=embedded message failed validation",
		},
		{
			name:                 "gateway returns fabric error",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			registrationFunc: func(ctx context.Context, _ *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name:                 "gateway returns other error",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			registrationFunc: func(context.Context, *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.registrationFunc == nil {
				test.registrationFunc = func(ctx context.Context, request *crpb.RegisterRequest) (*crpb.TransactionControlDocumentResponse, error) {
					return nil, nil
				}
			}
			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				crpb.RegisterCustomerRulesAPIServer(server, MockVTCServer{
					registrationFunc: test.registrationFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.Registration(context.Background(), test.primaryAccountNumber)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

func TestClient_List(t *testing.T) {
	tests := []struct {
		name                 string
		primaryAccountNumber string
		ListFunc             func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error)
		want                 *crpb.Resource
		wantErr              string
	}{
		{
			name:                 "happy path",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			ListFunc: func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
				return &crpb.TransactionControlList{
					ReceivedTimestamp:  "",
					ProcessingTimeInMs: 0,
					Resource: &crpb.RepeatedResource{
						ControlDocuments: []*crpb.Resource{
							{
								DocumentId: documentID,
							},
						},
					},
				}, nil
			},
			want: &crpb.Resource{DocumentId: documentID},
		},
		{
			name:                 "no resource",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			ListFunc: func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
				return &crpb.TransactionControlList{}, nil
			},
			wantErr: "fabric error: status_code=NotFound, error_code=25001, message=failed args, reason=no resource found",
		},
		{
			name:                 "no control document",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			ListFunc: func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
				return &crpb.TransactionControlList{
					ReceivedTimestamp:  "",
					ProcessingTimeInMs: 0,
					Resource: &crpb.RepeatedResource{
						ControlDocuments: []*crpb.Resource{},
					},
				}, nil
			},
			wantErr: "fabric error: status_code=Internal, error_code=25001, message=failed args, reason=no control documents found",
		},
		{
			name:                 "multiple control documents",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			ListFunc: func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
				return &crpb.TransactionControlList{
					ReceivedTimestamp:  "",
					ProcessingTimeInMs: 0,
					Resource: &crpb.RepeatedResource{
						ControlDocuments: []*crpb.Resource{
							{
								DocumentId: documentID,
							},
							{
								DocumentId: fmt.Sprintf("ANZx%s", documentID),
							},
						},
					},
				}, nil
			},
			want: &crpb.Resource{DocumentId: documentID},
		},
		{
			name:                 "invalid PAN input",
			primaryAccountNumber: "not a card number",
			wantErr:              "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=cannot parse requested card number",
		},
		{
			name:                 "gateway returns fabric error",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			ListFunc: func(ctx context.Context, _ *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name:                 "gateway returns other error",
			primaryAccountNumber: data.AUserWithACard().CardNumber(),
			ListFunc: func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.ListFunc == nil {
				test.ListFunc = func(context.Context, *crpb.ListControlDocumentsRequest) (*crpb.TransactionControlList, error) {
					return nil, nil
				}
			}
			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				crpb.RegisterCustomerRulesAPIServer(server, MockVTCServer{
					listFunc: test.ListFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.ListControlDocuments(context.Background(), test.primaryAccountNumber)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	globalControls := []*crpb.GlobalControl{
		{
			ShouldDeclineAll: util.ToBoolPtr(true),
			IsControlEnabled: true,
			UserIdentifier:   util.ToStringPtr("1234567890"),
		},
	}
	tests := []struct {
		name       string
		documentID string
		request    *crpb.ControlRequest
		CreateFunc func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error)
		want       *crpb.Resource
		wantErr    string
	}{
		{
			name:       "happy path",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			CreateFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{
					Resource: &crpb.Resource{
						GlobalControls: globalControls,
					},
				}, nil
			},
			want: &crpb.Resource{
				GlobalControls: globalControls,
			},
		},
		{
			name:       "no document ID",
			documentID: "",
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "no args",
			documentID: documentID,
			request:    nil,
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "empty args",
			documentID: documentID,
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "no control document returned",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			CreateFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{}, nil
			},
			wantErr: "fabric error: status_code=NotFound, error_code=25001, message=failed args, reason=no resource found",
		},
		{
			name:       "invalid documentID input",
			documentID: "",
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "gateway returns fabric error",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			CreateFunc: func(ctx context.Context, _ *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name:       "gateway returns other error",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			CreateFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.CreateFunc == nil {
				test.CreateFunc = func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
					return nil, nil
				}
			}
			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				crpb.RegisterCustomerRulesAPIServer(server, MockVTCServer{
					createFunc: test.CreateFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.Create(context.Background(), test.documentID, test.request)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	globalControls := []*crpb.GlobalControl{
		{
			ShouldDeclineAll: util.ToBoolPtr(true),
			IsControlEnabled: true,
			UserIdentifier:   util.ToStringPtr("1234567890"),
		},
	}
	tests := []struct {
		name       string
		documentID string
		request    *crpb.ControlRequest
		UpdateFunc func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error)
		want       *crpb.Resource
		wantErr    string
	}{
		{
			name:       "happy path",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			UpdateFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{
					Resource: &crpb.Resource{
						GlobalControls: globalControls,
					},
				}, nil
			},
			want: &crpb.Resource{
				GlobalControls: globalControls,
			},
		},
		{
			name:       "no document ID",
			documentID: "",
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "no args",
			documentID: documentID,
			request:    nil,
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "empty args",
			documentID: documentID,
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "no control document returned",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			UpdateFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{}, nil
			},
			wantErr: "fabric error: status_code=NotFound, error_code=25001, message=failed args, reason=no resource found",
		},
		{
			name:       "invalid documentID input",
			documentID: "",
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "gateway returns fabric error",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			UpdateFunc: func(ctx context.Context, _ *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name:       "gateway returns other error",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			UpdateFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.UpdateFunc == nil {
				test.UpdateFunc = func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
					return nil, nil
				}
			}
			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				crpb.RegisterCustomerRulesAPIServer(server, MockVTCServer{
					updateFunc: test.UpdateFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.Update(context.Background(), test.documentID, test.request)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	globalControls := []*crpb.GlobalControl{
		{
			ShouldDeclineAll: util.ToBoolPtr(true),
			IsControlEnabled: true,
			UserIdentifier:   util.ToStringPtr("1234567890"),
		},
	}
	tests := []struct {
		name       string
		documentID string
		request    *crpb.ControlRequest
		DeleteFunc func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error)
		want       *crpb.Resource
		wantErr    string
	}{
		{
			name:       "happy path",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			DeleteFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{
					Resource: &crpb.Resource{
						GlobalControls: globalControls,
					},
				}, nil
			},
			want: &crpb.Resource{
				GlobalControls: globalControls,
			},
		},
		{
			name:       "no document ID",
			documentID: "",
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "no args",
			documentID: documentID,
			request:    nil,
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "empty args",
			documentID: documentID,
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "no control document returned",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			DeleteFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return &crpb.TransactionControlDocumentResponse{}, nil
			},
			wantErr: "fabric error: status_code=NotFound, error_code=25001, message=failed args, reason=no resource found",
		},
		{
			name:       "invalid documentID input",
			documentID: "",
			request:    &crpb.ControlRequest{},
			wantErr:    "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=unable to make args with provided value",
		},
		{
			name:       "gateway returns fabric error",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			DeleteFunc: func(ctx context.Context, _ *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name:       "gateway returns other error",
			documentID: documentID,
			request: &crpb.ControlRequest{
				GlobalControls: globalControls,
			},
			DeleteFunc: func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.DeleteFunc == nil {
				test.DeleteFunc = func(context.Context, *crpb.TransactionControlDocumentRequest) (*crpb.TransactionControlDocumentResponse, error) {
					return nil, nil
				}
			}
			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				crpb.RegisterCustomerRulesAPIServer(server, MockVTCServer{
					deleteFunc: test.DeleteFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.Delete(context.Background(), test.documentID, test.request)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

func TestClient_Replacement(t *testing.T) {
	type args struct {
		currentAccountID, newAccountID string
	}
	tests := []struct {
		name        string
		args        args
		replaceFunc func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error)
		want        bool
		wantErr     string
	}{
		{
			name: "happy path",
			args: args{
				currentAccountID: "0987654321123456",
				newAccountID:     "1234567890123456",
			},
			replaceFunc: func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
				return &crpb.UpdateAccountResponse{
					Resource: &crpb.UpdateAccountResponse_UpdatedResource{Status: "SUCCESS"},
				}, nil
			},
			want: true,
		},
		{
			name: "happy path, not enrolled",
			args: args{
				currentAccountID: "0987654321123456",
				newAccountID:     "1234567890123456",
			},
			replaceFunc: func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
				return &crpb.UpdateAccountResponse{
					Resource: &crpb.UpdateAccountResponse_UpdatedResource{Status: "CARD_NOT_ENROLLED"},
				}, nil
			},
			want: true,
		},
		{
			name: "unhappy path",
			args: args{
				currentAccountID: "0987654321123456",
				newAccountID:     "1234567890123456",
			},
			replaceFunc: func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
				return &crpb.UpdateAccountResponse{
					Resource: &crpb.UpdateAccountResponse_UpdatedResource{Status: "FAILED"},
				}, nil
			},
			wantErr: "fabric error: status_code=Internal, error_code=25002, message=replace args, reason=failed to replace card number",
		},
		{
			name: "no new account ID",
			args: args{
				currentAccountID: "098765432123456",
				newAccountID:     "",
			},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=replace failed, reason=cannot parse requested card number",
		},
		{
			name: "no current account ID",
			args: args{
				currentAccountID: "",
				newAccountID:     "098765432123456",
			},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=replace failed, reason=cannot parse requested card number",
		},
		{
			name:    "no args",
			args:    args{},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=replace failed, reason=cannot parse requested card number",
		},
		{
			name: "no resource returned",
			args: args{
				currentAccountID: "1234567890123456",
				newAccountID:     "0987654321234567",
			},
			replaceFunc: func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
				return &crpb.UpdateAccountResponse{}, nil
			},
			wantErr: "fabric error: status_code=NotFound, error_code=25001, message=failed args, reason=no resource found",
		},
		{
			name: "gateway returns fabric error",
			args: args{
				currentAccountID: "1234567890123456",
				newAccountID:     "0987654321234567",
			},
			replaceFunc: func(ctx context.Context, _ *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name: "gateway returns other error",
			args: args{
				currentAccountID: "1234567890123456",
				newAccountID:     "0987654321234561",
			},
			replaceFunc: func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.replaceFunc == nil {
				test.replaceFunc = func(context.Context, *crpb.UpdateAccountRequest) (*crpb.UpdateAccountResponse, error) {
					return nil, nil
				}
			}
			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				crpb.RegisterCustomerRulesAPIServer(server, MockVTCServer{
					replaceFunc: test.replaceFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.Replace(context.Background(), test.args.currentAccountID, test.args.newAccountID)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
		crpb.RegisterCustomerRulesAPIServer(server, MockVTCServer{})
	})

	client := NewClient(cc)
	require.NotNil(t, client)
}

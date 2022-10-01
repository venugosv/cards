package dcvv2

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anz-bank/equals"
	"github.com/anzx/fabric-cards/test/data"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/test/util/bufconn"
	"google.golang.org/grpc"

	dcvv2pb "github.com/anzx/fabricapis/pkg/gateway/visa/service/dcvv2"
)

const (
	expirationDate = "2021-09"
	count          = 1
	success        = "00"
	fail           = "01"
)

type mockDCVV2Server struct {
	GenerateFunc func(context.Context, *dcvv2pb.Request) (*dcvv2pb.Response, error)
}

func (m mockDCVV2Server) Generate(ctx context.Context, in *dcvv2pb.Request) (*dcvv2pb.Response, error) {
	return m.GenerateFunc(ctx, in)
}

func TestClient_GenerateList(t *testing.T) {
	type args struct {
		expiryDate, pan string
		count           int
	}
	tests := []struct {
		name         string
		args         args
		generateFunc func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error)
		want         *dcvv2pb.Response
		wantErr      string
	}{
		{
			name: "happy path",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
				count:      count,
			},
			want: &dcvv2pb.Response{
				AccountInfo: &dcvv2pb.AccountInfo{
					PrimaryAccountNumber: &dcvv2pb.PrimaryAccountNumber{
						ExpirationDate: expirationDate,
						Pan:            data.AUserWithACard().CardNumber(),
					},
				},
				MessageIdentification: &dcvv2pb.MessageIdentification{
					TransactionId: uuid.NewString(),
				},
				Dcvv2ItemList: []*dcvv2pb.Dcvv2ItemList{
					{
						Dcvv2Expiry: fmt.Sprintf("%v", time.Time{}.Unix()),
						Dcvv2Value:  fmt.Sprintf("%d", gofakeit.Number(100, 999)),
					},
				},
				TransactionResults: &dcvv2pb.TransactionResults{
					ActionCode: success,
				},
			},
		},
		{
			name: "invalid expiry date input",
			args: args{
				expiryDate: "202109",
				pan:        data.AUserWithACard().CardNumber(),
				count:      count,
			},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=embedded message failed validation",
		},
		{
			name: "invalid pan input",
			args: args{
				expiryDate: expirationDate,
				pan:        "",
				count:      count,
			},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=embedded message failed validation",
		},
		{
			name: "invalid dcvv2 response",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
				count:      1,
			},
			wantErr: "fabric error: status_code=Internal, error_code=2, message=grpc: error while marshaling: proto: Marshal called with nil, reason=invalid response from visa gateway",
		},
		{
			name: "gateway returns fabric error",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
			},
			generateFunc: func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name: "gateway returns other error",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
			},
			generateFunc: func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
		{
			name: "unhappy path",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
				count:      count,
			},
			want: &dcvv2pb.Response{
				AccountInfo: &dcvv2pb.AccountInfo{
					PrimaryAccountNumber: &dcvv2pb.PrimaryAccountNumber{
						ExpirationDate: expirationDate,
						Pan:            data.AUserWithACard().CardNumber(),
					},
				},
				MessageIdentification: &dcvv2pb.MessageIdentification{
					TransactionId: uuid.NewString(),
				},
				TransactionResults: &dcvv2pb.TransactionResults{
					ActionCode: fail,
				},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=2, message=invalid dcvv2 response, reason=invalid action code returned 01",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.generateFunc == nil {
				test.generateFunc = func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
					return test.want, nil
				}
			}

			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				dcvv2pb.RegisterDCVV2APIServer(server, mockDCVV2Server{
					GenerateFunc: test.generateFunc,
				})
			})

			client := NewClient(gofakeit.UUID(), cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.GenerateList(context.Background(), test.args.expiryDate, test.args.pan, test.args.count)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				equals.AssertJson(t, test.want.GetDcvv2ItemList(), got)
			}
		})
	}
}

func TestClient_Generate(t *testing.T) {
	type args struct {
		expiryDate, pan string
	}
	tests := []struct {
		name         string
		args         args
		generateFunc func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error)
		want         *dcvv2pb.Dcvv2ItemList
		wantErr      string
	}{
		{
			name: "happy path",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
			},
			want: &dcvv2pb.Dcvv2ItemList{
				Dcvv2Expiry: fmt.Sprintf("%v", time.Time{}.Unix()),
				Dcvv2Value:  fmt.Sprintf("%d", gofakeit.Number(100, 999)),
			},
		},
		{
			name: "invalid expiry date input",
			args: args{
				expiryDate: "202109",
				pan:        data.AUserWithACard().CardNumber(),
			},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=embedded message failed validation",
		},
		{
			name: "invalid pan input",
			args: args{
				expiryDate: expirationDate,
				pan:        "",
			},
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=embedded message failed validation",
		},
		{
			name: "gateway returns fabric error",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
			},
			generateFunc: func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
				return nil, anzerrors.New(codes.Unavailable, "visa failed",
					anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to call visa"))
			},
			wantErr: "fabric error: status_code=Unavailable, error_code=2, message=visa failed, reason=invalid response from visa gateway",
		},
		{
			name: "gateway returns other error",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
			},
			generateFunc: func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
				return nil, errors.New("oh no")
			},
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=oh no, reason=invalid response from visa gateway",
		},
		{
			name: "unhappy path",
			args: args{
				expiryDate: expirationDate,
				pan:        data.AUserWithACard().CardNumber(),
			},
			generateFunc: func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
				return &dcvv2pb.Response{
					AccountInfo: &dcvv2pb.AccountInfo{
						PrimaryAccountNumber: &dcvv2pb.PrimaryAccountNumber{
							ExpirationDate: expirationDate,
							Pan:            data.AUserWithACard().CardNumber(),
						},
					},
					MessageIdentification: &dcvv2pb.MessageIdentification{
						TransactionId: uuid.NewString(),
					},
					TransactionResults: &dcvv2pb.TransactionResults{
						ActionCode: "01",
					},
				}, nil
			},
			want:    nil,
			wantErr: "fabric error: status_code=NotFound, error_code=2, message=invalid dcvv2 response, reason=invalid action code returned 01",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.generateFunc == nil {
				test.generateFunc = func(ctx context.Context, request *dcvv2pb.Request) (*dcvv2pb.Response, error) {
					return &dcvv2pb.Response{
						AccountInfo: &dcvv2pb.AccountInfo{
							PrimaryAccountNumber: &dcvv2pb.PrimaryAccountNumber{
								ExpirationDate: expirationDate,
								Pan:            data.AUserWithACard().CardNumber(),
							},
						},
						MessageIdentification: &dcvv2pb.MessageIdentification{
							TransactionId: uuid.NewString(),
						},
						Dcvv2ItemList: []*dcvv2pb.Dcvv2ItemList{
							test.want,
						},
						TransactionResults: &dcvv2pb.TransactionResults{
							ActionCode: success,
						},
					}, nil
				}
			}

			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				dcvv2pb.RegisterDCVV2APIServer(server, mockDCVV2Server{
					GenerateFunc: test.generateFunc,
				})
			})

			client := NewClient(gofakeit.UUID(), cc)
			require.NotNil(t, client)
			require.IsType(t, &Client{}, client)

			got, err := client.Generate(context.Background(), test.args.expiryDate, test.args.pan)
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
		dcvv2pb.RegisterDCVV2APIServer(server, mockDCVV2Server{})
	})

	client := NewClient(gofakeit.UUID(), cc)
	require.NotNil(t, client)
}

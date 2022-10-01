package cardonfile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anz-bank/equals"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabricapis/pkg/gateway/visa/service/cardonfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/test/util/bufconn"
	"google.golang.org/grpc"
)

const (
	cardAcceptorID = "103456789123456"
)

type mockCardOnFileServer struct {
	cardonfile.UnimplementedCardOnFileAPIServer
	InquiryFunc func(context.Context, *cardonfile.Request) (*cardonfile.Response, error)
}

func (m mockCardOnFileServer) Inquiry(ctx context.Context, in *cardonfile.Request) (*cardonfile.Response, error) {
	return m.InquiryFunc(ctx, in)
}

func TestNewClient(t *testing.T) {
	cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
		cardonfile.RegisterCardOnFileAPIServer(server, mockCardOnFileServer{})
	})

	client := NewClient(cc)
	require.NotNil(t, client)
}

func TestClient_createRequest(t *testing.T) {
	cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
		cardonfile.RegisterCardOnFileAPIServer(server, mockCardOnFileServer{})
	})

	client := NewClient(cc)
	primaryAccountNumber := []string{data.AUserWithACard().CardNumber()}
	want := &cardonfile.Request{
		Header: &cardonfile.Request_Header{
			RequestMessageId: "",
			MessageDateTime:  time.Now().Format("2006-01-02 15:04:05.000"),
		},
		Data: &cardonfile.Request_Data{
			PrimaryAccountNumbers: []string{data.AUserWithACard().CardNumber()},
			Group:                 group,
		},
	}
	got := client.createRequest(primaryAccountNumber)
	assert.NotNil(t, got)
	assert.Equal(t, want.Data, got.Data)
}

func TestClient_Inquiry(t *testing.T) {
	tests := []struct {
		name        string
		cardNumber  string
		inquiryFunc func(context.Context, *cardonfile.Request) (*cardonfile.Response, error)
		want        []*cardonfile.PANList
		wantErr     string
	}{
		{
			name:       "happy path",
			cardNumber: data.AUserWithACard().CardNumber(),
			inquiryFunc: func(context.Context, *cardonfile.Request) (*cardonfile.Response, error) {
				return &cardonfile.Response{
					Data: &cardonfile.Response_Data{
						PanList: []*cardonfile.PANList{
							{
								PanData: &cardonfile.PANList_Data{
									PanResponseMsg: "Success",
									Pan:            data.AUserWithACard().CardNumber(),
									Merchants: []*cardonfile.Merchants{
										{
											CardAcceptorId: cardAcceptorID,
										},
									},
								},
							},
						},
						Group: group,
					},
				}, nil
			},
			want: []*cardonfile.PANList{
				{
					PanData: &cardonfile.PANList_Data{
						PanResponseMsg: "Success",
						Pan:            data.AUserWithACard().CardNumber(),
						Merchants: []*cardonfile.Merchants{
							{
								CardAcceptorId: cardAcceptorID,
							},
						},
					},
				},
			},
		},
		{
			name:    "invalid PANs input",
			wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=invalid argument, reason=embedded message failed validation",
		},
		{
			name:       "gateway returns error",
			cardNumber: data.AUserWithACard().CardNumber(),
			inquiryFunc: func(context.Context, *cardonfile.Request) (*cardonfile.Response, error) {
				return nil, errors.New("gateway error")
			},
			want:    nil,
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=gateway error, reason=invalid response from visa gateway",
		},
		{
			name:       "unhappy path",
			cardNumber: data.AUserWithACard().CardNumber(),
			inquiryFunc: func(context.Context, *cardonfile.Request) (*cardonfile.Response, error) {
				return &cardonfile.Response{
					Data: &cardonfile.Response_Data{
						Group: group,
					},
				}, nil
			},
			want:    nil,
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=failed to get data from visa response, reason=data was not returned from Visa",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				cardonfile.RegisterCardOnFileAPIServer(server, mockCardOnFileServer{
					InquiryFunc: test.inquiryFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)

			var got []*cardonfile.PANList
			var err error
			if test.cardNumber != "" {
				got, err = client.Inquiry(context.Background(), test.cardNumber)
			} else {
				got, err = client.Inquiry(context.Background())
			}

			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.Nil(t, err)
				equals.AssertJson(t, test.want, got)
			}
		})
	}
}

func TestClient_GetCardAcceptorID(t *testing.T) {
	tests := []struct {
		name        string
		cardNumber  string
		inquiryFunc func(context.Context, *cardonfile.Request) (*cardonfile.Response, error)
		want        []string
		wantErr     string
	}{
		{
			name:       "happy path",
			cardNumber: data.AUserWithACard().CardNumber(),
			want:       []string{cardAcceptorID},
		},
		{
			name:    "invalid PANs input",
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=card number not provided, reason=invalid card number",
		},
		{
			name:       "gateway returns error",
			cardNumber: data.AUserWithACard().CardNumber(),
			inquiryFunc: func(context.Context, *cardonfile.Request) (*cardonfile.Response, error) {
				return nil, errors.New("gateway error")
			},
			want:    nil,
			wantErr: "fabric error: status_code=Unknown, error_code=2, message=gateway error, reason=invalid response from visa gateway",
		},
		{
			name:       "unhappy path",
			cardNumber: data.AUserWithACard().CardNumber(),
			inquiryFunc: func(context.Context, *cardonfile.Request) (*cardonfile.Response, error) {
				return &cardonfile.Response{
					Data: &cardonfile.Response_Data{
						Group: group,
					},
				}, nil
			},
			want:    nil,
			wantErr: "fabric error: status_code=NotFound, error_code=20000, message=failed to get data from visa response, reason=data was not returned from Visa",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.inquiryFunc == nil {
				test.inquiryFunc = func(context.Context, *cardonfile.Request) (*cardonfile.Response, error) {
					return &cardonfile.Response{
						Data: &cardonfile.Response_Data{
							PanList: []*cardonfile.PANList{
								{PanData: &cardonfile.PANList_Data{
									PanResponseMsg: "Success",
									Pan:            data.AUserWithACard().CardNumber(),
									Merchants: []*cardonfile.Merchants{{
										CardAcceptorId: cardAcceptorID,
									}},
								}},
							},
							Group: group,
						},
					}, nil
				}
			}

			cc, _ := bufconn.GetClientConn(t, func(server *grpc.Server) {
				cardonfile.RegisterCardOnFileAPIServer(server, mockCardOnFileServer{
					InquiryFunc: test.inquiryFunc,
				})
			})

			client := NewClient(cc)
			require.NotNil(t, client)

			got, err := client.GetCardAcceptorID(context.Background(), test.cardNumber)
			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

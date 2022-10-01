package selfservice

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/test/util/bufconn"

	"google.golang.org/grpc/credentials/insecure"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
	"github.com/stretchr/testify/require"

	"github.com/anzx/pkg/jwtauth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"gopkg.in/square/go-jose.v2/jwt"
)

type mockSelfServiceServer struct {
	sspb.UnimplementedPartyAPIServer
}

func TestNewSelfServiceClient(t *testing.T) {
	tests := []struct {
		name          string
		input         *Config
		wantErr       string
		listenerClose bool
		serverClose   bool
	}{
		{
			name: "Valid config",
			input: &Config{
				BaseURL: "localhost:9090",
			},
		},
		{
			name: "nil config",
		},
		{
			name:    "invalid config",
			input:   &Config{BaseURL: "%%"},
			wantErr: "fabric error: status_code=Internal, error_code=1, message=failed to create SelfService adapter, reason=unable to parse configured url",
		},
		{
			name: "listener closed",
			input: &Config{
				BaseURL: "localhost:9090",
			},
			listenerClose: true,
			serverClose:   false,
			wantErr:       "fabric error: status_code=Unavailable, error_code=2, message=failed to create SelfService adapter, reason=unable to make successful connection",
		},
		{
			name: "server closed",
			input: &Config{
				BaseURL: "localhost:9090",
			},
			listenerClose: false,
			serverClose:   true,
			wantErr:       "fabric error: status_code=Unavailable, error_code=2, message=failed to create SelfService adapter, reason=unable to make successful connection",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			register := func(server *grpc.Server) {
				sspb.RegisterPartyAPIServer(server, mockSelfServiceServer{})
			}

			listener := bufconn.GetListener(register)
			defer listener.Close()

			if test.listenerClose || test.serverClose {
				listener.Close()
			}

			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					return listener.Dial()
				}),
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			got, err := NewClient(ctx, tt.input, opts...)
			if test.wantErr != "" {
				require.Nil(t, got)
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				if test.input == nil {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}

func TestClient_GetParty(t *testing.T) {
	ctx := context.Background()

	ctxWithClaims := jwtauth.AddClaimsToContext(ctx,
		jwtauth.NewClaims(
			jwtauth.BaseClaims{
				Claims: jwt.Claims{
					Subject: "fake subject UUID",
				},
				OCVID: "OCVID",
			},
		),
	)

	tests := []struct {
		name        string
		context     context.Context
		want        *Party
		err         error
		expectError func(*testing.T, error)
	}{
		{
			name:    "basic happy",
			context: ctxWithClaims,
			want: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					LegalName: &sspb.Name{
						Name:       "Ms. Oprah Gail Winfrey",
						Prefix:     "Queen",
						Title:      "Ms",
						FirstName:  "Oprah",
						MiddleName: "Gail",
						LastName:   "Winfrey",
						Suffix:     "",
					},
					ResidentialAddress: &sspb.Address{
						LineOne:    "Level 13",
						LineTwo:    "839 Collins Street",
						City:       "Docklands",
						PostalCode: "3008",
						State:      "VIC",
						Country:    "AU",
					},
					MailingAddress: &sspb.Address{
						LineOne:    "Mailroom",
						LineTwo:    "833 Collins Street",
						City:       "Docklands",
						PostalCode: "3008",
						State:      "VIC",
						Country:    "AU",
					},
				},
			},
		},
		{
			name:    "unable to dial self service",
			context: ctxWithClaims,
			err: anzerrors.New(codes.NotFound, "SelfService failed",
				anzerrors.NewErrorInfo(ctx, anzcodes.ContextInvalid, "service unavailable")),
			expectError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "fabric error: status_code=NotFound, error_code=2, message=SelfService failed, reason=service unavailable")
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ss := &mockSelfServicePartyAPIClient{}

			ss.getPartyCall = func(ctx context.Context, in *sspb.GetPartyRequest, opts ...grpc.CallOption) (*sspb.GetPartyResponse, error) {
				if test.want != nil {
					return test.want.GetPartyResponse, test.err
				}
				return nil, test.err
			}

			c := &Client{PartyAPIClient: ss}
			got, err := c.GetParty(ctx)
			if test.expectError != nil {
				if err == nil {
					t.Fatal("Expected error, got none")
				} else {
					test.expectError(t, err)
				}
			} else {
				if err != nil {
					t.Fatal(err.Error())
				}
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

type mockSelfServicePartyAPIClient struct {
	getPartyCall    func(ctx context.Context, in *sspb.GetPartyRequest, opts ...grpc.CallOption) (*sspb.GetPartyResponse, error)
	updatePartyCall func(ctx context.Context, in *sspb.UpdatePartyRequest, opts ...grpc.CallOption) (*sspb.UpdatePartyResponse, error)
}

func (m *mockSelfServicePartyAPIClient) GetParty(ctx context.Context, in *sspb.GetPartyRequest, opts ...grpc.CallOption) (*sspb.GetPartyResponse, error) {
	return m.getPartyCall(ctx, in, opts...)
}

func (m *mockSelfServicePartyAPIClient) UpdateParty(ctx context.Context, in *sspb.UpdatePartyRequest, opts ...grpc.CallOption) (*sspb.UpdatePartyResponse, error) {
	return m.updatePartyCall(ctx, in, opts...)
}

func TestParty_GetAddress(t *testing.T) {
	tests := []struct {
		name    string
		party   *Party
		want    *sspb.Address
		wantErr string
	}{
		{
			name: "GetMailingAddress",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					ResidentialAddress: &sspb.Address{
						LineOne:    "ResidentialAddress_LineOne",
						LineTwo:    "ResidentialAddress_LineTwo",
						LineThree:  "ResidentialAddress_LineThree",
						LineFour:   "ResidentialAddress_LineFour",
						LineFive:   "ResidentialAddress_LineFive",
						LineSix:    "ResidentialAddress_LineSix",
						City:       "ResidentialAddress_City",
						PostalCode: "ResidentialAddress_PostalCode",
						State:      "ResidentialAddress_State",
						Country:    "ResidentialAddress_Country",
					},
					MailingAddress: nil,
				},
			},
			want: &sspb.Address{
				LineOne:    "ResidentialAddress_LineOne",
				LineTwo:    "ResidentialAddress_LineTwo",
				LineThree:  "ResidentialAddress_LineThree",
				LineFour:   "ResidentialAddress_LineFour",
				LineFive:   "ResidentialAddress_LineFive",
				LineSix:    "ResidentialAddress_LineSix",
				City:       "ResidentialAddress_City",
				PostalCode: "ResidentialAddress_PostalCode",
				State:      "ResidentialAddress_State",
				Country:    "ResidentialAddress_Country",
			},
		}, {
			name: "NoMailingAddressGetResidentialAddress",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					ResidentialAddress: nil,
					MailingAddress: &sspb.Address{
						LineOne:    "MailingAddress_LineOne",
						LineTwo:    "MailingAddress_LineTwo",
						LineThree:  "MailingAddress_LineThree",
						LineFour:   "MailingAddress_LineFour",
						LineFive:   "MailingAddress_LineFive",
						LineSix:    "MailingAddress_LineSix",
						City:       "MailingAddress_City",
						PostalCode: "MailingAddress_PostalCode",
						State:      "MailingAddress_State",
						Country:    "MailingAddress_Country",
					},
				},
			},
			want: &sspb.Address{
				LineOne:    "MailingAddress_LineOne",
				LineTwo:    "MailingAddress_LineTwo",
				LineThree:  "MailingAddress_LineThree",
				LineFour:   "MailingAddress_LineFour",
				LineFive:   "MailingAddress_LineFive",
				LineSix:    "MailingAddress_LineSix",
				City:       "MailingAddress_City",
				PostalCode: "MailingAddress_PostalCode",
				State:      "MailingAddress_State",
				Country:    "MailingAddress_Country",
			},
		}, {
			name: "NoAddress",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					ResidentialAddress: nil,
					MailingAddress:     nil,
				},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=20003, message=Party Incomplete, reason=address not found",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got, err := test.party.GetAddress(context.Background())
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestParty_GetName(t *testing.T) {
	tests := []struct {
		name    string
		party   *Party
		want    string
		wantErr string
	}{
		{
			name: "successfully with full name",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					LegalName: &sspb.Name{
						FirstName:  "First",
						MiddleName: "Middle",
						LastName:   "Last",
					},
				},
			},
			want: "First Middle Last",
		},
		{
			name: "successfully without first name",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					LegalName: &sspb.Name{
						MiddleName: "Middle",
						LastName:   "Last",
					},
				},
			},
			want: "Middle Last",
		},
		{
			name: "successfully without middle name",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					LegalName: &sspb.Name{
						FirstName: "First",
						LastName:  "Last",
					},
				},
			},
			want: "First Last",
		},
		{
			name: "successfully without last name",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					LegalName: &sspb.Name{
						FirstName:  "First",
						MiddleName: "Middle",
					},
				},
			},
			want: "First Middle",
		},
		{
			name: "successfully with only last name",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					LegalName: &sspb.Name{
						LastName: "Last",
					},
				},
			},
			want: "Last",
		},
		{
			name: "return empty name with error",
			party: &Party{
				GetPartyResponse: &sspb.GetPartyResponse{
					LegalName: &sspb.Name{},
				},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=2, message=Party Incomplete, reason=empty name returned",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got, err := test.party.GetName(context.Background())
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, got, test.want)
			}
		})
	}
}

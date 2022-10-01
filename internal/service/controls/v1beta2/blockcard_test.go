package v1beta2

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/anzx/pkg/auditlog"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"

	"github.com/anzx/fabric-cards/test/util"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"

	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
)

func TestBlockAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.BlockCard{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}

		builder := fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))).
			WithAuditLogHook(hook)
		request := &ccpb.BlockCardRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
			Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardControlsServer(builder)

		_, _ = s.BlockCard(ctx, request)

		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
	})
}

func TestBlockCard(t *testing.T) {
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     *ccpb.BlockCardRequest
		want    *ccpb.BlockCardResponse
		wantErr error
	}{
		{
			name:    "Successfully apply block",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_BLOCK,
			},
			want: &ccpb.BlockCardResponse{
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
					epb.Eligibility_ELIGIBILITY_BLOCK,
					epb.Eligibility_ELIGIBILITY_GET_DETAILS,
					epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
				},
			},
		},
		{
			name:    "Successfully remove block",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
			},
			want: &ccpb.BlockCardResponse{
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
					epb.Eligibility_ELIGIBILITY_BLOCK,
					epb.Eligibility_ELIGIBILITY_GET_DETAILS,
					epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
				},
			},
		},
		{
			name:    "Successfully remove T block",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
			},
			want: &ccpb.BlockCardResponse{
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
					epb.Eligibility_ELIGIBILITY_BLOCK,
					epb.Eligibility_ELIGIBILITY_GET_DETAILS,
					epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
				},
			},
		},
		{
			name: "audit log does not affect response",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))).
				WithAuditLogError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
			},
			want: &ccpb.BlockCardResponse{
				Eligibilities: []epb.Eligibility{
					epb.Eligibility_ELIGIBILITY_APPLE_PAY,
					epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
					epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
					epb.Eligibility_ELIGIBILITY_CHANGE_PIN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
					epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
					epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
					epb.Eligibility_ELIGIBILITY_BLOCK,
					epb.Eligibility_ELIGIBILITY_GET_DETAILS,
					epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
				},
			},
		},
		{
			name:    "Entitlements call returns not entitled",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_BLOCK,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=block failed, reason=service unavailable"),
		},
		{
			name:    "Eligibility call returns not eligible",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_BLOCK,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=block failed, reason=card not eligible"),
		},
		{
			name:    "Eligibility call returns not eligible",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEligibilityError(),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=unblock failed, reason=card not eligible"),
		},
		{
			name: "CTM call fails to update status",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))).
				WithCtmUpdateError(anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=unblock failed, reason=service unavailable"),
		},
		{
			name: "CTM inquiry call failed",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))).
				WithCtmInquiryError(anzerrors.New(codes.Unavailable, "failed request",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
			},
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=unblock failed, reason=service unavailable"),
		},
		{
			name:    "eligibility failed, unblock return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusTemporaryBlock))).WithEligibilityError(),
			req: &ccpb.BlockCardRequest{
				TokenizedCardNumber: data.AUserWithACard().Token(),
				Action:              ccpb.BlockCardRequest_ACTION_UNBLOCK,
			},
			wantErr: errors.New("fabric error: status_code=PermissionDenied, error_code=20002, message=unblock failed, reason=card not eligible"),
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			err := feature.FeatureGate.Set(map[feature.Feature]bool{
				feature.GCT_GLOBAL: true,
			})
			require.NoError(t, err)
			ctx, b := fixtures.GetTestContextWithLogger(nil)
			s := buildCardControlsServer(test.builder)
			got, err := s.BlockCard(ctx, test.req)
			util.CheckTestAndAuditLogs(t, got, test.want, test.wantErr, err, b)
		})
	}
}

func buildCardControlsServer(c *fixtures.ServerBuilder) ccpb.CardControlsAPIServer {
	fabric := Fabric{
		CommandCentre: &commandcentre.Client{
			Publisher: c.CommandCentreEnv,
		},
		Eligibility: &eligibility.Client{
			CardEligibilityAPIClient: c.CardEligibilityAPIClient,
		},
		Entitlements: entitlements.Client{
			CardEntitlementsAPIClient: c.CardEntitlementsAPIClient,
		},
		Visa: &customerrules.Client{
			CustomerRulesAPIClient: c.CustomerRulesClient,
		},
	}
	internal := Internal{}
	external := External{
		Vault: c.VaultClient,
		CTM:   c.CTMClient,
		AuditLog: &auditlogger.Client{
			Publisher: c.AuditLogPublisher,
		},
		OCV:       c.OCVClient,
		Forgerock: c.ForgerockClient,
	}
	return NewServer(fabric, internal, external)
}

func Test_actionAuditLog(t *testing.T) {
	tests := map[ccpb.BlockCardRequest_Action]auditlog.Event{
		ccpb.BlockCardRequest_ACTION_BLOCK:               auditlog.EventBlockCard,
		ccpb.BlockCardRequest_ACTION_UNBLOCK:             auditlog.EventUnblockCard,
		ccpb.BlockCardRequest_ACTION_UNKNOWN_UNSPECIFIED: "",
	}
	for in, want := range tests {
		t.Run(fmt.Sprintf("%s=%s", in, want), func(t *testing.T) {
			assert.Equal(t, want, actionAuditLog(in))
		})
	}
}

func Test_actionCardStatus(t *testing.T) {
	tests := map[ccpb.BlockCardRequest_Action]ctm.Status{
		ccpb.BlockCardRequest_ACTION_BLOCK:               ctm.StatusTemporaryBlock,
		ccpb.BlockCardRequest_ACTION_UNBLOCK:             ctm.StatusIssued,
		ccpb.BlockCardRequest_ACTION_UNKNOWN_UNSPECIFIED: "",
	}
	for in, want := range tests {
		t.Run(fmt.Sprintf("%s=%s", in, want), func(t *testing.T) {
			assert.Equal(t, want, actionCardStatus(in))
		})
	}
}

func Test_actionEligibility(t *testing.T) {
	tests := map[ccpb.BlockCardRequest_Action]epb.Eligibility{
		ccpb.BlockCardRequest_ACTION_BLOCK:               epb.Eligibility_ELIGIBILITY_BLOCK,
		ccpb.BlockCardRequest_ACTION_UNBLOCK:             epb.Eligibility_ELIGIBILITY_UNBLOCK,
		ccpb.BlockCardRequest_ACTION_UNKNOWN_UNSPECIFIED: epb.Eligibility_ELIGIBILITY_INVALID_UNSPECIFIED,
	}
	for in, want := range tests {
		t.Run(fmt.Sprintf("%s=%s", in, want), func(t *testing.T) {
			assert.Equal(t, want, actionEligibility(in))
		})
	}
}

func Test_getFailMessage(t *testing.T) {
	tests := map[ccpb.BlockCardRequest_Action]string{
		ccpb.BlockCardRequest_ACTION_BLOCK:               "block failed",
		ccpb.BlockCardRequest_ACTION_UNBLOCK:             "unblock failed",
		ccpb.BlockCardRequest_ACTION_UNKNOWN_UNSPECIFIED: "block/unblock failed",
	}
	for in, want := range tests {
		t.Run(fmt.Sprintf("%s=%s", in, want), func(t *testing.T) {
			assert.Equal(t, want, getFailMessage(in))
		})
	}
}

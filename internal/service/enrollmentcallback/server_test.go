package enrollmentcallback

import (
	"context"
	"testing"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/integration/fakerock"

	"github.com/stretchr/testify/require"

	ecpb "github.com/anzx/fabricapis/pkg/visa/service/enrollmentcallback"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/fixtures"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

var tests = []struct {
	name           string
	featureToggles map[feature.Feature]bool
	builder        *fixtures.ServerBuilder
	req            *ecpb.Request
	want           *ecpb.Response
	wantErr        string
}{
	{
		name: "Successfully set customerRulesAPI flag",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: true,
			feature.FORGEROCK_SYSTEM_LOGIN:         false,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
				{
					PrimaryAccountNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		want: &ecpb.Response{},
	},
	{
		name: "empty BulkEnrollmentObjectList fails",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: true,
			feature.FORGEROCK_SYSTEM_LOGIN:         false,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{},
		},
		wantErr: "fabric error: status_code=InvalidArgument, error_code=4, message=callback failed, reason=empty BulkEnrollmentObjectList provided",
	},
	{
		name: "error set customerRulesAPI flag",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: true,
			feature.FORGEROCK_SYSTEM_LOGIN:         false,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmSetPreferenceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
				{
					PrimaryAccountNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		wantErr: "fabric error: status_code=Unavailable, error_code=2, message=callback failed, reason=service unavailable",
	},
	{
		name: "vault fails",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: true,
			feature.FORGEROCK_SYSTEM_LOGIN:         false,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
				{
					PrimaryAccountNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		wantErr: "fabric error: status_code=Internal, error_code=2, message=callback failed, reason=service unavailable",
	},
	{
		name: "fakerock fails",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: true,
			feature.FORGEROCK_SYSTEM_LOGIN:         false,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()).WithFakerockError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
				{
					PrimaryAccountNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		wantErr: "fabric error: status_code=Internal, error_code=2, message=callback failed, reason=system login failed",
	},
	{
		name: "forgerock enabled",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: true,
			feature.FORGEROCK_SYSTEM_LOGIN:         true,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
				{
					PrimaryAccountNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		want: &ecpb.Response{},
	},
	{
		name: "forgerock fails",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: true,
			feature.FORGEROCK_SYSTEM_LOGIN:         true,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()).WithForgerockError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
				{
					PrimaryAccountNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		wantErr: "fabric error: status_code=Unavailable, error_code=2, message=callback failed, reason=service unavailable",
	},
	{
		name: "not integrated, returns 200 OK",
		featureToggles: map[feature.Feature]bool{
			feature.ENROLLMENT_CALLBACK_INTEGRATED: false,
		},
		builder: fixtures.AServer().WithData(data.AUserWithACard()).WithForgerockError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
		req: &ecpb.Request{
			BulkEnrollmentObjectList: []*ecpb.BulkEnrollmentObjectList{
				{
					PrimaryAccountNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		want: &ecpb.Response{},
	},
}

func TestServer_Enroll(t *testing.T) {
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := &server{
				ctm:       test.builder.CTMClient,
				vault:     test.builder.VaultClient,
				forgerock: test.builder.ForgerockClient,
				fakerock: &fakerock.Client{
					FakerockAPIClient: test.builder.FakerockClient,
				},
			}
			err := feature.FeatureGate.Set(test.featureToggles)
			require.NoError(t, err)
			got, err := s.Enroll(fixtures.GetTestContext(), test.req)
			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestServer_Disenroll(t *testing.T) {
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := &server{
				ctm:       test.builder.CTMClient,
				vault:     test.builder.VaultClient,
				forgerock: test.builder.ForgerockClient,
				fakerock: &fakerock.Client{
					FakerockAPIClient: test.builder.FakerockClient,
				},
			}
			err := feature.FeatureGate.Set(test.featureToggles)
			require.NoError(t, err)
			got, err := s.Disenroll(fixtures.GetTestContext(), test.req)
			if test.wantErr != "" {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	c := fixtures.AServer().WithData(data.AUserWithACard())
	cTMClient := c.CTMClient
	vaultClient := c.VaultClient
	fakerockClient := &fakerock.Client{
		FakerockAPIClient: c.FakerockClient,
	}
	forgerockClient := c.ForgerockClient

	got := NewServer(cTMClient, vaultClient, fakerockClient, forgerockClient)
	assert.NotNil(t, got)
	assert.IsType(t, &server{}, got)
}

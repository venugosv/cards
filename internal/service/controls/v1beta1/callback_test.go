package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/test/data"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/assert"

	"github.com/anzx/fabric-cards/test/fixtures"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

var tests = []struct {
	name    string
	builder *fixtures.ServerBuilder
	req     *ccpb.CallbackRequest
	want    *ccpb.CallbackResponse
	wantErr error
}{
	{
		name:    "Successfully set customerRulesAPI flag",
		builder: fixtures.AServer().WithData(data.AUserWithACard()),
		req: &ccpb.CallbackRequest{
			BulkEnrollmentObjectList: []*ccpb.BulkEnrollmentObject{
				{
					CardNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		want: &ccpb.CallbackResponse{},
	},
	{
		name:    "empty BulkEnrollmentObjectList fails",
		builder: fixtures.AServer().WithData(data.AUserWithACard()),
		req: &ccpb.CallbackRequest{
			BulkEnrollmentObjectList: nil,
		},
		wantErr: errors.New("fabric error: status_code=InvalidArgument, error_code=4, message=callback failed, reason=empty BulkEnrollmentObjectList provided"),
	},
	{
		name:    "error set customerRulesAPI flag",
		builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmSetPreferenceError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
		req: &ccpb.CallbackRequest{
			BulkEnrollmentObjectList: []*ccpb.BulkEnrollmentObject{
				{
					CardNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=callback failed, reason=service unavailable"),
	},
	{
		name:    "vault fails",
		builder: fixtures.AServer().WithData(data.AUserWithACard()).WithVaultError(anzerrors.New(codes.Unavailable, "failed request", anzerrors.NewErrorInfo(context.Background(), anzcodes.DownstreamFailure, "service unavailable"))),
		req: &ccpb.CallbackRequest{
			BulkEnrollmentObjectList: []*ccpb.BulkEnrollmentObject{
				{
					CardNumber: data.AUserWithACard().CardNumber(),
				},
			},
		},
		wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=callback failed, reason=service unavailable"),
	},
}

func TestFlagEnrol(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildCardControlsServer(tt.builder)
			got, err := s.Enrol(fixtures.GetTestContext(), tt.req)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFlagDisenrol(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildCardControlsServer(tt.builder)
			got, err := s.Disenrol(fixtures.GetTestContext(), tt.req)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

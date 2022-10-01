package eligibility

import (
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/test/fixtures"
	"github.com/pkg/errors"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestCan(t *testing.T) {
	t.Parallel()
	type req struct {
		tokenizedCardNumber string
		eligibility         epb.Eligibility
	}
	tests := []struct {
		name    string
		builder *fixtures.ServerBuilder
		req     req
		want    *epb.CanResponse
		wantErr bool
	}{
		{
			name:    "entitlements fails may return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEntMayError(errors.New("oh no")),
			req: req{
				tokenizedCardNumber: data.AUserWithACard().Token(),
				eligibility:         epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			},
			want:    &epb.CanResponse{},
			wantErr: true,
		},
		{
			name:    "ctm fails may return false",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithCtmInquiryError(errors.New("oh no")),
			req: req{
				tokenizedCardNumber: data.AUserWithACard().Token(),
				eligibility:         epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			},
			want:    &epb.CanResponse{},
			wantErr: true,
		},
		{
			name:    "ctm succeeds may return eligible true",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			req: req{
				tokenizedCardNumber: data.AUserWithACard().Token(),
				eligibility:         epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			},
			want:    &epb.CanResponse{},
			wantErr: false,
		},
		{
			name:    "ctm succeeds may return eligible false",
			builder: fixtures.AServer().WithData(data.AUserWithACard(data.WithStatus(ctm.StatusStolen))),
			req: req{
				tokenizedCardNumber: data.AUserWithACard().Token(),
				eligibility:         epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := BuildEligibilityServer(tt.builder)
			got, err := s.Can(fixtures.GetTestContext(), &epb.CanRequest{
				TokenizedCardNumber: tt.req.tokenizedCardNumber,
				Eligibility:         tt.req.eligibility,
			})
			if test.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func BuildEligibilityServer(c *fixtures.ServerBuilder) epb.CardEligibilityAPIServer {
	return NewServer(
		entitlements.Client{
			CardEntitlementsAPIClient: c.CardEntitlementsAPIClient,
		},
		c.CTMClient,
		c.VaultClient,
	)
}

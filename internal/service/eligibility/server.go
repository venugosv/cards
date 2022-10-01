package eligibility

import (
	"context"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/integration/vault"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
)

type server struct {
	epb.UnimplementedCardEligibilityAPIServer
	entitlements entitlements.Carder
	ctm          ctm.CardInquiryAPI
	vault        vault.Client
}

// NewServer constructs a new CardEntitlementsAPIClient from configured clients
func NewServer(entitlementsQueryAPIClient entitlements.Carder,
	cardInquiryAPI ctm.CardInquiryAPI, vaultClient vault.Client,
) *server {
	return &server{
		entitlements: entitlementsQueryAPIClient,
		ctm:          cardInquiryAPI,
		vault:        vaultClient,
	}
}

func (s server) Can(ctx context.Context, req *epb.CanRequest) (*epb.CanResponse, error) {
	if _, err := s.entitlements.GetEntitledCard(ctx, req.TokenizedCardNumber, entitlements.OPERATION_VIEW_CARD); err != nil {
		return nil, err
	}

	cardDetails, err := s.ctm.DebitCardInquiry(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, err
	}

	if eligible := cardDetails.HasEligibility(req.Eligibility); !eligible {
		return nil, anzerrors.New(codes.InvalidArgument, "eligibility failed", anzerrors.NewErrorInfo(ctx, anzcodes.CardIneligible, "card not eligible"))
	}
	return &epb.CanResponse{}, nil
}

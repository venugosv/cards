package cards

import (
	"context"
	"fmt"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	anzerrors "github.com/anzx/pkg/errors"
)

func (s server) getCard(ctx context.Context, tokenizedCardNumber string, service string) (*ctm.DebitCardResponse, error) {
	errMsg := fmt.Sprintf("%s failed", service)
	if _, err := s.Entitlements.GetEntitledCard(ctx, tokenizedCardNumber, entitlements.OPERATION_VIEW_CARD); err != nil {
		return nil, serviceErr(err, errMsg)
	}

	card, err := s.CTM.DebitCardInquiry(ctx, tokenizedCardNumber)
	if err != nil {
		return nil, serviceErr(err, errMsg)
	}

	return card, nil
}

func serviceErr(err error, msg string) error {
	return anzerrors.Wrap(err, anzerrors.GetStatusCode(err), msg, anzerrors.GetErrorInfo(err))
}

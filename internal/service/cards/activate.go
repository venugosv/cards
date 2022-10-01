package cards

import (
	"context"
	"regexp"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/pkg/ratelimit"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	"github.com/anzx/pkg/auditlog"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

const (
	activationFailed   = "activation failed"
	serviceUnavailable = "service unavailable"
)

func (s server) Activate(ctx context.Context, req *cpb.ActivateRequest) (retResponse *cpb.ActivateResponse, retError error) {
	serviceData := &servicedata.ActivateCard{
		TokenizedCardNumber: req.TokenizedCardNumber,
		Last_4Digits:        req.Last_6Digits[2:],
		CustomerClass:       "CNE",
	}

	defer func() {
		if err := serviceData.Validate(); err != nil {
			logf.Error(ctx, err, "invalid service data payload")
		}
		s.AuditLog.Publish(ctx, auditlog.EventActivateCard, retResponse, retError, serviceData)
	}()

	if err := s.RateLimit.Allow(ctx, ratelimit.Activate); err != nil {
		return nil, serviceErr(err, activationFailed)
	}

	cardNumber, err := s.Vault.DecodeCardNumber(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, anzerrors.New(codes.Internal, activationFailed, anzerrors.NewErrorInfo(ctx, anzcodes.CardTokenizationFailed, "service unavailable"))
	}

	if !last6DigitsMatch(cardNumber, req.Last_6Digits) {
		return nil, anzerrors.New(codes.InvalidArgument, activationFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardLastSixDigitsDontMatch, "last 6 Digits do not match"))
	}

	entitledCard, err := s.Entitlements.GetEntitledCard(ctx, req.TokenizedCardNumber, entitlements.OPERATION_MANAGE_CARD)
	if err != nil {
		return nil, serviceErr(err, activationFailed)
	}

	serviceData.AccountNumbers = entitledCard.GetAccountNumbers()

	card, err := s.CTM.DebitCardInquiry(ctx, req.TokenizedCardNumber)
	if err != nil {
		return nil, anzerrors.Wrap(err, codes.Internal, activationFailed, anzerrors.NewErrorInfo(ctx, anzcodes.CardNotFound, serviceUnavailable))
	}

	if card.Status != ctm.StatusIssued {
		return nil, anzerrors.Wrap(err, codes.Internal, activationFailed, anzerrors.NewErrorInfo(ctx, anzcodes.CardIneligible, "card ineligible"))
	}

	isActivated := false
	// If we are not eligible for activation, we can interpret that to mean our card is already activated
	if err := s.Eligibility.Can(ctx, epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION, req.TokenizedCardNumber); err != nil {
		isActivated = true
	}

	// If our card is not activated, we call Activate in CTM
	if !isActivated {
		if _, err := s.CTM.Activate(ctx, req.TokenizedCardNumber); err != nil {
			return nil, serviceErr(err, activationFailed)
		}
		s.CommandCentre.PublishEventAsync(ctx, event.CardStatusChange)
	}

	card, err = s.CTM.DebitCardInquiry(ctx, req.TokenizedCardNumber)
	if err != nil {
		return &cpb.ActivateResponse{}, nil
	}

	serviceData.IssueReason = card.IssueReason
	serviceData.Last_4Digits = card.CardNumber.Last4Digits

	return &cpb.ActivateResponse{
		Eligibilities: card.Eligibility(),
	}, nil
}

func last6DigitsMatch(cardNumber string, last6Digits string) bool {
	return regexp.MustCompile("^[0-9]{16}$").MatchString(cardNumber) && (cardNumber[len(cardNumber)-6:] == last6Digits)
}

package cards

import (
	"context"
	"strconv"

	"github.com/anzx/fabric-cards/pkg/date"

	pbtype "github.com/anzx/fabricapis/pkg/fabric/type"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

func (s server) AuditTrail(ctx context.Context, req *cpb.AuditTrailRequest) (*cpb.AuditTrailResponse, error) {
	card, err := s.getCard(ctx, req.TokenizedCardNumber, "audit trail")
	if err != nil {
		return nil, err
	}

	return &cpb.AuditTrailResponse{
		AccountsLinked:        card.AccountsLinkedCount,
		TotalCards:            card.TotalCards,
		Activated:             card.ActivationStatus,
		CardControlEnabled:    card.CardControlPreference,
		MerchantUpdateEnabled: card.MerchantUpdatePreference,
		ReplacedDate:          date.GetDate(ctx, date.YYYYMMDD, card.ReplacedDate),
		ReplacementCount:      card.ReplacementCount,
		IssueDate:             date.GetDate(ctx, date.YYYYMMDD, card.IssueDate),
		ReissueDate:           date.GetDate(ctx, date.YYYYMMDD, card.ReissueDate),
		ExpiryDate:            date.GetDate(ctx, date.YYMM, card.ExpiryDate),
		PreviousExpiryDate:    date.GetDate(ctx, date.YYYYMM, card.PrevExpiryDate),
		DetailsChangedDate:    date.GetDate(ctx, date.YYYYMMDD, card.DetailsChangedDate),
		ClosedDate:            GetTimestampPtr(ctx, date.YYYYMMDD, card.ClosedDate),
		Limits:                getLimits(ctx, card.Limits),
		NewCard:               getMaskedCard(card.NewCardNumber),
		OldCard:               getMaskedCard(card.OldCardNumber),
		PinChangeDate:         date.GetDate(ctx, date.YYYYMMDD, card.PinChangeDate),
		PinChangeCount:        card.PinChangedCount,
		LastPinFailed:         date.GetDate(ctx, date.YYYYMMDD, card.LastPinFailed),
		PinFailedCount:        card.PinFailedCount,
		Status:                card.Status.String(),
		StatusChangedDate:     date.GetDate(ctx, date.YYYYMMDD, card.StatusChangedDate),
	}, nil
}

func getMaskedCard(card *ctm.Card) *cpb.MaskedCard {
	if card == nil {
		return nil
	}

	return &cpb.MaskedCard{
		TokenizedCardNumber: card.Token,
		Last_4Digits:        card.Last4Digits,
	}
}

func getLimits(ctx context.Context, limits []ctm.NewLimits) []*cpb.Limit {
	var result []*cpb.Limit
	for _, limit := range limits {
		result = append(result, &cpb.Limit{
			DailyLimit:          strconv.FormatInt(limit.DailyLimit, 10),
			DailyLimitAvailable: strconv.FormatInt(limit.DailyLimitAvailable, 10),
			LastTransaction:     date.GetDate(ctx, date.YYYYMMDD, limit.LastTransaction),
			Type:                limit.Type.String(),
		})
	}
	return result
}

func GetTimestampPtr(ctx context.Context, format date.Format, input *string) *pbtype.Date {
	if input == nil {
		return nil
	}
	return date.GetDate(ctx, format, *input)
}

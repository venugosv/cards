package ctm

import (
	"fmt"
	"sort"
	"strings"
	"time"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
)

func (r DebitCardResponse) HasEligibility(eligibility epb.Eligibility) bool {
	for _, e := range r.Eligibility() {
		if e == eligibility {
			return true
		}
	}

	return false
}

func (r DebitCardResponse) Eligibility() []epb.Eligibility {
	var eligibilitySet []epb.Eligibility
	switch r.Status {
	case StatusBlockAtmPosExcludeCnp, StatusBlockAtmPosCnpBch, StatusBlockAtmPosCnp, StatusBlockPosExcludeCnp, StatusBlockAtm, StatusBlockCnp:
		eligibilitySet = []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_FRAUD_SUSPECTED,
		}
	case StatusDelinquentRetain:
		eligibilitySet = []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_FRAUD_CARD_CANCELLED,
		}
	case StatusClosed, StatusUnissuedNdIciCards:
		eligibilitySet = []epb.Eligibility{}
	case StatusTemporaryBlock:
		eligibilitySet = []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_UNBLOCK,
		}
	case StatusLost:
		eligibilitySet = []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
		}
	case StatusStolen:
		eligibilitySet = []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
		}
	case StatusDelinquentReturn:
		eligibilitySet = []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
		}
	case StatusIssued:
		eligibilitySet = []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
			epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
			epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
		}

		if r.PinChangedCount > 0 {
			eligibilitySet = append(eligibilitySet, epb.Eligibility_ELIGIBILITY_CHANGE_PIN)
		} else {
			eligibilitySet = append(eligibilitySet, epb.Eligibility_ELIGIBILITY_SET_PIN)
		}

		if r.eligibleForActivation() {
			eligibilitySet = append(eligibilitySet, epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION)
		} else {
			eligibilitySet = append(eligibilitySet, epb.Eligibility_ELIGIBILITY_GET_DETAILS, epb.Eligibility_ELIGIBILITY_BLOCK)
		}

		if !r.IssuedOrReplacedToday() {
			replacementEligibility := []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
			}
			eligibilitySet = append(eligibilitySet, replacementEligibility...)
		}
	default:
		eligibilitySet = []epb.Eligibility{}
	}

	sort.Slice(eligibilitySet, func(i, j int) bool {
		return eligibilitySet[i] < eligibilitySet[j]
	})
	return eligibilitySet
}

func (r DebitCardResponse) eligibleForActivation() bool {
	return r.Visible() && (r.Status != StatusTemporaryBlock) && !r.ActivationStatus
}

func (r DebitCardResponse) Visible() bool {
	switch r.Status {
	case StatusBlockAtm:
		return true
	case StatusBlockAtmPosExcludeCnp:
		return true
	case StatusBlockAtmPosCnpBch:
		return true
	case StatusBlockAtmPosCnp:
		return true
	case StatusBlockCnp:
		return true
	case StatusIssued:
		return true
	case StatusBlockPosExcludeCnp:
		return true
	case StatusTemporaryBlock:
		return true
	case StatusClosed:
		return false
	case StatusDelinquentReturn:
		return false
	case StatusLost:
		return false
	case StatusStolen:
		return false
	case StatusUnissuedNdIciCards:
		return false
	case StatusDelinquentRetain:
		return false
	default:
		return false
	}
}

func (r DebitCardResponse) IssuedOrReplacedToday() bool {
	year, month, day := time.Now().Date()
	today := fmt.Sprintf("%d-%02d-%02d", year, int(month), day)
	return strings.Compare(today, r.IssueDate) == 0 || strings.Compare(today, r.ReplacedDate) == 0
}

func (s Status) Reason() *StatusReason {
	switch s {
	case StatusDelinquentReturn:
		return StatusReasonDamaged.Pointer()
	case StatusLost:
		return StatusReasonWithPinOrAccountRelated.Pointer()
	case StatusStolen:
		return StatusReasonWithPinOrAccountRelated.Pointer()
	}
	return nil
}

func (s Status) ValidReason(reason *StatusReason) bool {
	switch s {
	case StatusTemporaryBlock, StatusIssued:
		if reason == nil {
			return true
		}
	case StatusLost, StatusStolen:
		if reason != nil && *reason == StatusReasonWithPinOrAccountRelated || *reason == StatusReasonWithoutPin {
			return true
		}
	case StatusDelinquentReturn:
		if reason != nil &&
			*reason == StatusReasonWithPinOrAccountRelated ||
			*reason == StatusReasonWithoutPin ||
			*reason == StatusReasonDamaged ||
			*reason == StatusReasonLastPrimeDebitLinkageDeleted ||
			*reason == StatusReasonClosed ||
			*reason == StatusReasonFraud {
			return true
		}
	}
	return false
}

func (s Status) ValidNextState(next Status) bool {
	switch s {
	case StatusIssued:
		if next == StatusDelinquentReturn || next == StatusLost || next == StatusStolen || next == StatusTemporaryBlock {
			return true
		}
	case StatusTemporaryBlock:
		if next == StatusDelinquentReturn || next == StatusLost || next == StatusStolen || next == StatusIssued {
			return true
		}
	case StatusLost, StatusStolen, StatusDelinquentRetain:
		return false
	}
	return false
}

package ctm

import (
	"fmt"
	"sort"
	"testing"
	"time"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/stretchr/testify/assert"
)

const issueDate = "2006-01-02"

func Test_Eligibility_LifeCycle(t *testing.T) {
	year, month, day := time.Now().Date()
	today := fmt.Sprintf("%d-%02d-%02d", year, int(month), day)
	t.Run("new card delivered, inactive", func(t *testing.T) {
		brandNewCard := DebitCardResponse{
			Status:           StatusIssued,
			ActivationStatus: false,
			PinChangedCount:  0,
		}

		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
			epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
			epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION,
			epb.Eligibility_ELIGIBILITY_SET_PIN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
			epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
		}
		assert.Equal(t, want, brandNewCard.Eligibility())
	})
	t.Run("new card activated", func(t *testing.T) {
		activatedCard := DebitCardResponse{
			Status:           StatusIssued,
			ActivationStatus: true,
			PinChangedCount:  0,
		}

		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
			epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
			epb.Eligibility_ELIGIBILITY_SET_PIN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
			epb.Eligibility_ELIGIBILITY_BLOCK,
			epb.Eligibility_ELIGIBILITY_GET_DETAILS,
			epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
		}
		assert.Equal(t, want, activatedCard.Eligibility())
	})
	t.Run("new card activated and pin set", func(t *testing.T) {
		setPinCard := DebitCardResponse{
			Status:           StatusIssued,
			ActivationStatus: true,
			PinChangedCount:  1,
		}

		want := []epb.Eligibility{
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
		}
		assert.Equal(t, want, setPinCard.Eligibility())
	})
	t.Run("temp block card", func(t *testing.T) {
		lostCard := DebitCardResponse{
			Status:           StatusTemporaryBlock,
			ActivationStatus: true,
			PinChangedCount:  1,
		}

		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_UNBLOCK,
		}
		assert.Equal(t, want, lostCard.Eligibility())
	})
	t.Run("lost card", func(t *testing.T) {
		lostCard := DebitCardResponse{
			Status:           StatusLost,
			ActivationStatus: true,
			PinChangedCount:  1,
			NewCardNumber: &Card{
				Token:       "1234567890123456",
				Last4Digits: "0987",
			},
		}

		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
		}
		assert.Equal(t, want, lostCard.Eligibility())
	})
	t.Run("stolen card", func(t *testing.T) {
		lostCard := DebitCardResponse{
			Status:           StatusStolen,
			ActivationStatus: true,
			PinChangedCount:  1,
			NewCardNumber: &Card{
				Token:       "1234567890123456",
				Last4Digits: "0987",
			},
		}

		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
		}
		assert.Equal(t, want, lostCard.Eligibility())
	})
	t.Run("stolen card", func(t *testing.T) {
		damagedCard := DebitCardResponse{
			Status:           StatusDelinquentReturn,
			ActivationStatus: true,
			PinChangedCount:  1,
			NewCardNumber: &Card{
				Token:       "1234567890123456",
				Last4Digits: "0987",
			},
		}

		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
		}
		assert.Equal(t, want, damagedCard.Eligibility())
	})
	t.Run("damaged card", func(t *testing.T) {
		damagedCard := DebitCardResponse{
			Status:           StatusDelinquentReturn,
			ActivationStatus: true,
			PinChangedCount:  1,
			NewCardNumber: &Card{
				Token:       "1234567890123456",
				Last4Digits: "0987",
			},
		}

		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
		}
		assert.Equal(t, want, damagedCard.Eligibility())
	})
	t.Run("new card issued today", func(t *testing.T) {
		replacedCard := DebitCardResponse{
			Status:           StatusIssued,
			ActivationStatus: false,
			PinChangedCount:  0,
			IssueDate:        today,
		}
		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
			epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
			epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION,
			epb.Eligibility_ELIGIBILITY_SET_PIN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
			epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
		}
		assert.Equal(t, want, replacedCard.Eligibility())
	})
	t.Run("new card replaced today", func(t *testing.T) {
		replacedCard := DebitCardResponse{
			Status:           StatusIssued,
			ActivationStatus: false,
			PinChangedCount:  0,
			IssueDate:        issueDate,
			ReplacedDate:     today,
		}
		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_APPLE_PAY,
			epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
			epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
			epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION,
			epb.Eligibility_ELIGIBILITY_SET_PIN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
			epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
		}
		assert.Equal(t, want, replacedCard.Eligibility())
	})
	t.Run("card fraud suspected", func(t *testing.T) {
		card := DebitCardResponse{
			Status: StatusBlockCnp,
		}
		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_FRAUD_SUSPECTED,
		}
		assert.Equal(t, want, card.Eligibility())
	})
	t.Run("card fraud card cancelled", func(t *testing.T) {
		card := DebitCardResponse{
			Status: StatusDelinquentRetain,
		}
		want := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_FRAUD_CARD_CANCELLED,
		}
		assert.Equal(t, want, card.Eligibility())
	})
}

func getFullEligibility(activationStatus bool, pinChangeCount int64) []epb.Eligibility {
	eligibilities := []epb.Eligibility{
		epb.Eligibility_ELIGIBILITY_APPLE_PAY,
		epb.Eligibility_ELIGIBILITY_GOOGLE_PAY,
		epb.Eligibility_ELIGIBILITY_SAMSUNG_PAY,
		epb.Eligibility_ELIGIBILITY_CARD_CONTROLS,
		epb.Eligibility_ELIGIBILITY_BLOCK,
		epb.Eligibility_ELIGIBILITY_CARD_ON_FILE,
	}
	if !activationStatus {
		eligibilities = append(eligibilities, epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION)
	} else {
		e := []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
			epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			epb.Eligibility_ELIGIBILITY_GET_DETAILS,
		}

		eligibilities = append(eligibilities, e...)
	}

	if pinChangeCount > 0 {
		eligibilities = append(eligibilities, epb.Eligibility_ELIGIBILITY_CHANGE_PIN)
	} else {
		eligibilities = append(eligibilities, epb.Eligibility_ELIGIBILITY_SET_PIN)
	}
	sort.Slice(eligibilities, func(i, j int) bool {
		return eligibilities[i] < eligibilities[j]
	})
	return eligibilities
}

func TestStatus_ValidReason(t *testing.T) {
	tests := []Status{
		StatusClosed, StatusDelinquentReturn, StatusDelinquentRetain,
		StatusIssued, StatusLost, StatusStolen, StatusUnissuedNdIciCards, StatusTemporaryBlock, StatusBlockAtm,
		StatusBlockAtmPosExcludeCnp, StatusBlockAtmPosCnpBch, StatusBlockAtmPosCnp, StatusBlockCnp,
		StatusBlockPosExcludeCnp,
	}
	reasons := []StatusReason{
		StatusReasonWithPinOrAccountRelated, StatusReasonWithoutPin, StatusReasonDamaged,
		StatusReasonLastPrimeDebitLinkageDeleted, StatusReasonClosed, StatusReasonFraud,
	}
	status_StatusReason := map[Status]map[StatusReason]struct{}{
		StatusTemporaryBlock: nil,
		StatusLost:           {StatusReasonWithPinOrAccountRelated: {}, StatusReasonWithoutPin: {}},
		StatusStolen:         {StatusReasonWithPinOrAccountRelated: {}, StatusReasonWithoutPin: {}},
		StatusDelinquentReturn: {
			StatusReasonWithPinOrAccountRelated: {}, StatusReasonWithoutPin: {},
			StatusReasonDamaged: {}, StatusReasonLastPrimeDebitLinkageDeleted: {},
			StatusReasonClosed: {}, StatusReasonFraud: {},
		},
		StatusIssued: nil,
	}
	for _, status := range tests {
		for _, reason := range reasons {
			t.Run(fmt.Sprintf("%s %s", status, reason), func(t *testing.T) {
				got := status.ValidReason(reason.Pointer())
				if _, ok := status_StatusReason[status][reason]; ok {
					assert.True(t, got)
				} else {
					assert.False(t, got)
				}
			})
		}
	}
}

func TestStatus_ValidNextState(t *testing.T) {
	tests := map[Status]string{
		StatusClosed:                "Closed",
		StatusDelinquentReturn:      "Delinquent (Return Card)",
		StatusDelinquentRetain:      "Delinquent (Retain Card)",
		StatusIssued:                "Issued",
		StatusLost:                  "Lost",
		StatusStolen:                "Stolen",
		StatusUnissuedNdIciCards:    "Unissued (N&D ICI Cards)",
		StatusTemporaryBlock:        "Temporary Block",
		StatusBlockAtm:              "Block ATM",
		StatusBlockAtmPosExcludeCnp: "Block ATM & POS (Exclude CNP)",
		StatusBlockAtmPosCnpBch:     "Block ATM, POS, CNP & BCH",
		StatusBlockAtmPosCnp:        "Block ATM, POS & CNP",
		StatusBlockCnp:              "Block CNP",
		StatusBlockPosExcludeCnp:    "Block POS (exclude CNP)",
	}
	want := map[Status]map[Status]struct{}{
		StatusIssued:           {StatusDelinquentReturn: {}, StatusLost: {}, StatusStolen: {}, StatusTemporaryBlock: {}},
		StatusTemporaryBlock:   {StatusDelinquentReturn: {}, StatusLost: {}, StatusStolen: {}, StatusIssued: {}},
		StatusLost:             nil,
		StatusStolen:           nil,
		StatusDelinquentRetain: nil,
	}
	for status := range tests {
		for next := range tests {
			t.Run(fmt.Sprintf("%s to %s", status, next), func(t *testing.T) {
				got := status.ValidNextState(next)
				if _, ok := want[status][next]; ok {
					assert.True(t, got)
				} else {
					assert.False(t, got)
				}
			})
		}
	}
}

func TestDebitCardResponse_Eligibility(t *testing.T) {
	tests := []struct {
		name string
		resp DebitCardResponse
		want []epb.Eligibility
	}{
		{
			name: "Status Closed",
			resp: DebitCardResponse{
				Status: StatusClosed,
			},
			want: []epb.Eligibility{},
		}, {
			name: "Status Delinquent (Retain)",
			resp: DebitCardResponse{
				Status: StatusDelinquentRetain,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_FRAUD_CARD_CANCELLED,
			},
		}, {
			name: "Status Delinquent (Return)",
			resp: DebitCardResponse{
				Status:       StatusDelinquentReturn,
				StatusReason: StatusReasonLastPrimeDebitLinkageDeleted,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			},
		}, {
			name: "Status Delinquent (Retain)",
			resp: DebitCardResponse{
				Status:       StatusDelinquentReturn,
				StatusReason: StatusReasonWithoutPin,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			},
		}, {
			name: "Status Delinquent (Retain)",
			resp: DebitCardResponse{
				Status:       StatusDelinquentReturn,
				StatusReason: StatusReasonWithPinOrAccountRelated,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			},
		}, {
			name: "Status Delinquent (Retain)",
			resp: DebitCardResponse{
				Status:       StatusDelinquentReturn,
				StatusReason: StatusReasonClosed,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			},
		}, {
			name: "Status Delinquent (Retain)",
			resp: DebitCardResponse{
				Status:       StatusDelinquentReturn,
				StatusReason: StatusReasonDamaged,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			},
		}, {
			name: "Status Delinquent (Retain)",
			resp: DebitCardResponse{
				Status:       StatusDelinquentReturn,
				StatusReason: StatusReasonFraud,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
			},
		}, {
			name: "Status Lost",
			resp: DebitCardResponse{
				Status: StatusLost,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
			},
		}, {
			name: "Status Stolen",
			resp: DebitCardResponse{
				Status: StatusStolen,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
			},
		}, {
			name: "Status Unissued",
			resp: DebitCardResponse{
				Status: StatusUnissuedNdIciCards,
			},
			want: []epb.Eligibility{},
		}, {
			name: "Status Temporary",
			resp: DebitCardResponse{
				Status: StatusTemporaryBlock,
			},
			want: []epb.Eligibility{
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST,
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN,
				epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED,
				epb.Eligibility_ELIGIBILITY_UNBLOCK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := DebitCardResponse{
				Status:          tt.resp.Status,
				StatusReason:    tt.resp.StatusReason,
				PinChangedCount: tt.resp.PinChangedCount,
			}
			assert.Equal(t, tt.want, r.Eligibility())
		})
	}
}

func TestDebitCardResponse_Eligibility_WithStatusAndPinCount(t *testing.T) {
	statuses := []Status{StatusIssued}
	pinChangedCounts := []int64{0, 1}
	activationStatuses := []bool{true}

	for _, s := range statuses {
		for _, pinChangedCount := range pinChangedCounts {
			for _, activationStatus := range activationStatuses {
				t.Run(fmt.Sprintf("status_%s_activation_%t_pinchangecount_%d", s, activationStatus, pinChangedCount), func(t *testing.T) {
					r := DebitCardResponse{
						Status:           s,
						PinChangedCount:  pinChangedCount,
						ActivationStatus: activationStatus,
					}
					want := getFullEligibility(r.ActivationStatus, r.PinChangedCount)
					assert.Equal(t, want, r.Eligibility())
				})
			}
		}
	}
}

func TestDebitCardResponse_ReplacedToday(t *testing.T) {
	year, month, day := time.Now().Date()
	today := fmt.Sprintf("%d-%02d-%02d", year, int(month), day)
	tests := []struct {
		name string
		card DebitCardResponse
		want bool
	}{
		{
			name: "happy path not issued today",
			card: DebitCardResponse{
				IssueDate: issueDate,
			},
			want: false,
		},
		{
			name: "happy path issued today",
			card: DebitCardResponse{
				IssueDate: today,
			},
			want: true,
		},
		{
			name: "no issued date",
			card: DebitCardResponse{},
			want: false,
		},
		{
			name: "unexpected value",
			card: DebitCardResponse{
				IssueDate: "2006/01/02",
			},
			want: false,
		},
		{
			name: "happy path not replaced today",
			card: DebitCardResponse{
				IssueDate:    issueDate,
				ReplacedDate: issueDate,
			},
			want: false,
		},
		{
			name: "happy path replaced today",
			card: DebitCardResponse{
				IssueDate:    issueDate,
				ReplacedDate: today,
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.card.IssuedOrReplacedToday()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestEligibilityAppleWallet(t *testing.T) {
	tests := []struct {
		name string
		card *DebitCardResponse
		want bool
	}{
		{
			name: "A - ATM Block returns true",
			card: &DebitCardResponse{Status: StatusBlockAtm},
			want: false,
		},
		{
			name: "B - ATM and POS Block returns false",
			card: &DebitCardResponse{Status: StatusBlockAtmPosExcludeCnp},
			want: false,
		},
		{
			name: "C - Closed returns false",
			card: &DebitCardResponse{Status: StatusClosed},
			want: false,
		},
		{
			name: "D - Delinquent Return returns false",
			card: &DebitCardResponse{Status: StatusDelinquentReturn},
			want: false,
		},
		{
			name: "E - Delinquent Retain returns false",
			card: &DebitCardResponse{Status: StatusDelinquentRetain},
			want: false,
		},
		{
			name: "F - ATM, POS, CNP, Branch Block returns false",
			card: &DebitCardResponse{Status: StatusBlockAtmPosCnpBch},
			want: false,
		},
		{
			name: "G ATM, POS, CNP Block returns false",
			card: &DebitCardResponse{Status: StatusBlockAtmPosCnp},
			want: false,
		},
		{
			name: "H - CNP Block returns true",
			card: &DebitCardResponse{Status: StatusBlockCnp},
			want: false,
		},
		{
			name: "L - Lost returns false",
			card: &DebitCardResponse{Status: StatusLost},
			want: false,
		},
		{
			name: "P - POS Block returns false",
			card: &DebitCardResponse{Status: StatusBlockPosExcludeCnp},
			want: false,
		},
		{
			name: "R - Registration returns false",
			card: &DebitCardResponse{Status: StatusDelinquentRetain},
			want: false,
		},
		{
			name: "S - Stolen returns false",
			card: &DebitCardResponse{Status: StatusStolen},
			want: false,
		},
		{
			name: "T - Temporary Block returns true",
			card: &DebitCardResponse{Status: StatusTemporaryBlock},
			want: false,
		},
		{
			name: "I - Issued returns true",
			card: &DebitCardResponse{Status: StatusIssued},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want {
				assert.Contains(t, test.card.Eligibility(), epb.Eligibility_ELIGIBILITY_APPLE_PAY)
			} else {
				assert.NotContains(t, test.card.Eligibility(), epb.Eligibility_ELIGIBILITY_APPLE_PAY)
			}
		})
	}
}

func TestDebitCardResponse_Visible(t *testing.T) {
	tests := []struct {
		name string
		card DebitCardResponse
		want bool
	}{
		{
			name: "StatusBlockAtm",
			card: DebitCardResponse{Status: StatusBlockAtm},
			want: true,
		},
		{
			name: "StatusBlockAtmPosExcludeCnp",
			card: DebitCardResponse{Status: StatusBlockAtmPosExcludeCnp},
			want: true,
		},
		{
			name: "StatusBlockAtmPosCnpBch",
			card: DebitCardResponse{Status: StatusBlockAtmPosCnpBch},
			want: true,
		},
		{
			name: "StatusBlockAtmPosCnp",
			card: DebitCardResponse{Status: StatusBlockAtmPosCnp},
			want: true,
		},
		{
			name: "StatusBlockCnp",
			card: DebitCardResponse{Status: StatusBlockCnp},
			want: true,
		},
		{
			name: "StatusIssued",
			card: DebitCardResponse{Status: StatusIssued},
			want: true,
		},
		{
			name: "StatusBlockPosExcludeCnp",
			card: DebitCardResponse{Status: StatusBlockPosExcludeCnp},
			want: true,
		},
		{
			name: "StatusTemporaryBlock",
			card: DebitCardResponse{Status: StatusTemporaryBlock},
			want: true,
		},
		{
			name: "StatusClosed",
			card: DebitCardResponse{Status: StatusClosed},
			want: false,
		},
		{
			name: "StatusDelinquentReturn",
			card: DebitCardResponse{Status: StatusDelinquentReturn},
			want: false,
		},
		{
			name: "StatusLost",
			card: DebitCardResponse{Status: StatusLost},
			want: false,
		},
		{
			name: "StatusStolen",
			card: DebitCardResponse{Status: StatusStolen},
			want: false,
		},
		{
			name: "StatusUnissuedNdIciCards",
			card: DebitCardResponse{Status: StatusUnissuedNdIciCards},
			want: false,
		},
		{
			name: "StatusDelinquentRetain",
			card: DebitCardResponse{Status: StatusDelinquentRetain},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.card.Visible()
			assert.Equal(t, test.want, got)
		})
	}
}

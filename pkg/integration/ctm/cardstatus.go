package ctm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anzx/fabric-cards/pkg/util/apic"
)

const (
	cardStatusUpdate = "CardStatusUpdate"
	cardActivate     = "CardActivate"
)

type ActivateDebitCardRequest struct{}

type UpdateDebitCardStatusRequest struct {
	Status       Status        `json:"status"`
	StatusReason *StatusReason `json:"statusReason,omitempty"`
}

// Activates a given debit card in CTM.
func (c client) Activate(ctx context.Context, tokenizedCardNumber string) (bool, error) {
	activateURL := fmt.Sprintf(activationAPIUrlTemplate, c.baseURL, tokenizedCardNumber)

	body, _ := json.Marshal(ActivateDebitCardRequest{})

	_, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPost, activateURL, body), fmt.Sprintf("ctm:%s", cardActivate))
	return err == nil, err
}

// Modifies the status of a debit card in CTM.
func (c client) UpdateStatus(ctx context.Context, tokenizedCardNumber string, status Status) (bool, error) {
	statusURL := fmt.Sprintf(statusAPIUrlTemplate, c.baseURL, tokenizedCardNumber)

	req := UpdateDebitCardStatusRequest{
		Status:       status,
		StatusReason: status.Reason(),
	}
	body, _ := json.Marshal(req)

	_, err := c.apicClient.Do(ctx, apic.NewRequest(http.MethodPatch, statusURL, body), fmt.Sprintf("ctm:%s", cardStatusUpdate))

	return err == nil, err
}

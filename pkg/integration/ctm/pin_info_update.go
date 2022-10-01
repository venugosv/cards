package ctm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/anzx/fabric-cards/pkg/util/apic"
)

const (
	pinInfoUpdateDetails = "PinInfoUpdateDetails"
	ax2                  = "AX2"
	fabric               = "Fabric"
	workStation          = "VPTST"
)

type PINInfoUpdateRequest struct {
	PINTxnDate    string `json:"pinTxnDate"`
	PINTxnTime    string `json:"pinTxnTime"`
	ForwardingApp string `json:"forwardingApp"`
}

// UpdatePINInfo is to update PINChangeCount in CTM.
func (c client) UpdatePINInfo(ctx context.Context, tokenizedCardNumber string) (bool, error) {
	pinInfoUpdateUrl := fmt.Sprintf(pinInfoUpdateUrlTemplate, c.baseURL, tokenizedCardNumber)

	t := time.Now()
	req := &PINInfoUpdateRequest{
		PINTxnDate:    t.Format("20060102"),
		PINTxnTime:    t.Format("150405"),
		ForwardingApp: ax2,
	}
	body, _ := json.Marshal(req)
	request := apic.NewRequest(http.MethodPost, pinInfoUpdateUrl, body)
	request.Headers = map[string]string{
		"X-Operator-Id": fabric,
		"X-Workstation": workStation,
	}

	_, err := c.apicClient.Do(ctx, request, fmt.Sprintf("ctm:%s", pinInfoUpdateDetails))

	return err == nil, err
}

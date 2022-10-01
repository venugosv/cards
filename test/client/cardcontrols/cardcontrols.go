package cardcontrols

import (
	"testing"

	ccpbv1beta2 "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
)

type TestClient interface {
	LoadCard(t *testing.T)
	Can(eligibility epbv1beta1.Eligibility) bool
	ListControls() (*ccpbv1beta2.ListControlsResponse, error)
	QueryControls() (*ccpbv1beta2.CardControlResponse, error)
	SetControls(controls ...ccpbv1beta2.ControlType) (*ccpbv1beta2.CardControlResponse, error)
	RemoveControls(controls ...ccpbv1beta2.ControlType) (*ccpbv1beta2.CardControlResponse, error)
	TransferControls(newTokenizedCardNumber string) (*ccpbv1beta2.TransferControlsResponse, error)
	BlockCard(action ccpbv1beta2.BlockCardRequest_Action) (*ccpbv1beta2.BlockCardResponse, error)
}

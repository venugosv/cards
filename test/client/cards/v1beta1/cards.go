package v1beta1

import (
	"testing"

	cpbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
)

const (
	pinBlock1 = "XoNd2TMDJxBJsZH6pc3Fl4VBO9O5FKCf1xMh8CkqosQ98Tbw8DOA0EIpf0x6OVT2czBhypu9DfidKrJbka/45mWBqAnXFY0tN6ozOEQhSlwVuyMt1iLwG6HBwrN/X/sFO0g//L+XANa/Lw01VGarkN4OjTifxe+k1XEwZRLN7MBzOatqMyzLL2uLsGMqCtuLZ4Tojd6SzY+3Jf+LBbqH4ikyzBT54Ah4j+iZrIZTdQzWYs1reES72CGOY8K3y1wdyTAkJqjep2vZeG03+hjXE/OFl8H2HHTFv0zROX39kX9WClewBvVNey/ttS631lpSKih56QPbdw2qEdaNYAM7kg=="
)

type TestClient interface {
	LoadCard(t *testing.T)
	GetCurrentCard() *cpbv1beta1.Card
	List() (*cpbv1beta1.ListResponse, error)
	Can(eligibility epbv1beta1.Eligibility) bool
	GetDetails() (*cpbv1beta1.GetDetailsResponse, error)
	Activate() (*cpbv1beta1.ActivateResponse, error)
	GetWrappingKey() (*cpbv1beta1.GetWrappingKeyResponse, error)
	ResetPIN() (*cpbv1beta1.ResetPINResponse, error)
	SetPIN() (*cpbv1beta1.SetPINResponse, error)
	Replace(reason cpbv1beta1.ReplaceRequest_Reason) (*cpbv1beta1.ReplaceResponse, error)
	AuditTrail() (*cpbv1beta1.AuditTrailResponse, error)
	CreateApplePaymentToken() (*cpbv1beta1.CreateApplePaymentTokenResponse, error)
	CreateGooglePaymentToken() (*cpbv1beta1.CreateGooglePaymentTokenResponse, error)
}

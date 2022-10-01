package gpay

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayload(t *testing.T) {
	tests := []struct {
		name             string
		card             *ctm.DebitCardResponse
		cardNumber       string
		address          *sspb.Address
		stableHardwareID string
		want             *Payload
		wantErr          string
	}{
		{
			name:    "no expiry provided",
			wantErr: "no expiry provided",
		},
		{
			name:    "plaintext card number not provided",
			wantErr: "plaintext card number not provided",
			card: &ctm.DebitCardResponse{
				ExpiryDate: "2101",
			},
		},
		{
			name:    "no address provided",
			wantErr: "no address provided",
			card: &ctm.DebitCardResponse{
				ExpiryDate: "2101",
			},
			cardNumber: "46223930000012340",
		},
		{
			name:    "no device hardware ID provided",
			wantErr: "no device hardware ID provided",
			card: &ctm.DebitCardResponse{
				ExpiryDate: "2101",
			},
			cardNumber: "46223930000012340",
			address:    &sspb.Address{LineOne: "Unit 4", LineTwo: "15 station street", LineThree: "", City: "Reservoir", State: "AU-VI", PostalCode: "3011", Country: "AUS"},
		},
		{
			name: "happy path",
			card: &ctm.DebitCardResponse{
				ExpiryDate: "2101",
			},
			cardNumber: "46223930000012340",
			address: &sspb.Address{
				LineOne:    "Unit 4",
				LineTwo:    "15 station street",
				LineThree:  "Reservoir",
				City:       "Reservoir",
				State:      "VIC",
				PostalCode: "3011",
				Country:    "AUS",
			},
			stableHardwareID: "thisisastableID",
			want: &Payload{
				AccountNumber: "46223930000012340",
				CVV2:          "",
				Name:          "",
				ExpirationDate: ExpirationDate{
					Month: "01",
					Year:  "2021",
				},
				BillingAddress: BillingAddress{
					Line1:      "Unit 4",
					Line2:      "15 station street",
					Line3:      "Reservoir",
					City:       "Reservoir",
					State:      "VIC",
					PostalCode: "3011",
					Country:    "AU",
				},
				Provider: Provider{
					Intent:                "PUSH_PROV_MOBILE",
					ClientWalletProvider:  "40010075001",
					ClientWalletAccountID: "",
					ClientDeviceID:        "thisisastableID",
					ClientAppID:           "lotus",
					IsIDnV:                "true",
				},
			},
		},
		{
			name: "passed when some fields in address is missing",
			card: &ctm.DebitCardResponse{
				ExpiryDate: "2101",
			},
			cardNumber: "46223930000012340",
			address: &sspb.Address{
				Country: "AU",
			},
			stableHardwareID: "thisisastableID",
			want: &Payload{
				AccountNumber: "46223930000012340",
				CVV2:          "",
				Name:          "",
				ExpirationDate: ExpirationDate{
					Month: "01",
					Year:  "2021",
				},
				BillingAddress: BillingAddress{
					Country: "AU",
				},
				Provider: Provider{
					Intent:                "PUSH_PROV_MOBILE",
					ClientWalletProvider:  "40010075001",
					ClientWalletAccountID: "",
					ClientDeviceID:        "thisisastableID",
					ClientAppID:           "lotus",
					IsIDnV:                "true",
				},
			},
		},
		{
			name: "passed with an empty address",
			card: &ctm.DebitCardResponse{
				ExpiryDate: "2101",
			},
			cardNumber:       "46223930000012340",
			address:          &sspb.Address{},
			stableHardwareID: "thisisastableID",
			want: &Payload{
				AccountNumber: "46223930000012340",
				CVV2:          "",
				Name:          "",
				ExpirationDate: ExpirationDate{
					Month: "01",
					Year:  "2021",
				},
				BillingAddress: BillingAddress{},
				Provider: Provider{
					Intent:                "PUSH_PROV_MOBILE",
					ClientWalletProvider:  "40010075001",
					ClientWalletAccountID: "",
					ClientDeviceID:        "thisisastableID",
					ClientAppID:           "lotus",
					IsIDnV:                "true",
				},
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got, err := NewPayload(context.Background(), test.card, test.cardNumber, test.address, test.stableHardwareID, "")
			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				want, err := json.Marshal(test.want)
				require.NoError(t, err)
				assert.JSONEq(t, string(want), string(got))
				checkEmptyFields(t, test.address, string(got))
			}
		})
	}
}

func checkEmptyFields(t *testing.T, address *sspb.Address, got string) {
	if address.GetLineOne() == "" {
		assert.NotContains(t, got, "line1")
	}
	if address.GetLineTwo() == "" {
		assert.NotContains(t, got, "line2")
	}
	if address.GetLineThree() == "" {
		assert.NotContains(t, got, "line3")
	}
	if address.GetCity() == "" {
		assert.NotContains(t, got, "city")
	}
	if address.GetState() == "" {
		assert.NotContains(t, got, "state")
	}
	if address.GetPostalCode() == "" {
		assert.NotContains(t, got, "postalCode")
	}
	if address.GetCountry() == "" {
		assert.NotContains(t, got, "country")
	}
}

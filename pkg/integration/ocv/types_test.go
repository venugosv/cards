package ocv

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestRetrievePartyRs_GetAccount(t *testing.T) {
	tests := []struct {
		name          string
		accountNumber string
		want          *RetrievePartyRsAccount
		wantErr       bool
	}{
		{
			name:          "Happy path",
			accountNumber: "6688223654",
			want: &RetrievePartyRsAccount{
				AccountBranchNumber: "6484",
				AccountNameOne:      "WENDY DOBB",
				AccountNumber:       "000000000000006688223654",
				AccountOpenedDate:   "2012-06-08",
				AccountSubProduct:   "CAP-CIS:DDASA",
				CompanyID:           "10",
				ProductCode:         "CAP-CIS:DDA",
				RelationshipType:    "SOL",
			},
		},
		{
			name:          "unable to get account by number",
			accountNumber: "1234567890",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var r []*RetrievePartyRs
			require.NoError(t, json.Unmarshal(partyRespOK, &r))

			got := r[0].GetAccount(test.accountNumber)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestRetrievePartyRs_GetCAPCSID(t *testing.T) {
	tests := []struct {
		name      string
		partyResp []byte
		want      string
		wantErr   bool
	}{
		{
			name:      "Happy path",
			partyResp: partyRespOK,
			want:      "4018443847",
		},
		{
			name:      "unable to get cap cis id",
			partyResp: []byte(`[  {    "identifiers": [      {        "identifierUsageType": "One Customer ID",        "identifier": "1000241689",        "startDate": "2021-03-20"      },      {        "identifierUsageType": "CAP ID",        "identifier": "",        "source": "CAP-CIS",        "startDate": "2021-03-20"      }    ],    "sourceSystems": [      {        "sourceSystemName": "CAP-CIS",        "sourceSystemId": ""      }    ]  }]`),
			wantErr:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var r []*RetrievePartyRs
			require.NoError(t, json.Unmarshal(test.partyResp, &r))

			got, err := r[0].GetCAPCSID()
			if (err != nil) != test.wantErr {
				require.NoError(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

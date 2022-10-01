package ocv

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/anzx/fabric-cards/test/util"

	"github.com/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/util/apic"

	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/pkg/util/testutil"
	"github.com/stretchr/testify/assert"
)

var (
	accountNumber  = "6688223654"
	newCard        = []byte(`{"title":"MS","firstName":"CYNTHIA","lastName":"MCGOWAN","productCode":"PDV","subProductCode":"001","status":"Issued","accountsLinkedCount":2,"expiryDate":"1709","activationStatus":true,"limits":[{"type":"ATMEFTPOS","dailyLimit":1000,"dailyLimitAvailable":780,"lastTransaction":"2016-09-05"},{"type":"APO","dailyLimit":1000,"dailyLimitAvailable":1000}],"embossingLine1":"MS CYNTHIA MCGOWAN","statusCode":"I","issueReason":"Replacement","totalCards":5,"dispatchedMethod":"Sent to Branch","issueBranch":2366,"collectionBranch":2366,"collectionStatus":"Card NOT Collected","replacedDate":"2016-08-15","cardNumber":{"token":"3483942089355737","last4digits":"1012"},"reissueDate":"2013-10-21","issueDate":"2010-11-12","oldCardNumber":{"token":"9421785341492916","last4digits":"5086"},"prevExpiryDate":"201311","pinChangeDate":"2015-05-21","pinChangedCount":1,"statusChangedDate":"2016-08-31","detailsChangedDate":"2016-08-15","statusChangedUserID":"eAPI","designCode":"002","designColor":"blue","merchantUpdatePreference":true,"cardControlPreference":false}`)
	oldCard        = []byte(`{"title":"MS","firstName":"CYNTHIA","lastName":"MCGOWAN","productCode":"PDV","subProductCode":"001","status":"Stolen","accountsLinkedCount":2,"expiryDate":"1709","activationStatus":true,"limits":[{"type":"ATMEFTPOS","dailyLimit":1000,"dailyLimitAvailable":780,"lastTransaction":"2016-09-05"},{"type":"APO","dailyLimit":1000,"dailyLimitAvailable":1000}],"embossingLine1":"MS CYNTHIA MCGOWAN","statusCode":"S","issueReason":"Replacement","totalCards":5,"dispatchedMethod":"Sent to Branch","issueBranch":2366,"collectionBranch":2366,"collectionStatus":"Card NOT Collected","replacedDate":"2016-08-15","cardNumber":{"token":"9421785341492916","last4digits":"5086"},"reissueDate":"2013-10-21","issueDate":"2010-11-12","prevExpiryDate":"201311","pinChangeDate":"2015-05-21","pinChangedCount":1,"statusChangedDate":"2016-08-31","detailsChangedDate":"2016-08-15","statusChangedUserID":"eAPI","designCode":"002","designColor":"blue","merchantUpdatePreference":true,"cardControlPreference":false}`)
	contractRespOK = []byte(`{"accountNumber": "6688223654", "accountKey": "6688223654_1032_27", "accountType": "PACKAGE", "linkedParties": [{"party": {"ocvId": "304456780" }    }  ],  "linkedAccounts": [    {      "account": {        "accountNumber": "30887255367",        "accountType": "ACCOUNT",        "accountKey": "6688223654_1032_27","linkedParties": [{"party": {"ocvId": "304456780"}}]}}]}`)
	partyRespOK    = []byte(`[{"ocvId": "1000241689", "partyType": "P", "dateOfBirth": "1976-04-04", "status": "Active", "source": "CAP-CIS", "kycDetails": { "status": "CO","verificationLevel": "VM"    },    "gender": "F",    "sourceEstablishedDate": "2010-06-28",    "employeeIndicator": "N",    "employerName": "ANZ TDM PRIVATISED VALUE",    "occupation": {      "code": "253000"    },    "addresses": [      {        "addressUsageType": "Primary Mailing",        "addressLineOne": "8 WEYDALE ST",        "city": "DOUBLEVIEW",        "postalCode": "6018",        "state": "AU-WA",        "country": "AUS",        "region": "WA",        "deliveryId": "0",        "source": "CAP-CIS",        "startDate": "2016-12-04",        "endDate": "2999-12-31"      },      {        "addressUsageType": "Primary Residential",        "addressLineOne": "8 WEYDALE ST",        "city": "DOUBLEVIEW",        "postalCode": "6018",        "state": "AU-WA",        "country": "AUS",        "region": "WA",        "deliveryId": "0",        "source": "CAP-CIS",        "startDate": "2016-12-04",        "endDate": "2999-12-31"      }    ],    "phones": [      {        "phoneUsageType": "Mobile Telephone",        "phone": "+61699999999",        "preferred": "Y",        "source": "CAP-CIS",        "startDate": "2017-12-19",        "endDate": "2999-12-31"      }    ],    "emails": [      {        "emailUsageType": "Email",        "email": "MASKED@ANZ.TESTING.COM",        "source": "CAP-CIS",        "startDate": "1901-01-01"      }    ],    "identifiers": [      {        "identifierUsageType": "One Customer ID",        "identifier": "1000241689",        "startDate": "2021-03-20"      },      {        "identifierUsageType": "CAP ID",        "identifier": "4018443847",        "source": "CAP-CIS",        "startDate": "2021-03-20"      }    ],    "names": [      {        "nameUsageType": "Salutation",        "lastName": "MS DOBB",        "source": "CAP-CIS",        "startDate": "1901-01-01"      },      {        "nameUsageType": "Full Name",        "title": "MS",        "lastName": "WENDY DOBB",        "source": "CAP-CIS",        "startDate": "1901-01-01"      },      {        "nameUsageType": "Mailing Name",        "lastName": "MS W DOBB",        "source": "CAP-CIS",        "startDate": "1901-01-01"      }    ],    "preferences": [      {        "preferenceType": "Disclosure Indicator",        "preferenceValue": "N",        "preferenceReason": "Not Allowed",        "source": "CAP-CIS",        "startDate": "1901-01-01"      },      {        "preferenceType": "Advertising Indicator",        "preferenceValue": "Y",        "preferenceReason": "Allowed",        "source": "CAP-CIS",        "startDate": "1901-01-01"      }    ],    "sourceSystems": [      {        "sourceSystemName": "CAP-CIS",        "sourceSystemId": "4018443847"      }    ],    "accounts": [      {        "accountNumber": "000000000000006688223654",        "accountOpenedDate": "2012-06-08",        "accountBranchNumber": "6484",        "accountNameOne": "WENDY DOBB",        "relationshipType": "SOL",        "productCode": "CAP-CIS:DDA",        "accountSubProduct": "CAP-CIS:DDASA",        "companyId": "10",        "addresses": [          {            "addressUsageType": "Statement",            "addressLineOne": "8 WEYDALE ST",            "city": "DOUBLEVIEW",            "postalCode": "6018",            "state": "AU-WA",            "country": "AUS",            "region": "WA",            "deliveryId": "0",            "startDate": "2021-03-24",            "endDate": "2999-12-31",            "occurrenceNumber": "0",            "atomicAttributes": [              {                "type": "ACCOUNT TITLE 1",                "value": "WENDY DOBB",                "startDate": "2018-12-28"              },              {                "type": "ELIGIBILITY FLAG",                "value": "N",                "startDate": "2021-03-24"              }            ]          }        ],        "atomicAttributes": [          {            "type": "NOMINATED MAILING INDICATOR",            "value": "Y",            "startDate": "2021-03-24"          },          {            "type": "NOMINATED ADDRESS USAGE TYPE",            "value": "Primary Residential",            "startDate": "2021-03-24"          }        ]      },      {        "accountNumber": "00000000000000202546052",        "accountOpenedDate": "2010-06-28",        "accountBranchNumber": "6281",        "accountNameOne": "ANCIA TAHENY",        "relationshipType": "SOL",        "productCode": "CAP-CIS:DDA",        "accountSubProduct": "CAP-CIS:DDAED",        "companyId": "10",        "addresses": [          {            "addressUsageType": "Statement",            "addressLineOne": "8 WEYDALE ST",            "city": "DOUBLEVIEW",            "postalCode": "6018",            "state": "AU-WA",            "country": "AUS",            "region": "WA",            "deliveryId": "0",            "startDate": "2021-03-24",            "endDate": "2999-12-31",            "occurrenceNumber": "0",            "atomicAttributes": [              {                "type": "ACCOUNT TITLE 1",                "value": "ANCIA TAHENY",                "startDate": "2018-12-28"              },              {                "type": "ELIGIBILITY FLAG",                "value": "N",                "startDate": "2021-03-24"              }            ]          }        ],        "atomicAttributes": [          {            "type": "NOMINATED MAILING INDICATOR",            "value": "Y",            "startDate": "2021-03-24"          },          {            "type": "NOMINATED ADDRESS USAGE TYPE",            "value": "Primary Residential",            "startDate": "2021-03-24"          }        ]      },      {        "accountNumber": "00000000000000186701752",        "accountOpenedDate": "2012-07-16",        "accountBranchNumber": "6281",        "accountNameOne": "ANCIA TAHENY",        "relationshipType": "TPS",        "productCode": "CAP-CIS:DDA",        "accountSubProduct": "CAP-CIS:DDASA",        "companyId": "10",        "addresses": [          {            "addressUsageType": "Statement",            "addressLineOne": "26 WRIGHT AV",            "city": "CLAREMONT",            "postalCode": "6010",            "state": "AU-WA",            "country": "AUS",            "region": "WA",            "deliveryId": "0",            "startDate": "2021-03-26",            "endDate": "2999-12-31",            "occurrenceNumber": "0",            "atomicAttributes": [              {                "type": "ACCOUNT TITLE 1",                "value": "ANCIA TAHENY",                "startDate": "2018-12-28"              },              {                "type": "ELIGIBILITY FLAG",                "value": "N",                "startDate": "2021-03-26"              }            ]          }        ],        "atomicAttributes": [          {            "type": "NOMINATED MAILING INDICATOR",            "value": "Y",            "startDate": "2021-03-26"          },          {            "type": "NOMINATED ADDRESS USAGE TYPE",            "value": "Primary Residential",            "startDate": "2021-03-26"          }        ]      },      {        "accountNumber": "00000000000000978269048",        "accountOpenedDate": "2018-01-10",        "accountBranchNumber": "6281",        "accountNameOne": "ANCIA TAHENY",        "relationshipType": "TPS",        "productCode": "CAP-CIS:CDA",        "accountSubProduct": "CAP-CIS:CDATE",        "companyId": "10",        "addresses": [          {            "addressUsageType": "Statement",            "addressLineOne": "26 WRIGHT AV",            "city": "CLAREMONT",            "postalCode": "6010",            "state": "AU-WA",            "country": "AUS",            "region": "WA",            "deliveryId": "0",            "startDate": "2021-03-24",            "endDate": "2999-12-31",            "occurrenceNumber": "0",            "atomicAttributes": [              {                "type": "ACCOUNT TITLE 1",                "value": "ANCIA TAHENY",                "startDate": "2018-12-28"              },              {                "type": "ELIGIBILITY FLAG",                "value": "N",                "startDate": "2021-03-24"              }            ]          }        ],        "atomicAttributes": [          {            "type": "NOMINATED MAILING INDICATOR",            "value": "Y",            "startDate": "2021-03-24"          },          {            "type": "NOMINATED ADDRESS USAGE TYPE",            "value": "Primary Residential",            "startDate": "2021-03-24"          }        ]      },      {        "accountNumber": "00000000000000000000016",        "accountOpenedDate": "2018-12-10",        "accountBranchNumber": "3026",        "accountNameOne": "DOBB",        "accountNameTwo": "WENDY",        "accountStatus": "Unknown",        "accountStatusRaw": "OP",        "relationshipType": "SOL",        "productCode": "CAP-CIS:SVC",        "accountSubProduct": "CAP-CIS:SVCASP",        "companyId": "10",        "addresses": [          {            "addressUsageType": "Statement",            "addressLineOne": "8 WEYDALE ST",            "city": "DOUBLEVIEW",            "postalCode": "6018",            "state": "AU-WA",            "country": "AUS",            "region": "WA",            "deliveryId": "0",            "startDate": "2021-03-24",            "endDate": "2999-12-31",            "occurrenceNumber": "0",            "atomicAttributes": [              {                "type": "ACCOUNT TITLE 1",                "value": "WENDY DOBB",                "startDate": "2018-12-28"              },              {                "type": "ELIGIBILITY FLAG",                "value": "N",                "startDate": "2021-03-24"              }            ]          }        ],        "atomicAttributes": [          {            "type": "NOMINATED MAILING INDICATOR",            "value": "Y",            "startDate": "2021-03-24"          },          {            "type": "NOMINATED ADDRESS USAGE TYPE",            "value": "Primary Residential",            "startDate": "2021-03-24"          }        ]      },      {        "accountNumber": "00000000000000202544882",        "accountOpenedDate": "2010-06-28",        "accountBranchNumber": "6281",        "accountNameOne": "ANCIA TAHENY",        "relationshipType": "SOL",        "productCode": "CAP-CIS:DDA",        "accountSubProduct": "CAP-CIS:DDAPT",        "companyId": "10",        "addresses": [          {            "addressUsageType": "Statement",            "addressLineOne": "8 WEYDALE ST",            "city": "DOUBLEVIEW",            "postalCode": "6018",            "state": "AU-WA",            "country": "AUS",            "region": "WA",            "deliveryId": "0",            "startDate": "2021-03-24",            "endDate": "2999-12-31",            "occurrenceNumber": "0",            "atomicAttributes": [              {                "type": "ACCOUNT TITLE 1",                "value": "ANCIA TAHENY",                "startDate": "2018-12-28"              },              {                "type": "ELIGIBILITY FLAG",                "value": "N",                "startDate": "2021-03-24"              }            ]          }        ],        "atomicAttributes": [          {            "type": "NOMINATED MAILING INDICATOR",            "value": "Y",            "startDate": "2021-03-24"          },          {            "type": "NOMINATED ADDRESS USAGE TYPE",            "value": "Primary Residential",            "startDate": "2021-03-24"          }        ]      },      {        "accountNumber": "00000000000000186701744",        "accountOpenedDate": "2012-07-16",        "accountBranchNumber": "6281",        "accountNameOne": "WENDY DOBB",        "relationshipType": "TPS",        "productCode": "CAP-CIS:DDA",        "accountSubProduct": "CAP-CIS:DDAPT",        "companyId": "10",        "addresses": [          {            "addressUsageType": "Statement",            "addressLineOne": "26 WRIGHT AV",            "city": "CLAREMONT",            "postalCode": "6010",            "state": "AU-WA",            "country": "AUS",            "region": "WA",            "deliveryId": "0",            "startDate": "2021-03-26",            "endDate": "2999-12-31",            "occurrenceNumber": "0",            "atomicAttributes": [              {                "type": "ACCOUNT TITLE 1",                "value": "WENDY DOBB",                "startDate": "2018-12-28"              },              {                "type": "ELIGIBILITY FLAG",                "value": "N",                "startDate": "2021-03-26"              }            ]          }        ],        "atomicAttributes": [          {            "type": "NOMINATED MAILING INDICATOR",            "value": "Y",            "startDate": "2021-03-26"          },          {            "type": "NOMINATED ADDRESS USAGE TYPE",            "value": "Primary Residential",            "startDate": "2021-03-26"          }        ]      }    ],    "links": [],    "historicDetails": {      "identifiers": []    }  }]`)
)

type mockSecretManager struct {
	name    string
	payload string
	err     error
}

func (m mockSecretManager) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name:    m.name,
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(m.payload)},
	}, m.err
}

func gsmClient() *gsm.Client {
	return &gsm.Client{
		SM: mockSecretManager{
			name:    "testName",
			payload: "password",
		},
	}
}

func TestClientFromConfig(t *testing.T) {
	gsmClient := gsmClient()
	key := "ClientIDKey"

	t.Run("New Client with httpClient supplied", func(t *testing.T) {
		server := httptest.NewServer(nil)
		got, err := ClientFromConfig(context.Background(), server.Client(), &Config{ClientIDEnvKey: key}, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("New Client without httpClient supplied", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, &Config{ClientIDEnvKey: key}, gsmClient)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("New Client without config supplied", func(t *testing.T) {
		got, err := ClientFromConfig(context.Background(), nil, nil, gsmClient)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestClient_AccountMaintenance(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr string
		apic    apic.Clienter
		party   []byte
	}{
		{
			name: "successfully complete call",
			apic: util.MockAPIcer{
				DoCall: func(ctx context.Context, r *apic.Request, op string) ([]byte, error) {
					require.Contains(t, r.Headers, "RequestTime")

					if strings.Contains(r.Destination, "contract") {
						return contractRespOK, nil
					}
					return nil, errors.New("oh no")
				},
			},
			party: partyRespOK,
		},
		{
			name:    "APIc returns a fail response",
			url:     "https://apisit04.corp.dev.anz/ocv",
			wantErr: "oh no",
			apic: util.MockAPIcer{
				DoCall: func(ctx context.Context, r *apic.Request, op string) ([]byte, error) {
					require.Contains(t, r.Headers, "RequestTime")
					return nil, errors.New("oh no")
				},
			},
			party: partyRespOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := client{
				baseURL:    test.url,
				apicClient: test.apic,
			}

			var newCardDetails *ctm.DebitCardResponse
			require.NoError(t, json.Unmarshal(newCard, &newCardDetails))
			var oldCardDetails *ctm.DebitCardResponse
			require.NoError(t, json.Unmarshal(oldCard, &oldCardDetails))
			var partyResp []*RetrievePartyRs
			require.NoError(t, json.Unmarshal(test.party, &partyResp))

			account, err := GetAccount(partyResp, accountNumber)
			require.NoError(t, err)

			got, err := c.AccountMaintenance(testutil.GetContext(true), "OCVID", oldCardDetails, newCardDetails, account)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.True(t, got)
			}
		})
	}
}

func TestClient_RetrieveParty(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		apic    apic.Clienter
		wantErr string
	}{
		{
			name: "successfully get party response",
			url:  "https://apisit04.corp.dev.anz/ocv",
			apic: util.MockAPIcer{
				DoCall: func(context.Context, *apic.Request, string) ([]byte, error) {
					return partyRespOK, nil
				},
			},
		},
		{
			name: "APIc returns fail get party response",
			url:  "https://apisit04.corp.dev.anz/ocv",
			apic: util.MockAPIcer{
				DoCall: func(context.Context, *apic.Request, string) ([]byte, error) {
					return nil, errors.New("oh no")
				},
			},
			wantErr: "oh no",
		},
		{
			name: "APIc returns unexpected get party response",
			url:  "https://apisit04.corp.dev.anz/ocv",
			apic: util.MockAPIcer{
				DoCall: func(context.Context, *apic.Request, string) ([]byte, error) {
					return []byte(`%%`), nil
				},
			},
			wantErr: "invalid character",
		},
		{
			name:    "fail to parse URL",
			url:     "%%",
			wantErr: "invalid URL escape \"%%/\"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := client{
				baseURL:    test.url,
				apicClient: test.apic,
			}
			got, err := c.RetrieveParty(testutil.GetContext(true), "ODVID")
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestGetPartyAccount(t *testing.T) {
	t.Run("test.name", func(t *testing.T) {
		wantAccount := &RetrievePartyRsAccount{
			AccountBranchNumber: "6484",
			AccountNameOne:      "WENDY DOBB",
			AccountNumber:       "000000000000006688223654",
			AccountOpenedDate:   "2012-06-08",
			AccountSubProduct:   "CAP-CIS:DDASA",
			CompanyID:           "10",
			ProductCode:         "CAP-CIS:DDA",
			RelationshipType:    "SOL",
		}

		var parties []*RetrievePartyRs
		require.NoError(t, json.Unmarshal(partyRespOK, &parties))

		account, err := GetAccount(parties, accountNumber)
		assert.NoError(t, err)
		assert.Equal(t, wantAccount, account)
	})
}

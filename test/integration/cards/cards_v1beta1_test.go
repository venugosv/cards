//go:build integration
// +build integration

package cards

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/anzx/fabric-cards/test/client/cards/v1beta1"
	"github.com/anzx/fabric-cards/test/config"

	"github.com/anzx/fabric-cards/pkg/integration/vault"
	"github.com/anzx/fabric-cards/test/common"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"
)

type suiteTarget int

const (
	TargetGrpc suiteTarget = iota
	TargetRest
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type v1beta1TestSuite struct {
	suite.Suite
	v1beta1  v1beta1.TestClient
	tearDown bool
	target   suiteTarget
	toggle   *config.Toggle
}

// Make sure that the v1beta1 is set up before the suite begins

func (c *v1beta1TestSuite) SetupSuite() {
	cfg, err := config.Load(c.T())
	if err != nil {
		c.T().Fatal("SetupSuite failed to load config", err)
	}
	c.T().Logf("config loaded: \n%v", cfg.String())
	ctx, _ := context.WithTimeout(context.Background(), cfg.Timeout)

	cardsConfig := cfg.Cards
	c.toggle = config.NewToggle(cardsConfig.Toggle)
	c.tearDown = cardsConfig.TearDown

	conn := common.GetConnection(c.T(), ctx, cardsConfig.BaseURL, cardsConfig.Insecure, cardsConfig.Headers...)

	vaultClient, err := vault.NewClient(ctx, nil, cfg.Vault)
	require.NoError(c.T(), err, "failed to create test vault client")
	require.NotNil(c.T(), vaultClient, "vault client not created")

	user := common.GetUser(c.T(), cardsConfig, conn, common.V1beta1CardAPI)

	auth := common.GetAuthHeaders(c.T(), user, cardsConfig.Auth, common.CardsScope...)

	switch c.target {
	case TargetGrpc:
		ctx = auth.Context(c.T(), ctx, cardsConfig.Headers...)
		c.v1beta1 = v1beta1.NewGRPCClient(ctx, conn, vaultClient)
	case TargetRest:
		headers := auth.GetHeadersHTTP(cardsConfig.Headers...)
		c.v1beta1 = v1beta1.NewRESTTestClient(ctx, cardsConfig.BaseURL, cardsConfig.Insecure, headers, vaultClient)
	default:
		c.T().Errorf("unknown target, should be GRPC (%d) or REST (%d), but it was = %d", TargetGrpc, TargetRest, c.target)
		c.T().FailNow()
	}
}

// will run before each test in the suite.
func (c *v1beta1TestSuite) SetupTest() {
	c.v1beta1.LoadCard(c.T())
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_0_ReplaceDamaged() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIReplaceDamaged)

	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED) {
		resp, err := c.v1beta1.Replace(cpb.ReplaceRequest_REASON_DAMAGED)
		if err != nil {
			if knownError(err) {
				c.T().Skip("Skipped Replace (Damaged) due to known error")
			} else {
				assert.NoError(c.T(), err)
			}
		}
		assert.NotNil(c.T(), resp)
		assert.True(c.T(), len(resp.GetEligibilities()) > 0)
	} else {
		c.T().Skip("Skipped Replace (Damaged) due to no eligibility")
	}
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_0_ReplaceLost() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIReplaceLost)

	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST) {
		resp, err := c.v1beta1.Replace(cpb.ReplaceRequest_REASON_LOST)
		if err != nil {
			if knownError(err) {
				c.T().Skip("Skipped Replace (Lost) due to related data error")
			} else {
				require.NoError(c.T(), err)
			}
		}
		assert.NotNil(c.T(), resp)
		assert.True(c.T(), len(resp.GetEligibilities()) > 0)
	} else {
		c.T().Skip("Skipped Replace (Lost) due to no eligibility")
	}
}

// All methods that begin with "Test" are run as tests within a suite.
func (c *v1beta1TestSuite) TestV1beta1CardAPI_1_List() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIList)

	resp, err := c.v1beta1.List()
	require.NoError(c.T(), err)
	for _, card := range resp.GetCards() {
		verifyCard(c.T(), card)
	}
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_2_Activate() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIActivate)

	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION) {
		_, err := c.v1beta1.Activate()
		assert.NoError(c.T(), err)
	} else {
		c.T().Skip("Skipped Activate due to no eligibility")
	}
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_3_GetWrappingKey() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIGetWrappingKey)

	resp, err := c.v1beta1.GetWrappingKey()
	require.NoError(c.T(), err)
	assert.NotNil(c.T(), resp.EncodedKey)
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_4_SetPIN() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPISetPIN)

	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_SET_PIN) {
		_, err := c.v1beta1.SetPIN()
		assert.NoError(c.T(), err)
		c.v1beta1.LoadCard(c.T())
		assert.True(c.T(), c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_CHANGE_PIN))
	} else {
		c.T().Skip("Skipped SetPIN due to no eligibility")
	}
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_5_ChangePIN() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIChangePIN)

	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_CHANGE_PIN) {
		_, err := c.v1beta1.ResetPIN()
		assert.NoError(c.T(), err)
		c.v1beta1.LoadCard(c.T())
		assert.False(c.T(), c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_SET_PIN))
	} else {
		c.T().Skip("Skipped ChangePIN due to no eligibility")
	}
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_6_GetDetails() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIGetDetails)

	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_GET_DETAILS) {
		resp, err := c.v1beta1.GetDetails()
		require.NoError(c.T(), err)
		verifyCardDetails(c.T(), resp)
	} else {
		c.T().Skip("Skipped GetDetails due to no eligibility")
	}
}

func (c *v1beta1TestSuite) TestV1beta1CardAPI_7_AuditTrail() {
	c.toggle.Skip(c.T(), config.V1beta1CardAPIAuditTrail)

	resp, err := c.v1beta1.AuditTrail()
	require.NoError(c.T(), err)
	assert.True(c.T(), resp.GetActivated())
	assert.Equal(c.T(), resp.GetStatus(), "Issued")
	assert.True(c.T(), resp.GetAccountsLinked() > 0)
}

func (c *v1beta1TestSuite) TestV1beta1WalletAPI_CreateApplePaymentToken() {
	c.toggle.Skip(c.T(), config.V1beta1WalletAPICreateApplePaymentToken)

	if c.target == TargetRest {
		c.T().Skip()
	}
	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_APPLE_PAY) {
		resp, err := c.v1beta1.CreateApplePaymentToken()
		if err != nil {
			require.NoError(c.T(), err)
		}
		assert.NotNil(c.T(), resp.GetActivationData())
		assert.NotNil(c.T(), resp.GetEncryptedPassData())
		assert.NotNil(c.T(), resp.GetEphemeralPublicKey())
	} else {
		c.T().Skip("Skipped create apple payment token due to no eligibility")
	}
}

func (c *v1beta1TestSuite) TestV1beta1WalletAPI_CreateGooglePaymentToken() {
	c.toggle.Skip(c.T(), config.V1beta1WalletAPICreateGooglePaymentToken)

	if c.target == TargetRest {
		c.T().Skip()
	}
	if c.v1beta1.Can(epb.Eligibility_ELIGIBILITY_GOOGLE_PAY) {
		resp, err := c.v1beta1.CreateGooglePaymentToken()
		if err != nil {
			require.NoError(c.T(), err)
		}
		assert.NotNil(c.T(), resp.GetOpaquePaymentCard())
		assert.Equal(c.T(), cpb.TokenProvider_TOKEN_PROVIDER_VISA, resp.GetTokenProvider())
		assert.Equal(c.T(), cpb.CardNetwork_CARD_NETWORK_VISA, resp.GetCardNetwork())
		assert.NotNil(c.T(), resp.GetUserAddress())
	} else {
		c.T().Skip("Skipped create google payment token due to no eligibility")
	}
}

func verifyCard(t *testing.T, card *cpb.Card) {
	t.Helper()
	assert.Regexp(t, regexp.MustCompile(`^.+`), card.TokenizedCardNumber, fmt.Sprintf("wrong card number :%v", card.TokenizedCardNumber))
}

func verifyCardDetails(t *testing.T, card *cpb.GetDetailsResponse) {
	t.Helper()
	assert.Regexp(t, regexp.MustCompile(`^\d{16}`), card.GetCardNumber(), fmt.Sprintf("wrong card number :%v", card.GetCardNumber()))
	assert.True(t, len(card.GetEligibilities()) > 0, fmt.Sprintf("wrong card number :%v", card.GetEligibilities()))
}

func knownError(err error) bool {
	if strings.Contains(err.Error(), "cannot replace card number on the same day it was created") ||
		strings.Contains(err.Error(), "unable to get party and account information for card") {
		return true
	}
	return false
}

func (c *v1beta1TestSuite) TearDownSuite() {
	if c.tearDown {
		require.NoError(c.T(), common.TearDown(c.T()))
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGRPCCardsSuite(t *testing.T) {
	suite.Run(t, &v1beta1TestSuite{
		target: TargetGrpc,
	})
}

func TestRESTCardsSuite(t *testing.T) {
	suite.Run(t, &v1beta1TestSuite{
		target: TargetRest,
	})
}

//go:build integration
// +build integration

package cardcontrols

import (
	"context"
	"fmt"
	"testing"

	"github.com/anzx/fabric-cards/test/client/cardcontrols"
	"github.com/anzx/fabric-cards/test/client/cards/v1beta1"
	"github.com/anzx/fabric-cards/test/config"

	v1beta2pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/anzx/fabric-cards/test/common"

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
type V1beta2TestSuite struct {
	suite.Suite
	v1beta2  cardcontrols.TestClient
	tearDown bool
	toggle   *config.Toggle
	target   suiteTarget
}

// Make sure that the v1beta2 is set up before the suite begins
func (c *V1beta2TestSuite) SetupSuite() {
	cfg, err := config.Load(c.T())
	if err != nil {
		c.T().Fatal("SetupSuite failed to load config", err)
	}
	c.T().Logf("config loaded: \n%v", cfg.String())
	cards := cfg.Cards
	cardControls := cfg.CardControls
	c.toggle = config.NewToggle(cardControls.Toggle)
	c.tearDown = cardControls.TearDown

	// Cards
	ctx, _ := context.WithTimeout(context.Background(), cfg.Timeout)
	cardsConn := common.GetConnection(c.T(), ctx, cards.BaseURL, cards.Insecure)
	user := common.GetUser(c.T(), cardControls, cardsConn, common.V1beta2CardControlsAPI)
	cardsAuth := common.GetAuthHeaders(c.T(), user, cardControls.Auth, common.CardsScope...)
	ctx = cardsAuth.Context(c.T(), ctx)
	cardsClient := v1beta1.NewGRPCClient(ctx, cardsConn, nil)

	// CardControls
	ctx, _ = context.WithTimeout(context.Background(), cfg.Timeout)
	cardControlsConn := common.GetConnection(c.T(), ctx, cardControls.BaseURL, cardControls.Insecure, cardControls.Headers...)
	auth := common.GetAuthHeaders(c.T(), user, cardControls.Auth, common.CardControlScope...)
	switch c.target {
	case TargetGrpc:
		ctx = auth.Context(c.T(), ctx, cardControls.Headers...)
		c.v1beta2 = cardcontrols.NewGRPCV1beta2Client(ctx, cardsClient, cardControlsConn)
	case TargetRest:
		headers := auth.GetHeadersHTTP(cardControls.Headers...)
		cardsHeaders := cardsAuth.GetHeadersHTTP()
		c.v1beta2 = cardcontrols.NewRESTV1beta2Client(ctx, cards.BaseURL, cardControls.BaseURL, cardControls.Insecure, headers, cardsHeaders)
	default:
		c.T().Errorf("unknown target, should be GRPC (%d) or REST (%d), but it was = %d", TargetGrpc, TargetRest, c.target)
		c.T().FailNow()
	}
}

// will run before each test in the suite.
func (c *V1beta2TestSuite) SetupTest() {
	c.v1beta2.LoadCard(c.T())
}

const testingControlType = v1beta2pb.ControlType_TCT_CONTACTLESS

// All methods that begin with "Test" are run as tests within a suite.
func (c *V1beta2TestSuite) TestV1beta2CardControlsAPI_ListControls() {
	c.toggle.Skip(c.T(), config.V1beta2CardControlsAPIListControls)

	if c.v1beta2.Can(epb.Eligibility_ELIGIBILITY_CARD_CONTROLS) {
		resp, err := c.v1beta2.ListControls()
		require.NoError(c.T(), err)
		require.NotNil(c.T(), resp)
		assert.True(c.T(), len(resp.GetCardControls()) > 0)
	} else {
		c.T().Skip("Skipped ListControls due to no eligibility")
	}
}

func (c *V1beta2TestSuite) TestV1beta2CardControlsAPI_QueryControls() {
	c.toggle.Skip(c.T(), config.V1beta2CardControlsAPIQueryControls)

	if c.v1beta2.Can(epb.Eligibility_ELIGIBILITY_CARD_CONTROLS) {
		resp, err := c.v1beta2.QueryControls()
		require.NoError(c.T(), err)
		require.NotNil(c.T(), resp)
	} else {
		c.T().Skip("Skipped queryControls due to no eligibility")
	}
}

func (c *V1beta2TestSuite) TestV1beta2CardControlsAPI_SetControls() {
	c.toggle.Skip(c.T(), config.V1beta2CardControlsAPISetControls)

	if c.v1beta2.Can(epb.Eligibility_ELIGIBILITY_CARD_CONTROLS) {
		resp, err := c.v1beta2.SetControls(testingControlType)
		if err != nil {
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.AlreadyExists {
					c.T().Skip(fmt.Sprintf("Control type already exists: %v", testingControlType))
				}
			}
			require.NoError(c.T(), err)
		}

		verifyContainControl(c.T(), resp.GetCardControls(), testingControlType)
	} else {
		c.T().Skip("Skipped setControls due to no eligibility")
	}
}

func (c *V1beta2TestSuite) TestV1beta2CardControlsAPI_RemoveControls() {
	c.toggle.Skip(c.T(), config.V1beta2CardControlsAPIRemoveControls)

	if c.v1beta2.Can(epb.Eligibility_ELIGIBILITY_CARD_CONTROLS) {
		resp, err := c.v1beta2.RemoveControls(testingControlType)
		require.NoError(c.T(), err)

		verifyNotContainControl(c.T(), resp.GetCardControls(), testingControlType)
	} else {
		c.T().Skip("Skipped removeControls due to no eligibility")
	}
}

func (c *V1beta2TestSuite) TestV1beta2CardControlsAPI_BlockCard() {
	c.toggle.Skip(c.T(), config.V1beta2CardControlsAPIBlockCard)

	if c.target == TargetRest {
		c.T().Skip()
	}

	if c.v1beta2.Can(epb.Eligibility_ELIGIBILITY_BLOCK) {
		resp, err := c.v1beta2.BlockCard(v1beta2pb.BlockCardRequest_ACTION_BLOCK)
		require.NoError(c.T(), err)

		verifyNotContainEligibility(c.T(), resp.GetEligibilities(), epb.Eligibility_ELIGIBILITY_BLOCK)
		verifyContainEligibility(c.T(), resp.GetEligibilities(), epb.Eligibility_ELIGIBILITY_UNBLOCK)
	} else {
		c.T().Skip("Skipped block due to no eligibility")
	}
}

func (c *V1beta2TestSuite) TestV1beta2CardControlsAPI_UnBlockCard() {
	c.toggle.Skip(c.T(), config.V1beta2CardControlsAPIUnBlockCard)

	if c.target == TargetRest {
		c.T().Skip()
	}

	if c.v1beta2.Can(epb.Eligibility_ELIGIBILITY_UNBLOCK) {
		resp, err := c.v1beta2.BlockCard(v1beta2pb.BlockCardRequest_ACTION_UNBLOCK)
		require.NoError(c.T(), err)

		verifyContainEligibility(c.T(), resp.GetEligibilities(), epb.Eligibility_ELIGIBILITY_BLOCK)
		verifyNotContainEligibility(c.T(), resp.GetEligibilities(), epb.Eligibility_ELIGIBILITY_UNBLOCK)
	} else {
		c.T().Skip("Skipped unblock due to no eligibility")
	}
}

func verifyContainControl(t *testing.T, controlSet []*v1beta2pb.CardControl, want v1beta2pb.ControlType) {
	t.Helper()
	control := findControl(controlSet, want)
	assert.NotNil(t, control, fmt.Sprintf("cannot find control type: %v in %+v", want.String(), controlSet))
}

func verifyNotContainControl(t *testing.T, controlSet []*v1beta2pb.CardControl, want v1beta2pb.ControlType) {
	t.Helper()
	control := findControl(controlSet, want)
	assert.Nil(t, control, fmt.Sprintf("control type: %v is not supposed in %+v", want.String(), controlSet))
}

func findControl(controlSet []*v1beta2pb.CardControl, want v1beta2pb.ControlType) *v1beta2pb.CardControl {
	var control *v1beta2pb.CardControl
	for _, c := range controlSet {
		if c.ControlType == want {
			control = c
			break
		}
	}
	return control
}

func verifyContainEligibility(t *testing.T, eligibilitySet []epb.Eligibility, want epb.Eligibility) {
	t.Helper()
	eligibility := findEligibility(eligibilitySet, want)
	assert.Equal(t, want, eligibility, fmt.Sprintf("cannot find eligibility: %v in %+v", want.String(), eligibilitySet))
}

func verifyNotContainEligibility(t *testing.T, eligibilitySet []epb.Eligibility, want epb.Eligibility) {
	t.Helper()
	eligibility := findEligibility(eligibilitySet, want)
	assert.NotEqual(t, want, eligibility, fmt.Sprintf("Eligibility: %v is not supposed to be in %+v", want.String(), eligibilitySet))
}

func findEligibility(eligibilitySet []epb.Eligibility, want epb.Eligibility) epb.Eligibility {
	var eligibility epb.Eligibility
	for _, c := range eligibilitySet {
		if c == want {
			eligibility = c
			break
		}
	}
	return eligibility
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGRPCV1beta2Suite(t *testing.T) {
	suite.Run(t, &V1beta2TestSuite{
		target: TargetGrpc,
	})
}

func TestRESTV1beta2Suite(t *testing.T) {
	suite.Run(t, &V1beta2TestSuite{
		target: TargetRest,
	})
}

func (c *V1beta2TestSuite) TearDownSuite() {
	if c.tearDown {
		require.NoError(c.T(), common.TearDown(c.T()))
	}
}

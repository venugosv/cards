//go:build integration
// +build integration

package callback

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/anzx/fabric-cards/test/client/callback"
	"github.com/anzx/fabric-cards/test/config"
	"github.com/golang/protobuf/proto"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/commandcentre/v1beta1"
	"github.com/stretchr/testify/assert"

	"github.com/anzx/fabric-cards/test/common"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CallbackTestSuite struct {
	suite.Suite
	callback                      *callback.Callback
	tearDown                      bool
	commandCentreEvents           *pubsub.Subscription
	commandCentreConnectionCancel context.CancelFunc
	pubsubTimeout                 time.Duration
	pubsubSkip                    bool
}

// Make sure that the callback is set up before the suite begins
func (c *CallbackTestSuite) SetupSuite() {
	cfg, err := config.Load(c.T())
	if err != nil {
		c.T().Fatal("SetupSuite failed to load config", err)
	}
	c.T().Logf("running callback integration tests with config: %v", cfg)
	ctx, _ := context.WithTimeout(context.Background(), cfg.Timeout)

	callbackConfig := cfg.Callback
	conn := common.GetConnection(c.T(), ctx, callbackConfig.BaseURL, callbackConfig.Insecure, callbackConfig.Headers...)
	user := common.GetUser(c.T(), callbackConfig.Service, conn, common.CallBack)
	auth := common.GetAuthHeaders(c.T(), user, callbackConfig.Auth)
	ctx = auth.Context(c.T(), ctx, callbackConfig.Headers...)

	c.tearDown = callbackConfig.TearDown

	c.callback = callback.NewCallback(ctx, conn, callbackConfig.CurrentCard)

	c.pubsubTimeout = callbackConfig.PubsubTimeout
	c.pubsubSkip = callbackConfig.PubsubSkip

	if callbackConfig.PubsubSkip {
		c.commandCentreConnectionCancel = func() {}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.pubsubTimeout)
	pubsubCon, err := grpc.DialContext(ctx, cfg.CommandCentre.EmulatorHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(c.T(), err, "failed to dial pubsub")
	pubsubClient, err := pubsub.NewClient(ctx, cfg.CommandCentre.ProjectID, option.WithGRPCConn(pubsubCon))
	require.NoError(c.T(), err, "failed to create pubsub client")
	c.commandCentreEvents = pubsubClient.Subscription(cfg.CommandCentre.Subscription)
	c.commandCentreConnectionCancel = cancel
}

// All methods that begin with "Test" are run as tests within a suite.
func (c *CallbackTestSuite) TestEnrollmentAPI() {
	c.T().Run("Enroll", func(t *testing.T) {
		got, err := c.callback.Enroll()
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
	c.T().Run("Disenroll", func(t *testing.T) {
		got, err := c.callback.Disenroll()
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
}

func (c *CallbackTestSuite) TestNotificationAPINoPubsub_Alert() {
	got, err := c.callback.Alert()
	require.NoError(c.T(), err)
	assert.NotNil(c.T(), got)
}

func (c *CallbackTestSuite) TestNotificationAPI_Alert() {
	if c.pubsubSkip {
		c.T().Skip("cannot run in this environment")
	}
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubsubMessageValue := make(chan string)

	// Timeout gracefully, cancelling the context, closing the channel and failing the test
	timeout := time.NewTimer(c.pubsubTimeout)
	go func() {
		<-timeout.C
		cancel()
		pubsubMessageValue <- ""
		c.T().Fail()
	}()

	go func() {
		err := c.commandCentreEvents.Receive(cancelCtx, func(_ context.Context, msg *pubsub.Message) {
			defer msg.Ack()
			var notification ccpb.CommandTrigger
			err := proto.Unmarshal(msg.Data, &notification)
			require.NoError(c.T(), err)
			previewBody := notification.GetContent().GetNotification().GetSimpleNotification().GetPreview().GetBody()
			if len(previewBody) > 0 {
				pubsubMessageValue <- previewBody
			} else {
				c.T().Logf("other message on pubsub = %+v", &notification)
			}
		})
		require.NoError(c.T(), err, "something bad happened receiving from CC pubsub")
	}()

	got, err := c.callback.Alert()
	require.NoError(c.T(), err)
	assert.NotNil(c.T(), got)

	cmp := <-pubsubMessageValue
	cancel()
	expected := fmt.Sprintf("A transaction of %s%0.2f (Grill'd Healthy Burgers) was declined because of a control you placed on your card ending in %s", callback.TransactionCurrencyCodeName, callback.TransactionValue, c.callback.GetLast4Digits())
	require.Equal(c.T(), expected, cmp)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCallbackSuite(t *testing.T) {
	suite.Run(t, new(CallbackTestSuite))
}

func (c *CallbackTestSuite) TearDownSuite() {
	c.commandCentreConnectionCancel()
	if c.tearDown {
		require.NoError(c.T(), common.TearDown(c.T()))
	}
}

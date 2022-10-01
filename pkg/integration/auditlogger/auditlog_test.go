package auditlogger

import (
	"context"
	"testing"

	"github.com/anzx/pkg/auditlog/pubsub"

	"github.com/stretchr/testify/require"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/anzx/fabric-cards/test/fixtures"
	"github.com/anzx/pkg/auditlog"
)

func TestPublish(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		ctx, b := fixtures.GetTestContextWithLogger(nil)

		c := Client{
			Publisher: fixtures.AServer().AuditLogPublisher,
		}
		c.Publish(ctx, auditlog.EventActivateCard, &cpb.ActivateResponse{Eligibilities: []epb.Eligibility{
			epb.Eligibility_ELIGIBILITY_CARD_ACTIVATION,
		}}, nil, nil)

		assert.Contains(t, b.String(), "AuditLog Published Successfully")
	})
	t.Run("error case", func(t *testing.T) {
		ctx, b := fixtures.GetTestContextWithLogger(nil)

		c := Client{
			Publisher: fixtures.AServer().WithAuditLogError(errors.New("oh no")).AuditLogPublisher,
		}
		c.Publish(ctx, auditlog.EventActivateCard, &cpb.ActivateResponse{Eligibilities: []epb.Eligibility{}}, nil, nil)

		assert.Contains(t, b.String(), "Error publishing auditLog message")
	})
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	t.Run("nil pubsub config", func(t *testing.T) {
		config := &auditlog.Config{
			PubSub: nil,
		}
		got, err := NewClient(ctx, config)
		require.NoError(t, err)
		require.Nil(t, got)
	})
	t.Run("nil config", func(t *testing.T) {
		got, err := NewClient(ctx, nil)
		require.NoError(t, err)
		require.Nil(t, got)
	})
	t.Run("nil pubsub config", func(t *testing.T) {
		config := &auditlog.Config{
			PubSub: &pubsub.Config{
				ProjectID:                 "",
				TopicID:                   "",
				EmulatorHost:              "",
				Block:                     false,
				ConnectionTimeoutDuration: 0,
			},
		}
		got, err := NewClient(ctx, config)
		require.Error(t, err)
		require.Nil(t, got)
	})
}

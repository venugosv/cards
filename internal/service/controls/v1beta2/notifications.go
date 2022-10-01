package v1beta2

import (
	"context"
	"fmt"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/notification"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	"github.com/anzx/pkg/log"
	"github.com/google/uuid"
)

var controlFormats = map[ccpb.ControlType]string{
	ccpb.ControlType_GCT_GLOBAL:       "",
	ccpb.ControlType_TCT_ATM_WITHDRAW: "ATM withdrawals",
	ccpb.ControlType_TCT_CROSS_BORDER: "overseas transactions (in-store)",
	ccpb.ControlType_TCT_E_COMMERCE:   "online transactions",
	ccpb.ControlType_TCT_CONTACTLESS:  "contactless payments",
}

// SendNotifications will be called when controls are removed or set by a coach (i.e. not the customer)
// If controls are being set, settingControl will be true, and notifications will say xxx has been DISABLED
// If false, ENABLED
func (s server) sendNotifications(ctx context.Context, controlTypes []ccpb.ControlType, personaId string, setControls bool) {
	for _, controlType := range controlTypes {
		controlFormatted, has := controlFormats[controlType]
		// If the controlType is not in the map, we skip sending notification. This should not occur however.
		if has {
			var preview notification.Preview
			if controlType == ccpb.ControlType_GCT_GLOBAL {
				preview = getGlobalPreview(setControls)
			} else {
				preview = getPreview(controlFormatted, setControls)
			}
			notify := &sdk.NotificationForPersona{
				PersonaID: personaId,
				Notification: notification.Simple{
					ActionURL: "https://plus.anz/cards",
				},
				Preview:        preview,
				IdempotencyKey: uuid.NewString(),
			}
			res, err := s.CommandCentre.Publish(ctx, notify)
			if err != nil {
				log.Error(ctx, err, "Unable to publish Card Controls notification to CommandCentre.")
			} else {
				log.Info(ctx, fmt.Sprintf("Successfully published Card Controls notification to CommandCentre: %v", res.Status))
			}
		}
	}
}

func getPreview(controlFormatted string, setControls bool) notification.Preview {
	verb := "enabled"
	if setControls {
		verb = "disabled"
	}

	return notification.Preview{
		Title: "Card Controls",
		Body:  fmt.Sprintf("We've %s %s with your physical and digital card.", verb, controlFormatted),
	}
}

func getGlobalPreview(setControls bool) notification.Preview {
	if setControls {
		return notification.Preview{
			Title: "Card Locked 🔒",
			Body:  "We've temporarily locked your card. You can go to the Card tab to learn more.",
		}
	} else {
		return notification.Preview{
			Title: "Card Unlocked",
			Body:  "We've unlocked your card. This may take up to 15 minutes to take effect.",
		}
	}
}

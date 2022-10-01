package commandcentre

import (
	"fmt"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
)

type NotificationMatcher struct {
	Notification *sdk.NotificationForPersona
}

func (m *NotificationMatcher) Matches(x interface{}) bool {
	notification, ok := x.(*sdk.NotificationForPersona)
	if !ok {
		return false
	}
	if notification.Notification != m.Notification.Notification {
		return false
	}
	if notification.Preview != m.Notification.Preview {
		return false
	}
	if notification.PersonaID != m.Notification.PersonaID {
		return false
	}
	return true
}

func (m *NotificationMatcher) String() string {
	return fmt.Sprintf("Matches %v objects, ignoring idempotency key and attributes", m.Notification)
}

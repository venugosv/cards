package v1beta2

import (
	"context"
	"testing"

	matchers "github.com/anzx/fabric-cards/pkg/integration/commandcentre/matchers"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/notification"
	"github.com/golang/mock/gomock"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	mock "github.com/anzx/fabric-cards/pkg/integration/commandcentre/mocks"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
)

func TestGetNotificationsForControlTypes(t *testing.T) {
	type args struct {
		controlTypes []ccpb.ControlType
		personaId    string
		setControls  bool
	}
	tests := []struct {
		name   string
		args   args
		titles []string
		bodies []string
	}{
		{
			name: "Single Online Transactions notification (set control)",
			args: args{
				controlTypes: []ccpb.ControlType{ccpb.ControlType_TCT_E_COMMERCE},
				personaId:    "1234",
				setControls:  true,
			},
			titles: []string{"Card Controls"},
			bodies: []string{"We've disabled online transactions with your physical and digital card."},
		},
		{
			name: "Single Online Transactions notification (remove control)",
			args: args{
				controlTypes: []ccpb.ControlType{ccpb.ControlType_TCT_E_COMMERCE},
				personaId:    "1234",
				setControls:  false,
			},
			titles: []string{"Card Controls"},
			bodies: []string{"We've enabled online transactions with your physical and digital card."},
		},
		{
			name: "Unsupported notification (remove control)",
			args: args{
				controlTypes: []ccpb.ControlType{ccpb.ControlType_MCT_ADULT_ENTERTAINMENT},
				personaId:    "1234",
				setControls:  false,
			},
			titles: []string{},
			bodies: []string{},
		},
		{
			name: "Single global transactions notification (set control)",
			args: args{
				controlTypes: []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
				personaId:    "1234",
				setControls:  true,
			},
			titles: []string{"Card Locked 🔒"},
			bodies: []string{"We've temporarily locked your card. You can go to the Card tab to learn more."},
		},
		{
			name: "Single global transactions notification (remove control)",
			args: args{
				controlTypes: []ccpb.ControlType{ccpb.ControlType_GCT_GLOBAL},
				personaId:    "1234",
				setControls:  false,
			},
			titles: []string{"Card Unlocked"},
			bodies: []string{"We've unlocked your card. This may take up to 15 minutes to take effect."},
		},
		{
			name: "Multiple notifications (set control)",
			args: args{
				controlTypes: []ccpb.ControlType{
					ccpb.ControlType_TCT_CONTACTLESS,
					ccpb.ControlType_TCT_ATM_WITHDRAW, ccpb.ControlType_TCT_CROSS_BORDER,
				},
				personaId:   "1234",
				setControls: true,
			},
			titles: []string{"Card Controls", "Card Controls", "Card Controls"},
			bodies: []string{
				"We've disabled contactless payments with your physical and digital card.",
				"We've disabled ATM withdrawals with your physical and digital card.",
				"We've disabled overseas transactions (in-store) with your physical and digital card.",
			},
		},
		{
			name: "Multiple notifications (set control)",
			args: args{
				controlTypes: []ccpb.ControlType{
					ccpb.ControlType_TCT_CONTACTLESS,
					ccpb.ControlType_TCT_ATM_WITHDRAW, ccpb.ControlType_TCT_CROSS_BORDER,
				},
				personaId:   "1234",
				setControls: false,
			},
			titles: []string{"Card Controls", "Card Controls", "Card Controls"},
			bodies: []string{
				"We've enabled contactless payments with your physical and digital card.",
				"We've enabled ATM withdrawals with your physical and digital card.",
				"We've enabled overseas transactions (in-store) with your physical and digital card.",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cc := mock.NewMockPublisher(ctrl)
			for i, title := range test.titles {
				notificationToMatch := &sdk.NotificationForPersona{
					PersonaID: test.args.personaId,
					Notification: notification.Simple{
						ActionURL: "https://plus.anz/cards",
					},
					Preview: notification.Preview{
						Title: title,
						Body:  test.bodies[i],
					},
				}
				cc.EXPECT().Publish(gomock.Any(), &matchers.NotificationMatcher{Notification: notificationToMatch}).Times(1).Return(&sdk.PublishResponse{
					Status: "",
				}, nil)
			}

			s := &server{
				Fabric: Fabric{CommandCentre: &commandcentre.Client{Publisher: cc}},
			}
			s.sendNotifications(context.Background(), test.args.controlTypes, test.args.personaId, test.args.setControls)
		})
	}
}

package feature

import (
	"context"
	"fmt"
	"testing"

	healthpb "github.com/anz-bank/pkg/health/pb"

	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	proto "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestEnabled(t *testing.T) {
	for k, v := range RegisteredRPCs {
		assert.Equal(t, v, RPCGate.Enabled(k))
	}
}

func TestNewFeatureGate(t *testing.T) {
	fg := newFeatureGate(RegisteredRPCs).(*featureGate)
	assert.NotNil(t, fg)
	fm, isFeatureMap := fg.featureMap.Load().(map[Feature]bool)
	assert.Equal(t, true, isFeatureMap)
	assert.Equal(t, len(RegisteredRPCs), len(fm))
}

func TestAddFeatureMap(t *testing.T) {
	tests := []struct {
		description        string
		featureMap         map[Feature]bool
		expectedError      error
		expectedFeatureMap map[Feature]bool
	}{
		{
			description:        "change state of feature",
			featureMap:         map[Feature]bool{CardActivate: true},
			expectedFeatureMap: map[Feature]bool{CardActivate: true},
		},
		{
			description:   "unregistered feature",
			featureMap:    map[Feature]bool{Feature("UnregisteredFeature"): true},
			expectedError: fmt.Errorf("feature is not registered in feature gate: UnregisteredFeature"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// reset feature states to default
			if err := RPCGate.Set(RegisteredRPCs); err != nil {
				assert.FailNow(t, "failed to reset feature map. err: %v. test: %s", err, test.description)
			}
			// set new states of features
			err := RPCGate.Set(test.featureMap)
			// assert
			assert.Equal(t, test.expectedError, err)
			for k, v := range test.expectedFeatureMap {
				assert.Equal(t, v, RPCGate.Enabled(k))
			}
		})
	}
}

// TestAPIFeatureGate
func TestAPIFeatureGate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description      string
		requestGenerator func() interface{}
		fullmethod       string
		features         map[Feature]bool
		expectedError    error
	}{
		{
			description: "List API available",
			requestGenerator: func() interface{} {
				return &proto.ListRequest{}
			},
			fullmethod:    "/fabric.service.card.v1beta1.CardAPI/List",
			features:      map[Feature]bool{CardList: true},
			expectedError: nil,
		},
		{
			description: "List API unavailable",
			requestGenerator: func() interface{} {
				return &proto.ListRequest{}
			},
			fullmethod: "/fabric.service.card.v1beta1.CardAPI/List",
			features:   map[Feature]bool{CardList: false},
			expectedError: anzerrors.New(codes.Unavailable, "method is unavailable",
				anzerrors.NewErrorInfo(context.Background(), anzcodes.FeatureDisabled, "feature disabled")),
		},
		{
			description: "Activate API available",
			requestGenerator: func() interface{} {
				return &proto.ActivateRequest{
					TokenizedCardNumber: "1234567890123456",
					Last_6Digits:        "098765",
				}
			},
			fullmethod:    "/fabric.service.card.v1beta1.CardAPI/Activate",
			features:      map[Feature]bool{CardActivate: true},
			expectedError: nil,
		},
		{
			description: "Activate API available",
			requestGenerator: func() interface{} {
				return &proto.ActivateRequest{
					TokenizedCardNumber: "1234567890123456",
					Last_6Digits:        "098765",
				}
			},
			fullmethod: "/fabric.service.card.v1beta1.CardAPI/Activate",
			features:   map[Feature]bool{CardActivate: false},
			expectedError: anzerrors.New(codes.Unavailable, "method is unavailable",
				anzerrors.NewErrorInfo(context.Background(), anzcodes.FeatureDisabled, "feature disabled")),
		},
		{
			description: "feature disabled",
			requestGenerator: func() interface{} {
				return &proto.ActivateRequest{}
			},
			fullmethod: "/fabric.service.card.v1alpha1.CardAPI/NotAMethod",
			features:   map[Feature]bool{CardActivate: false},
			expectedError: anzerrors.New(codes.Unavailable, "method is unavailable",
				anzerrors.NewErrorInfo(context.Background(), anzcodes.FeatureDisabled, "feature disabled")),
		},
		{
			description: "Unknown request type",
			requestGenerator: func() interface{} {
				return &proto.Card{}
			},
			expectedError: anzerrors.New(codes.Unavailable, "method is unavailable",
				anzerrors.NewErrorInfo(context.Background(), anzcodes.FeatureDisabled, "feature disabled")),
		},
		{
			description: "Alive feature enabled by default",
			requestGenerator: func() interface{} {
				return healthpb.AliveRequest{}
			},
			fullmethod: "/anz.health.v1.Health/Alive",
		},
		{
			description: "Ready feature enabled by default",
			requestGenerator: func() interface{} {
				return healthpb.ReadyRequest{}
			},
			fullmethod: "/anz.health.v1.Health/Ready",
		},
		{
			description: "Version feature enabled by default",
			requestGenerator: func() interface{} {
				return healthpb.VersionRequest{}
			},
			fullmethod: "/anz.health.v1.Health/Version",
		},
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil }

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := RPCGate.Set(test.features)
			if err != nil {
				assert.FailNow(t, "cannot add features for testing: %v", err)
			}
			request := test.requestGenerator()
			_, err = APIFeatureGate()(context.Background(), request, &grpc.UnaryServerInfo{FullMethod: test.fullmethod}, handler)
			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestControlFeatureGate(t *testing.T) {
	rules := []ccpb.ControlType{
		ccpb.ControlType_UNKNOWN_UNSPECIFIED,
		ccpb.ControlType_TCT_ATM_WITHDRAW,
		ccpb.ControlType_TCT_AUTO_PAY,
		ccpb.ControlType_TCT_BRICK_AND_MORTAR,
		ccpb.ControlType_TCT_CROSS_BORDER,
		ccpb.ControlType_TCT_E_COMMERCE,
		ccpb.ControlType_TCT_CONTACTLESS,
		ccpb.ControlType_MCT_ADULT_ENTERTAINMENT,
		ccpb.ControlType_MCT_AIRFARE,
		ccpb.ControlType_MCT_ALCOHOL,
		ccpb.ControlType_MCT_APPAREL_AND_ACCESSORIES,
		ccpb.ControlType_MCT_AUTOMOTIVE,
		ccpb.ControlType_MCT_CAR_RENTAL,
		ccpb.ControlType_MCT_ELECTRONICS,
		ccpb.ControlType_MCT_SPORT_AND_RECREATION,
		ccpb.ControlType_MCT_GAMBLING,
		ccpb.ControlType_MCT_GAS_AND_PETROLEUM,
		ccpb.ControlType_MCT_GROCERY,
		ccpb.ControlType_MCT_HOTEL_AND_LODGING,
		ccpb.ControlType_MCT_HOUSEHOLD,
		ccpb.ControlType_MCT_PERSONAL_CARE,
		ccpb.ControlType_MCT_SMOKE_AND_TOBACCO,
		ccpb.ControlType_GCT_GLOBAL,
	}
	for _, control := range rules {
		t.Run(fmt.Sprintf("%s default", control.String()), func(t *testing.T) {
			assert.False(t, FeatureGate.Enabled(Feature(control.String())))
		})
	}
	controls := map[Feature]bool{
		"TCT_ATM_WITHDRAW": true,
		"TCT_E_COMMERCE":   true,
		"TCT_CONTACTLESS":  true,
		"MCT_GAMBLING":     true,
	}
	if err := FeatureGate.Set(controls); err != nil {
		t.Fatal(err)
	}
	for _, control := range rules {
		t.Run(fmt.Sprintf("%s set", control.String()), func(t *testing.T) {
			feature := Feature(control.String())
			assert.Equal(t, FeatureGate.Enabled(feature), controls[feature])
		})
	}
}

func TestControlsFeatureSet(t *testing.T) {
	rules := map[Feature]bool{
		"UNKNOWN_UNSPECIFIED":            true,
		"TCT_ATM_WITHDRAW":               true,
		"TCT_AUTO_PAY":                   true,
		"TCT_BRICK_AND_MORTAR":           true,
		"TCT_CROSS_BORDER":               true,
		"TCT_E_COMMERCE":                 true,
		"TCT_CONTACTLESS":                true,
		"MCT_ADULT_ENTERTAINMENT":        true,
		"MCT_AIRFARE":                    true,
		"MCT_ALCOHOL":                    true,
		"MCT_APPAREL_AND_ACCESSORIES":    true,
		"MCT_AUTOMOTIVE":                 true,
		"MCT_CAR_RENTAL":                 true,
		"MCT_ELECTRONICS":                true,
		"MCT_SPORT_AND_RECREATION":       true,
		"MCT_GAMBLING":                   true,
		"MCT_GAS_AND_PETROLEUM":          true,
		"MCT_GROCERY":                    true,
		"MCT_HOTEL_AND_LODGING":          true,
		"MCT_HOUSEHOLD":                  true,
		"MCT_PERSONAL_CARE":              true,
		"MCT_SMOKE_AND_TOBACCO":          true,
		"GCT_GLOBAL":                     true,
		"DCVV2":                          true,
		"FORGEROCK_SYSTEM_LOGIN":         true,
		"ENROLLMENT_CALLBACK_INTEGRATED": true,
	}

	if err := FeatureGate.Set(rules); err != nil {
		t.Fatal(err)
	}

	for control := range rules {
		t.Run(fmt.Sprintf("%s set", control), func(t *testing.T) {
			assert.True(t, FeatureGate.Enabled(control))
		})
	}
}

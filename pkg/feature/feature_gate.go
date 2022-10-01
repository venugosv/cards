// Package feature provides functionality to turn a feature or on off using config
package feature

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/pkg/errors"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Config struct {
	RPCs     map[Feature]bool `json:"rpc,omitempty"       yaml:"rpc,omitempty"      mapstructure:"rpc"`
	Features map[Feature]bool `json:"features,omitempty"  yaml:"features,omitempty" mapstructure:"features"`
}

// Feature is a string representing the name of the feature to be used for feature gating
type Feature string

const (
	// RPCs
	HealthAlive                    Feature = "/anz.health.v1.health/alive"
	HealthReady                    Feature = "/anz.health.v1.health/ready"
	HealthVersion                  Feature = "/anz.health.v1.health/version"
	CardActivate                   Feature = "/fabric.service.card.v1beta1.cardapi/activate"
	CardAuditTrail                 Feature = "/fabric.service.card.v1beta1.cardapi/audittrail"
	CardChangePin                  Feature = "/fabric.service.card.v1beta1.cardapi/changepin"
	CardGetDetails                 Feature = "/fabric.service.card.v1beta1.cardapi/getdetails"
	CardGetWrappingKey             Feature = "/fabric.service.card.v1beta1.cardapi/getwrappingkey"
	CardList                       Feature = "/fabric.service.card.v1beta1.cardapi/list"
	CardReplace                    Feature = "/fabric.service.card.v1beta1.cardapi/replace"
	CardSetPin                     Feature = "/fabric.service.card.v1beta1.cardapi/setpin"
	CardVerifyPin                  Feature = "/fabric.service.card.v1beta1.cardapi/verifypin"
	CardResetPin                   Feature = "/fabric.service.card.v1beta1.cardapi/resetpin"
	WalletCreateApplePaymentToken  Feature = "/fabric.service.card.v1beta1.walletapi/createapplepaymenttoken"
	WalletCreateGooglePaymentToken Feature = "/fabric.service.card.v1beta1.walletapi/creategooglepaymenttoken"
	EligibilityCan                 Feature = "/fabric.service.eligibility.v1beta1.cardeligibilityapi/can"
	ControlV1beta1Block            Feature = "/fabric.service.cardcontrols.v1beta1.cardcontrolsapi/block"
	ControlV1beta1List             Feature = "/fabric.service.cardcontrols.v1beta1.cardcontrolsapi/list"
	ControlV1beta1Query            Feature = "/fabric.service.cardcontrols.v1beta1.cardcontrolsapi/query"
	ControlV1beta1Remove           Feature = "/fabric.service.cardcontrols.v1beta1.cardcontrolsapi/remove"
	ControlV1beta1Set              Feature = "/fabric.service.cardcontrols.v1beta1.cardcontrolsapi/set"
	ControlV1beta2Block            Feature = "/fabric.service.cardcontrols.v1beta2.cardcontrolsapi/blockcard"
	ControlV1beta2List             Feature = "/fabric.service.cardcontrols.v1beta2.cardcontrolsapi/listcontrols"
	ControlV1beta2Query            Feature = "/fabric.service.cardcontrols.v1beta2.cardcontrolsapi/querycontrols"
	ControlV1beta2Remove           Feature = "/fabric.service.cardcontrols.v1beta2.cardcontrolsapi/removecontrols"
	ControlV1beta2Set              Feature = "/fabric.service.cardcontrols.v1beta2.cardcontrolsapi/setcontrols"
	ControlV1beta2Transfer         Feature = "/fabric.service.cardcontrols.v1beta2.cardcontrolsapi/transfercontrols"
	CallbackEnroll                 Feature = "/visa.service.enrollmentcallback.v1.enrollmentcallbackapi/enroll"
	CallbackDisenroll              Feature = "/visa.service.enrollmentcallback.v1.enrollmentcallbackapi/disenroll"
	CallbackAlert                  Feature = "/visa.service.notificationcallback.v1.notificationcallbackapi/alert"
)

// RegisteredRPCs defines a list of registered features and only these features which state can be changed
var RegisteredRPCs = map[Feature]bool{
	HealthAlive:                    true,
	HealthReady:                    true,
	HealthVersion:                  true,
	CardActivate:                   false,
	CardAuditTrail:                 false,
	CardChangePin:                  false,
	CardGetDetails:                 false,
	CardGetWrappingKey:             false,
	CardList:                       false,
	CardReplace:                    false,
	CardSetPin:                     false,
	CardVerifyPin:                  false,
	CardResetPin:                   false,
	WalletCreateApplePaymentToken:  false,
	WalletCreateGooglePaymentToken: false,
	EligibilityCan:                 false,
	ControlV1beta1Block:            false,
	ControlV1beta1List:             false,
	ControlV1beta1Query:            false,
	ControlV1beta1Remove:           false,
	ControlV1beta1Set:              false,
	ControlV1beta2Block:            false,
	ControlV1beta2List:             false,
	ControlV1beta2Query:            false,
	ControlV1beta2Remove:           false,
	ControlV1beta2Set:              false,
	ControlV1beta2Transfer:         false,
	CallbackEnroll:                 false,
	CallbackDisenroll:              false,
	CallbackAlert:                  false,
}

const (
	// Features
	UNKNOWN_UNSPECIFIED               Feature = "UNKNOWN_UNSPECIFIED"
	TCT_ATM_WITHDRAW                  Feature = "TCT_ATM_WITHDRAW"
	TCT_AUTO_PAY                      Feature = "TCT_AUTO_PAY"
	TCT_BRICK_AND_MORTAR              Feature = "TCT_BRICK_AND_MORTAR"
	TCT_CROSS_BORDER                  Feature = "TCT_CROSS_BORDER"
	TCT_E_COMMERCE                    Feature = "TCT_E_COMMERCE"
	TCT_CONTACTLESS                   Feature = "TCT_CONTACTLESS"
	MCT_ADULT_ENTERTAINMENT           Feature = "MCT_ADULT_ENTERTAINMENT"
	MCT_AIRFARE                       Feature = "MCT_AIRFARE"
	MCT_ALCOHOL                       Feature = "MCT_ALCOHOL"
	MCT_APPAREL_AND_ACCESSORIES       Feature = "MCT_APPAREL_AND_ACCESSORIES"
	MCT_AUTOMOTIVE                    Feature = "MCT_AUTOMOTIVE"
	MCT_CAR_RENTAL                    Feature = "MCT_CAR_RENTAL"
	MCT_ELECTRONICS                   Feature = "MCT_ELECTRONICS"
	MCT_SPORT_AND_RECREATION          Feature = "MCT_SPORT_AND_RECREATION"
	MCT_GAMBLING                      Feature = "MCT_GAMBLING"
	MCT_GAS_AND_PETROLEUM             Feature = "MCT_GAS_AND_PETROLEUM"
	MCT_GROCERY                       Feature = "MCT_GROCERY"
	MCT_HOTEL_AND_LODGING             Feature = "MCT_HOTEL_AND_LODGING"
	MCT_HOUSEHOLD                     Feature = "MCT_HOUSEHOLD"
	MCT_PERSONAL_CARE                 Feature = "MCT_PERSONAL_CARE"
	MCT_SMOKE_AND_TOBACCO             Feature = "MCT_SMOKE_AND_TOBACCO"
	GCT_GLOBAL                        Feature = "GCT_GLOBAL"
	REASON_UNKNOWN_UNSPECIFIED        Feature = "REASON_UNKNOWN_UNSPECIFIED"
	REASON_LOST                       Feature = "REASON_LOST"
	REASON_STOLEN                     Feature = "REASON_STOLEN"
	REASON_DAMAGED                    Feature = "REASON_DAMAGED"
	DCVV2                             Feature = "DCVV2"
	FORGEROCK_SYSTEM_LOGIN            Feature = "FORGEROCK_SYSTEM_LOGIN"
	ENROLLMENT_CALLBACK_INTEGRATED    Feature = "ENROLLMENT_CALLBACK_INTEGRATED"
	PIN_CHANGE_COUNT                  Feature = "PIN_CHANGE_COUNT"
	NotificationCallbackDeclinedEvent Feature = "NOTIFICATION_CALLBACK_DECLINED_EVENT"
)

// RegisteredFeatures defines a list of registered Card Features and only these features which state can be changed
var RegisteredFeatures = map[Feature]bool{
	UNKNOWN_UNSPECIFIED:               false,
	TCT_ATM_WITHDRAW:                  false,
	TCT_AUTO_PAY:                      false,
	TCT_BRICK_AND_MORTAR:              false,
	TCT_CROSS_BORDER:                  false,
	TCT_E_COMMERCE:                    false,
	TCT_CONTACTLESS:                   false,
	MCT_ADULT_ENTERTAINMENT:           false,
	MCT_AIRFARE:                       false,
	MCT_ALCOHOL:                       false,
	MCT_APPAREL_AND_ACCESSORIES:       false,
	MCT_AUTOMOTIVE:                    false,
	MCT_CAR_RENTAL:                    false,
	MCT_ELECTRONICS:                   false,
	MCT_SPORT_AND_RECREATION:          false,
	MCT_GAMBLING:                      false,
	MCT_GAS_AND_PETROLEUM:             false,
	MCT_GROCERY:                       false,
	MCT_HOTEL_AND_LODGING:             false,
	MCT_HOUSEHOLD:                     false,
	MCT_PERSONAL_CARE:                 false,
	MCT_SMOKE_AND_TOBACCO:             false,
	GCT_GLOBAL:                        false,
	REASON_UNKNOWN_UNSPECIFIED:        false,
	REASON_LOST:                       false,
	REASON_STOLEN:                     false,
	REASON_DAMAGED:                    false,
	DCVV2:                             false,
	FORGEROCK_SYSTEM_LOGIN:            false,
	ENROLLMENT_CALLBACK_INTEGRATED:    false,
	PIN_CHANGE_COUNT:                  false,
	NotificationCallbackDeclinedEvent: false,
}

var (
	// RPCGate gives direct access to current feature gate and is initialised with RegisteredRPCs
	RPCGate = newFeatureGate(RegisteredRPCs)
	// FeatureGate gives direct access to current feature gate and is initialised with RegisteredFeatures
	FeatureGate = newFeatureGate(RegisteredFeatures)
)

// Gate indicates whether a given feature is enabled or not.
type Gate interface {
	// Enabled returns true if the key is enabled
	Enabled(key Feature) bool

	// Add features to the feature map
	Set(features map[Feature]bool) error
}

// featureGate implements Gate.
type featureGate struct {
	// featureMap holds a map[RPCs]bool.
	featureMap *atomic.Value

	// lock guards writes to known, enabled, and reads/writes of closed
	// currently realtime change to the featuremap is not available however this is needed for parallel testing
	lock sync.Mutex
}

// Enabled returns true if the key is enabled.
// If the key is not known, this will return false
func (fg *featureGate) Enabled(key Feature) bool {
	if v, ok := fg.featureMap.Load().(map[Feature]bool)[key]; ok {
		return v
	}
	return false
}

// newFeatureGate initialises Gate with default feature map and returns its instance
func newFeatureGate(features map[Feature]bool) Gate {
	fm := map[Feature]bool{}
	for k, v := range features {
		fm[k] = v
	}

	mapValue := &atomic.Value{}
	mapValue.Store(fm)

	fg := &featureGate{featureMap: mapValue}
	return fg
}

// Set adds features
func (fg *featureGate) Set(features map[Feature]bool) error {
	fg.lock.Lock()
	defer fg.lock.Unlock()

	fm := map[Feature]bool{}
	// Copy state
	for k, v := range fg.featureMap.Load().(map[Feature]bool) {
		fm[k] = v
	}

	// Set feature from input
	for k, v := range features {
		if _, ok := fm[k]; !ok {
			return fmt.Errorf("feature is not registered in feature gate: %s", k)
		}
		fm[k] = v
	}

	// Persist changes
	fg.featureMap.Store(fm)

	return nil
}

// APIFeatureGate to check whether feature has been enabled via the feature map from the config file.
// It returns codes.Unavailable if the feature is not defined or enabled.
func APIFeatureGate() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		feature := Feature(strings.ToLower(info.FullMethod))
		if !RPCGate.Enabled(feature) {
			return nil, anzerrors.New(codes.Unavailable, "method is unavailable",
				anzerrors.NewErrorInfo(ctx, anzcodes.FeatureDisabled, "feature disabled"),
				anzerrors.WithCause(errors.New("method is behind feature toggle")))
		}

		return handler(ctx, req)
	}
}

package customerrules

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/util"

	l "log"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"

	"github.com/anzx/fabric-cards/pkg/feature"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"
)

const (
	controlNotAllowed   = "control not allowed"
	behindFeatureToggle = "%s is behind feature toggle"
	noTimeRemaining     = "00:00:00"
)

// Enrolled checks if the current resource is enrolled
func Enrolled(d *crpb.Resource) bool {
	return d.DocumentId != "NOT_ENROLLED"
}

func GetDeleteRequest(d *crpb.Resource, controlTypes []ccpb.ControlType) (*crpb.ControlRequest, bool) {
	deleteRequest := &crpb.ControlRequest{}
	for _, controlType := range controlTypes {
		i, ok := FindControlByType(d, controlType)
		if !ok {
			continue
		}
		switch GetCategory(controlType) {
		case GLOBAL:
			deleteRequest.GlobalControls = append(deleteRequest.GlobalControls, d.GlobalControls...)
		case TRANSACTION:
			deleteRequest.TransactionControls = append(deleteRequest.TransactionControls, d.TransactionControls[i])
		case MERCHANT:
			deleteRequest.MerchantControls = append(deleteRequest.MerchantControls, d.MerchantControls[i])
		}
	}

	if deleteRequest.GlobalControls == nil && deleteRequest.TransactionControls == nil && deleteRequest.MerchantControls == nil {
		return nil, false
	}

	return deleteRequest, true
}

func FindControlByType(d *crpb.Resource, controlType ccpb.ControlType) (int, bool) {
	switch GetCategory(controlType) {
	case GLOBAL:
		for i, control := range d.GlobalControls {
			if control.IsControlEnabled {
				return i, true
			}
		}
	case TRANSACTION:
		for i, control := range d.TransactionControls {
			if control.ControlType == controlType.String() && control.IsControlEnabled {
				return i, true
			}
		}
	case MERCHANT:
		for i, control := range d.MerchantControls {
			if control.ControlType == controlType.String() && control.IsControlEnabled {
				return i, true
			}
		}
	}

	return -1, false
}

func ControlRequest(requestBuilders ...func(*crpb.ControlRequest)) *crpb.ControlRequest {
	request := &crpb.ControlRequest{
		GlobalControls:      []*crpb.GlobalControl{},
		MerchantControls:    []*crpb.MerchantControl{},
		TransactionControls: []*crpb.TransactionControl{},
	}

	for _, build := range requestBuilders {
		build(request)
	}

	return request
}

func WithControls(ctx context.Context, controlRequest []*ccpb.ControlRequest, id string) ([]func(*crpb.ControlRequest), error) {
	var out []func(*crpb.ControlRequest)

	for _, request := range controlRequest {
		if !feature.FeatureGate.Enabled(feature.Feature(request.ControlType.String())) {
			return nil, anzerrors.New(codes.Unavailable, controlNotAllowed,
				anzerrors.NewErrorInfo(ctx, anzcodes.FeatureDisabled, "control is disabled"),
				anzerrors.WithCause(fmt.Errorf(behindFeatureToggle, request.String())))
		}
		out = append(out, addRequest(request, id))
	}

	return out, nil
}

func addRequest(request *ccpb.ControlRequest, id string) func(*crpb.ControlRequest) {
	var out func(*crpb.ControlRequest)

	switch GetCategory(request.ControlType) {
	case GLOBAL:
		out = globalControlRequest(request, id)
	case MERCHANT:
		out = merchantControlRequest(request, id)
	case TRANSACTION:
		out = transactionControlRequest(request, id)
	}

	return out
}

func transactionControlRequest(request *ccpb.ControlRequest, id string) func(*crpb.ControlRequest) {
	control := crpb.TransactionControl{
		ShouldDeclineAll:     util.ToBoolPtr(true),
		ShouldAlertOnDecline: util.ToBoolPtr(true),
		IsControlEnabled:     true,
		ControlType:          request.ControlType.String(),
		UserIdentifier:       &id,
	}
	return func(r *crpb.ControlRequest) {
		r.TransactionControls = append(r.TransactionControls, &control)
	}
}

func merchantControlRequest(request *ccpb.ControlRequest, id string) func(*crpb.ControlRequest) {
	control := crpb.MerchantControl{
		ShouldDeclineAll:     util.ToBoolPtr(true),
		ShouldAlertOnDecline: util.ToBoolPtr(true),
		IsControlEnabled:     true,
		ControlType:          request.ControlType.String(),
		UserIdentifier:       &id,
	}
	return func(r *crpb.ControlRequest) {
		r.MerchantControls = append(r.MerchantControls, &control)
	}
}

func globalControlRequest(_ *ccpb.ControlRequest, id string) func(*crpb.ControlRequest) {
	control := crpb.GlobalControl{
		ShouldAlertOnDecline:              util.ToBoolPtr(true),
		IsControlEnabled:                  true,
		UserIdentifier:                    &id,
		DeclineAllNonTokenizeTransactions: util.ToBoolPtr(true),
	}
	return func(r *crpb.ControlRequest) {
		r.GlobalControls = append(r.GlobalControls, &control)
	}
}

type Category int32

const (
	TRANSACTION Category = iota
	MERCHANT
	GLOBAL
)

func (c Category) String() string {
	switch c {
	case TRANSACTION:
		return "TRANSACTION"
	case MERCHANT:
		return "MERCHANT"
	case GLOBAL:
		return "GLOBAL"
	default:
		return ""
	}
}

func GetCategory(c ccpb.ControlType) Category {
	switch c {
	case ccpb.ControlType_TCT_ATM_WITHDRAW, ccpb.ControlType_TCT_AUTO_PAY, ccpb.ControlType_TCT_BRICK_AND_MORTAR,
		ccpb.ControlType_TCT_CROSS_BORDER, ccpb.ControlType_TCT_E_COMMERCE, ccpb.ControlType_TCT_CONTACTLESS:
		return TRANSACTION
	case ccpb.ControlType_MCT_ADULT_ENTERTAINMENT, ccpb.ControlType_MCT_AIRFARE, ccpb.ControlType_MCT_ALCOHOL,
		ccpb.ControlType_MCT_APPAREL_AND_ACCESSORIES, ccpb.ControlType_MCT_AUTOMOTIVE, ccpb.ControlType_MCT_CAR_RENTAL,
		ccpb.ControlType_MCT_ELECTRONICS, ccpb.ControlType_MCT_SPORT_AND_RECREATION, ccpb.ControlType_MCT_GAMBLING,
		ccpb.ControlType_MCT_GAS_AND_PETROLEUM, ccpb.ControlType_MCT_GROCERY, ccpb.ControlType_MCT_HOTEL_AND_LODGING,
		ccpb.ControlType_MCT_HOUSEHOLD, ccpb.ControlType_MCT_PERSONAL_CARE, ccpb.ControlType_MCT_SMOKE_AND_TOBACCO:
		return MERCHANT
	case ccpb.ControlType_GCT_GLOBAL:
		return GLOBAL
	default:
		return 0
	}
}

func GetImpulseDelayStartTimestamp(c *crpb.MerchantControl) *timestamppb.Timestamp {
	if c.GetImpulseDelayStart() == "" {
		l.Printf("impulse delay start time retrieved when nil")
		return nil
	}
	impulseDelayStart := strings.ReplaceAll(c.GetImpulseDelayStart(), "-", "/")
	delayStartTime, err := time.Parse("2006/01/02 15:04:05", impulseDelayStart)
	if err != nil {
		l.Printf("unable to parse impulse delay start time %v", err)
		return nil
	}
	return timestamppb.New(delayStartTime)
}

func GetImpulseDelayPeriodProto(c *crpb.MerchantControl) *durationpb.Duration {
	if c.ImpulseDelayPeriod == nil {
		l.Printf("impulse delay period retrieved when nil")
		return nil
	}
	impulseDelayPeriod := strings.ReplaceAll(*c.ImpulseDelayPeriod, ":", "h")
	delayPeriodTime, err := time.ParseDuration(fmt.Sprintf("%sm", impulseDelayPeriod))
	if err != nil {
		l.Printf("unable to parse impulse delay period %v", err)
		return nil
	}
	return durationpb.New(delayPeriodTime)
}

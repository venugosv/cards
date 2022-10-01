package visa

import (
	"context"
	"fmt"

	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-cards/pkg/integration/util"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"
)

const (
	controlNotAllowed   = "control not allowed"
	behindFeatureToggle = "%s is behind feature toggle"
)

// Enrolled checks if the current resource is enrolled
func (d Resource) Enrolled() bool {
	return d.DocumentID != "NOT_ENROLLED"
}

func (d *Resource) RemoveControlByType(controlType ccpb.ControlType) bool {
	i, ok := d.FindControlByType(controlType)
	if !ok {
		return false
	}
	switch GetCategory(controlType) {
	case GLOBAL:
		d.GlobalControls = []*GlobalControl{}
	case TRANSACTION:
		d.removeTransactionControl(i)
	case MERCHANT:
		if controlType == ccpb.ControlType_MCT_GAMBLING {
			return d.removeGamblingControl(i)
		}
		d.removeMerchantControl(i)
	}
	return true
}

func (d *Resource) removeTransactionControl(index int) {
	d.TransactionControls = append(d.TransactionControls[:index], d.TransactionControls[index+1:]...)
}

func (d *Resource) removeMerchantControl(index int) {
	d.MerchantControls = append(d.MerchantControls[:index], d.MerchantControls[index+1:]...)
}

func (d *Resource) removeGamblingControl(index int) bool {
	// get control
	control := d.MerchantControls[index]
	// check its a gambling control
	if control.ControlType != ccpb.ControlType_MCT_GAMBLING.String() {
		return false
	}
	// if impulse doesnt exist, add it, return true
	if control.impulseDelayExists() {
		// if it does exist, validate enforcement
		if control.enforceImpulseDelay() {
			return false
		}
		// if it has expired, remove control, return true,
		d.removeMerchantControl(index)
		return true
	}

	control.setImpulseDelay()

	return true
}

func (d Resource) FindControlByType(controlType ccpb.ControlType) (int, bool) {
	switch GetCategory(controlType) {
	case GLOBAL:
		for i, control := range d.GlobalControls {
			if control.ControlEnabled {
				return i, true
			}
		}
	case TRANSACTION:
		for i, control := range d.TransactionControls {
			if control.ControlType == controlType.String() && control.ControlEnabled {
				return i, true
			}
		}
	case MERCHANT:
		for i, control := range d.MerchantControls {
			if control.ControlType == controlType.String() && control.ControlEnabled {
				return i, true
			}
		}
	}

	return -1, false
}

// MerchantControl Methods
func (c *MerchantControl) enforceImpulseDelay() bool {
	return c.ImpulseDelayRemaining != nil && *c.ImpulseDelayRemaining != "00:00:00"
}

func (c *MerchantControl) impulseDelayExists() bool {
	return c.ImpulseDelayStart != nil &&
		c.ImpulseDelayRemaining != nil &&
		c.ImpulseDelayEnd != nil &&
		c.ImpulseDelayPeriod != nil
}

func (c *MerchantControl) setImpulseDelay() *MerchantControl {
	const fortyEightHours = "48:00"
	c.ImpulseDelayPeriod = util.ToStringPtr(fortyEightHours)
	return c
}

func ControlRequest(requestBuilders ...func(*Request)) *Request {
	request := &Request{
		GlobalControls:      []*GlobalControl{},
		MerchantControls:    []*MerchantControl{},
		TransactionControls: []*TransactionControl{},
	}

	for _, build := range requestBuilders {
		build(request)
	}

	return request
}

func WithControls(ctx context.Context, controlRequest []*ccpb.ControlRequest, id string) ([]func(*Request), error) {
	var out []func(*Request)

	for _, request := range controlRequest {
		if !feature.FeatureGate.Enabled(feature.Feature(request.ControlType.String())) {
			return nil, anzerrors.New(codes.Unavailable, controlNotAllowed,
				anzerrors.NewErrorInfo(ctx, anzcodes.FeatureDisabled, "control is disabled"),
				anzerrors.WithCause(fmt.Errorf(behindFeatureToggle, request.String())))
		}
		out = append(out, AddRequest(request, id))
	}

	return out, nil
}

func AddRequest(request *ccpb.ControlRequest, id string) func(*Request) {
	var out func(*Request)

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

func transactionControlRequest(request *ccpb.ControlRequest, id string) func(*Request) {
	control := TransactionControl{
		ShouldDeclineAll:     true,
		ControlEnabled:       true,
		ControlType:          request.ControlType.String(),
		UserIdentifier:       id,
		AlertThreshold:       util.ToFloat64Ptr(15),
		DeclineThreshold:     util.ToFloat64Ptr(0),
		ShouldAlertOnDecline: false,
	}
	return func(r *Request) {
		r.TransactionControls = append(r.TransactionControls, &control)
	}
}

func merchantControlRequest(request *ccpb.ControlRequest, id string) func(*Request) {
	control := MerchantControl{
		ShouldDeclineAll:     true,
		ControlEnabled:       true,
		UserIdentifier:       id,
		ControlType:          request.ControlType.String(),
		AlertThreshold:       util.ToFloat64Ptr(15),
		DeclineThreshold:     util.ToFloat64Ptr(0),
		ShouldAlertOnDecline: false,
	}
	return func(r *Request) {
		r.MerchantControls = append(r.MerchantControls, &control)
	}
}

func globalControlRequest(_ *ccpb.ControlRequest, id string) func(*Request) {
	control := GlobalControl{
		ShouldDeclineAll:                  true,
		ControlEnabled:                    true,
		UserIdentifier:                    id,
		AlertThreshold:                    util.ToFloat64Ptr(15),
		DeclineThreshold:                  util.ToFloat64Ptr(0),
		DeclineAllNonTokenizeTransactions: true,
		ShouldAlertOnDecline:              false,
	}
	return func(r *Request) {
		r.GlobalControls = append(r.GlobalControls, &control)
	}
}

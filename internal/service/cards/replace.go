package cards

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/anzx/pkg/xcontext"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/notification"
	"github.com/anzx/pkg/log"
	"github.com/google/uuid"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/integration/selfservice"

	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"

	"github.com/anzx/fabric-cards/pkg/date"

	"github.com/anzx/fabric-cards/pkg/feature"

	"golang.org/x/sync/errgroup"

	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	"github.com/anzx/pkg/auditlog"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-commandcentre-sdk/pkg/sdk/event"

	"github.com/anzx/fabric-cards/pkg/integration/ocv"

	"github.com/anzx/fabric-cards/pkg/identity"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"

	"google.golang.org/grpc/codes"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"

	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
)

const (
	replacementFailed = "replacement failed"
	MaxEmbossedName   = 21
	MaxNameLength     = 20 // Max length of FirstName & lastname allowed by CTM
	MaxCtmNameLength  = 24 // using 24 as max combination of title + firstname + lastname WITHOUT spaces (26 with spaces)
)

func (s server) Replace(ctx context.Context, req *cpb.ReplaceRequest) (retResponse *cpb.ReplaceResponse, retError error) {
	serviceData := &servicedata.ReplaceCard{
		TokenizedCardNumber: req.TokenizedCardNumber,
		NewMediaType:        "N",
		OldMediaType:        "N",
		NewIssueReason:      req.Reason.String(),
	}

	defer func() {
		if err := serviceData.Validate(); err != nil {
			logf.Error(ctx, err, "invalid service data payload")
		}
		s.AuditLog.Publish(ctx, auditlog.EventReplaceCard, retResponse, retError, serviceData)
	}()

	if !feature.FeatureGate.Enabled(feature.Feature(req.Reason.String())) {
		return nil, anzerrors.New(codes.Unavailable, "reason not allowed",
			anzerrors.NewErrorInfo(ctx, anzcodes.FeatureDisabled, "reason is disabled"),
			anzerrors.WithCause(fmt.Errorf("%s is behind feature toggle", req.Reason.String())))
	}

	entitledCard, err := s.Entitlements.GetEntitledCard(ctx, req.TokenizedCardNumber, entitlements.OPERATION_MANAGE_CARD)
	if err != nil {
		return nil, serviceErr(err, replacementFailed)
	}

	if len(entitledCard.GetAccountNumbers()) == 0 {
		return nil, anzerrors.New(codes.Internal, "Invalid response from GetEntitledCard", anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "Card has no account numbers"))
	}

	serviceData.AccountNumbers = entitledCard.GetAccountNumbers()

	if err := s.Eligibility.Can(ctx, getEligibility(req.Reason), req.TokenizedCardNumber); err != nil {
		return nil, serviceErr(err, replacementFailed)
	}

	id, err := identity.Get(ctx)
	if err != nil {
		return nil, serviceErr(err, replacementFailed)
	}

	oldCard, accounts, party, err := s.preamble(ctx, entitledCard, req.GetReason(), id.OcvID)
	if err != nil {
		return nil, serviceErr(err, replacementFailed)
	}

	newCard, oldCard, err := s.requestReplacementCard(ctx, oldCard, req.Reason, party)
	if err != nil {
		return nil, err
	}

	defer func() {
		s.CommandCentre.PublishEventAsync(ctx, event.CardStatusChange)
		if id.HasDifferentSubject {
			// This request was likely made by a staff member or coach on customer's behalf, so we should notify customer
			go s.publishNotification(xcontext.Detach(ctx), id.PersonaID)
		}
		populateServiceData(ctx, serviceData, oldCard, newCard)
	}()

	if plasticType(req.Reason) == ctm.NewNumber {
		if err = s.processNewCard(ctx, id.OcvID, accounts, oldCard, newCard); err != nil {
			return nil, err
		}
	}

	return &cpb.ReplaceResponse{
		NewTokenizedCardNumber: newCard.CardNumber.Token,
		Eligibilities:          newCard.Eligibility(),
	}, nil
}

func (s server) processNewCard(ctx context.Context, ocvID string, accounts []*ocv.RetrievePartyRsAccount, oldCard, newCard *ctm.DebitCardResponse) error {
	g, gctx := errgroup.WithContext(ctx)
	for _, account := range accounts {
		a := account
		g.Go(func() error {
			if ok, err := s.OCV.AccountMaintenance(gctx, ocvID, oldCard, newCard, a); !ok {
				return serviceErr(err, replacementFailed)
			}
			return nil
		})
	}
	g.Go(func() error {
		if err := s.Entitlements.Register(gctx, newCard.CardNumber.Token); err != nil {
			return serviceErr(err, replacementFailed)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if err := s.Entitlements.Latest(ctx); err != nil {
		return serviceErr(err, replacementFailed)
	}

	if oldCard.CardControlPreference {
		if err := s.CardControls.TransferControls(ctx, oldCard.CardNumber.Token, newCard.CardNumber.Token); err != nil {
			logf.Error(ctx, err, "new card has been created but failed to transfer controls")
			return serviceErr(err, replacementFailed)
		}
	}

	return nil
}

func (s server) preamble(ctx context.Context, entitledCard *entpb.EntitledCard, reason cpb.ReplaceRequest_Reason, ocvID string) (*ctm.DebitCardResponse, []*ocv.RetrievePartyRsAccount, *selfservice.Party, error) {
	var (
		parties []*ocv.RetrievePartyRs
		oldCard *ctm.DebitCardResponse
		party   *selfservice.Party
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() (err error) {
		oldCard, err = s.CTM.DebitCardInquiry(gctx, entitledCard.GetTokenizedCardNumber())
		if err != nil {
			return serviceErr(err, replacementFailed)
		}
		return nil
	})

	g.Go(func() (err error) {
		party, err = s.SelfService.GetParty(gctx)
		if err != nil {
			return serviceErr(err, replacementFailed)
		}
		return nil
	})

	g.Go(func() (err error) {
		parties, err = s.OCV.RetrieveParty(gctx, ocvID)
		if err != nil {
			return serviceErr(err, replacementFailed)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}

	accounts, err := ocv.GetAccounts(parties, entitledCard.GetAccountNumbers())
	if err != nil {
		return nil, nil, nil, serviceErr(err, replacementFailed)
	}

	cardAccount, err := ocv.GetAccount(parties, fmt.Sprintf("enc(%s)", oldCard.CardNumber.Token))
	if err != nil {
		return nil, nil, nil, serviceErr(err, replacementFailed)
	}

	if cardAccount.IssuedToday() && reason != cpb.ReplaceRequest_REASON_DAMAGED {
		return nil, nil, nil, anzerrors.New(codes.InvalidArgument, replacementFailed,
			anzerrors.NewErrorInfo(gctx, anzcodes.CardSameDayReplacement, "cannot replace card number on the same day it was created"))
	}

	return oldCard, accounts, party, nil
}

func (s server) requestReplacementCard(ctx context.Context, currentCard *ctm.DebitCardResponse, reason cpb.ReplaceRequest_Reason, party *selfservice.Party) (*ctm.DebitCardResponse, *ctm.DebitCardResponse, error) {
	// if a card is already replaced, skip the part for getting a new card and just try to link the card to customer/account
	var newTokenizedCardNumber string
	switch {
	case currentCard.NewCardNumber != nil:
		newTokenizedCardNumber = currentCard.NewCardNumber.Token
	default:
		var err error
		newTokenizedCardNumber, err = s.replaceCard(ctx, currentCard, reason, party)
		if err != nil {
			return nil, nil, err
		}
	}

	var oldCard, newCard *ctm.DebitCardResponse

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() (err error) {
		newCard, err = s.CTM.DebitCardInquiry(gctx, newTokenizedCardNumber)
		if err != nil {
			return serviceErr(err, replacementFailed)
		}
		return nil
	})

	g.Go(func() (err error) {
		oldCard, err = s.CTM.DebitCardInquiry(gctx, currentCard.CardNumber.Token)
		if err != nil {
			return serviceErr(err, replacementFailed)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	return newCard, oldCard, nil
}

func (s server) replaceCard(ctx context.Context, currentCard *ctm.DebitCardResponse, reason cpb.ReplaceRequest_Reason, party *selfservice.Party) (string, error) {
	address, err := party.GetAddress(ctx)
	if err != nil {
		return "", serviceErr(err, replacementFailed)
	}

	mailingAddress, err := ctm.GetAddress(ctx, address)
	if err != nil {
		return "", err
	}

	if err = s.updateStatus(ctx, currentCard, reason); err != nil {
		return "", err
	}

	firstName, lastName := toCTMName(party.LegalName.FirstName, party.LegalName.LastName)

	embossedName := toEmbossedName(firstName, lastName)

	request := &ctm.ReplaceCardRequest{
		PlasticType:              plasticType(reason),
		FirstName:                firstName,
		LastName:                 lastName,
		EmbossingLine1:           embossedName,
		DispatchedMethod:         ctm.DispatchedMethodMail,
		DesignCode:               currentCard.DesignCode,
		MerchantUpdatePreference: currentCard.MerchantUpdatePreference,
		MailingAddress:           mailingAddress,
	}

	newTokenizedCardNumber, err := s.CTM.ReplaceCard(ctx, request, currentCard.CardNumber.Token)
	if err != nil {
		return "", serviceErr(err, replacementFailed)
	}

	return newTokenizedCardNumber, nil
}

func (s server) updateStatus(ctx context.Context, currentCard *ctm.DebitCardResponse, reason cpb.ReplaceRequest_Reason) error {
	if currentCard.Status == status(reason) {
		return nil
	}

	if !currentCard.Status.ValidNextState(status(reason)) {
		return anzerrors.New(codes.PermissionDenied, replacementFailed,
			anzerrors.NewErrorInfo(ctx, anzcodes.CardStatusNotAvailable, "card ineligible for replacement"))
	}

	if ok, err := s.CTM.UpdateStatus(ctx, currentCard.CardNumber.Token, status(reason)); !ok {
		return serviceErr(err, replacementFailed)
	}

	return nil
}

func getEligibility(in cpb.ReplaceRequest_Reason) epb.Eligibility {
	switch in {
	case cpb.ReplaceRequest_REASON_LOST:
		return epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_LOST
	case cpb.ReplaceRequest_REASON_STOLEN:
		return epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_STOLEN
	case cpb.ReplaceRequest_REASON_DAMAGED:
		return epb.Eligibility_ELIGIBILITY_CARD_REPLACEMENT_DAMAGED
	default:
		return epb.Eligibility_ELIGIBILITY_INVALID_UNSPECIFIED
	}
}

func GetExpiry(ctx context.Context, in string) string {
	d := date.GetDate(ctx, date.YYMM, in)
	return fmt.Sprintf("%d%02d01", d.GetYear().GetValue(), d.GetMonth().GetValue())
}

func plasticType(in cpb.ReplaceRequest_Reason) ctm.PlasticType {
	switch in {
	case cpb.ReplaceRequest_REASON_LOST:
		return ctm.NewNumber
	case cpb.ReplaceRequest_REASON_STOLEN:
		return ctm.NewNumber
	case cpb.ReplaceRequest_REASON_DAMAGED:
		return ctm.SameNumber
	default:
		return ""
	}
}

func status(in cpb.ReplaceRequest_Reason) ctm.Status {
	switch in {
	case cpb.ReplaceRequest_REASON_LOST:
		return ctm.StatusLost
	case cpb.ReplaceRequest_REASON_STOLEN:
		return ctm.StatusStolen
	case cpb.ReplaceRequest_REASON_DAMAGED:
		return ctm.StatusIssued
	default:
		return ""
	}
}

func populateServiceData(ctx context.Context, serviceData *servicedata.ReplaceCard, oldCard *ctm.DebitCardResponse, newCard *ctm.DebitCardResponse) {
	serviceData.OldNameOnInstrument = strings.TrimSpace(oldCard.EmbossingLine1 + " " + oldCard.EmbossingLine2)
	serviceData.NewNameOnInstrument = strings.TrimSpace(newCard.EmbossingLine1 + " " + newCard.EmbossingLine2)

	oldLastIssueDate := strings.ReplaceAll(oldCard.ReissueDate, "-", "")
	if oldLastIssueDate == "" {
		oldLastIssueDate = strings.ReplaceAll(oldCard.IssueDate, "-", "")
	}
	serviceData.OldLastIssueDate = oldLastIssueDate

	newLastIssueDate := strings.ReplaceAll(newCard.ReissueDate, "-", "")
	if newLastIssueDate == "" {
		newLastIssueDate = strings.ReplaceAll(newCard.IssueDate, "-", "")
	}
	serviceData.NewLastIssueDate = strings.ReplaceAll(newLastIssueDate, "-", "")

	serviceData.OldExpiryDate = GetExpiry(ctx, oldCard.ExpiryDate)

	serviceData.NewExpiryDate = GetExpiry(ctx, newCard.ExpiryDate)

	serviceData.NewTokenizedCardNumber = newCard.CardNumber.Token
	serviceData.NewCardLast_4Digits = newCard.CardNumber.Last4Digits
	serviceData.Last_4Digits = oldCard.CardNumber.Last4Digits
	serviceData.OldIssueReason = oldCard.IssueReason
}

/*
Apply the following, in this order (trimming all whitespace at either end);
Embossed Name = FirstName+" "+LastName
If (embossedName >21) then EmbossedName   = FirstInitial+" "+LastName
If (embossedName >21) then EmbossedName   = FirstInitial+" "+LastName
If (embossedName >21) then EmbossedName   = FirstInitial+" "+(LastName,21)
There cannot be spaces or zeroes else error. no additional spaces or error, no 0's or error in name
currently assumes all chars are ascii
*/
func toEmbossedName(firstName string, lastName string) string {
	firstName = strings.ToUpper(removeSpaces(firstName))
	lastName = strings.ToUpper(removeSpaces(lastName))

	embossedName := removeSpaces(firstName + " " + lastName)
	if len(embossedName) > MaxEmbossedName {
		// edge case no last name but length > 21
		if len(lastName) == 0 {
			return removeSpaces(truncate(firstName, MaxEmbossedName))
		}
		embossedName = truncate(toFirstInitial(firstName)+" "+lastName, MaxEmbossedName)
	}
	return removeSpaces(embossedName)
}

/*
Pass title, Firstname, LastName, and Embossing Name values to CTM Debit Card Setup
title - Hardcoded
firstName - truncate from right to 20 characters.
lastName - truncate from right to 20 characters.
Total Length and Trimming Rule:
length of (title " " first name " " last name) <= 26.
If it exceeds 26 then:
Trim first name from right to left, leaving at least first initial
Trim last name
Using Max Length as 24, as implementation without the 2 spaces is preferred.
*/
func toCTMName(firstName string, lastName string) (string, string) {
	firstName = strings.ToUpper(truncate(removeSpaces(firstName), MaxNameLength))
	lastName = strings.ToUpper(truncate(removeSpaces(lastName), MaxNameLength))

	// Handle no last name edge case
	if len(lastName) == 0 {
		return firstName, lastName
	}

	if len(firstName+lastName) > MaxCtmNameLength {
		firstName = toFirstInitial(firstName)
	}
	return firstName, lastName
}

func removeSpaces(s string) string {
	return strings.TrimSpace(replaceMultipleSpaces(s))
}

// ReplaceMultipleWhiteSpace replaces extra spaces with a single space
func replaceMultipleSpaces(str string) string {
	r := regexp.MustCompile(`\s+`)
	return r.ReplaceAllLiteralString(str, " ")
}

func truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length]
}

func toFirstInitial(str string) string {
	if len(str) > 0 {
		return strings.ToUpper(str[0:1])
	}
	return ""
}

func (s server) publishNotification(ctx context.Context, personaID string) {
	notify := &sdk.NotificationForPersona{
		PersonaID: personaID,
		Notification: notification.Simple{
			ActionURL: "https://plus.anz/cards",
		},
		Preview: notification.Preview{
			Title: "Card Ordered",
			Body:  "We've cancelled your current card. Your new one should arrive in 5 to 10 days.",
		},
		IdempotencyKey: uuid.NewString(),
	}
	res, err := s.CommandCentre.Publish(ctx, notify)
	if err != nil {
		log.Error(ctx, err, "Unable to publish Card Replacement notification to CommandCentre.")
	} else {
		log.Info(ctx, fmt.Sprintf("Successfully published Card Replacement notification to CommandCentre: %v", res.Status))
	}
}

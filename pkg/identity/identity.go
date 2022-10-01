package identity

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/pkg/errors"

	"github.com/anzx/pkg/jwtauth"
)

// Identity contains information extracted from an incoming JWT token about the customer and the id of the user calling
// an API.
type Identity struct {
	// An identifier for the current person the user is acting as (1 user can map to many personas).
	PersonaID string
	// An ID that identifies the current user in the OCV (One Customer View).
	OcvID string
	// An identifier to know the auth type (FakeRock or ForgeRock)
	Issuer string
	// If the customer and the overall claims subject are different
	// This indicates a staff member is likely making a request on a customer's behalf
	HasDifferentSubject bool
}

// Get retrieves id information from JWT claims embedded in the context.
func Get(ctx context.Context) (*Identity, error) {
	claims, ok := jwtauth.GetClaimsFromContext(ctx)
	if !ok {
		err := errors.New("failed to extract JWT claims from context")
		logf.Err(ctx, err)
		return nil, anzerrors.Wrap(err, codes.Internal, "identity error",
			anzerrors.NewErrorInfo(ctx, anzcodes.ContextInvalid, "could not retrieve user identification"))
	}

	customer := claims.GetCustomer()

	if customer.GetSubject() == "" {
		err := errors.New("failed to fetch subject from JWT claims that is extracted from context")
		logf.Err(ctx, err)
		return nil, anzerrors.Wrap(err, codes.Internal, "identity error",
			anzerrors.NewErrorInfo(ctx, anzcodes.ContextInvalid, "could not retrieve user identification"))
	}

	if customer.GetPersona() == nil {
		err := errors.New("failed to fetch persona from JWT claims that is extracted from context")
		logf.Err(ctx, err)
		return nil, anzerrors.Wrap(err, codes.Internal, "identity error",
			anzerrors.NewErrorInfo(ctx, anzcodes.ContextInvalid, "could not retrieve user identification"))
	}

	actor := claims.GetActor()
	hasDifferentSubject := false
	if actor != nil && actor.GetSubject() != customer.GetSubject() {
		hasDifferentSubject = true
	}

	return &Identity{
		PersonaID:           customer.GetPersona().PersonaID,
		OcvID:               customer.GetOCVID(),
		Issuer:              claims.GetBaseClaims().Issuer,
		HasDifferentSubject: hasDifferentSubject,
	}, nil
}

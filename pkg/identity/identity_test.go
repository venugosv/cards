package identity

import (
	"context"
	"testing"

	"github.com/anzx/pkg/jwtauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2/jwt"
)

func TestGetIdentityNoClaims(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id, err := Get(ctx)
	assert.Nil(t, id)
	assert.EqualError(t, err, "fabric error: status_code=Internal, error_code=5, message=identity error, reason=could not retrieve user identification")
}

func TestGetIdentityWithNoPersonaID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = jwtauth.AddClaimsToContext(ctx, jwtauth.NewClaims(jwtauth.BaseClaims{
		Claims: jwt.Claims{
			Subject: "Subject",
		},
	}))

	id, err := Get(ctx)
	assert.Nil(t, id)
	require.Error(t, err)
	assert.EqualError(t, err, "fabric error: status_code=Internal, error_code=5, message=identity error, reason=could not retrieve user identification")
}

func TestGetIdentityWithClaims(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = jwtauth.AddClaimsToContext(ctx, jwtauth.NewClaims(jwtauth.BaseClaims{
		Claims: jwt.Claims{
			Subject: "Subject",
		},
		Persona: &jwtauth.Persona{PersonaID: "Persona"},
		OCVID:   "OCVID",
		Actor: &jwtauth.Actor{
			Subject: "Subject",
		},
	}))

	id, err := Get(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Persona", id.PersonaID)
	assert.Equal(t, "OCVID", id.OcvID)
	assert.False(t, id.HasDifferentSubject)
}

func TestGetIdentityWithClaimsAndDifferentSubject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = jwtauth.AddClaimsToContext(ctx, jwtauth.NewClaims(jwtauth.BaseClaims{
		Claims: jwt.Claims{
			Subject: "Subject",
		},
		Persona: &jwtauth.Persona{PersonaID: "Persona"},
		OCVID:   "OCVID",
		Actor: &jwtauth.Actor{
			Subject: "Coach",
		},
	}))

	id, err := Get(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Persona", id.PersonaID)
	assert.Equal(t, "OCVID", id.OcvID)
	assert.True(t, id.HasDifferentSubject)
}

func TestGetIdentityWithNoSubject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = jwtauth.AddClaimsToContext(ctx, jwtauth.NewClaims(jwtauth.BaseClaims{Claims: jwt.Claims{}}))

	id, err := Get(ctx)
	assert.Nil(t, id)
	require.Error(t, err)
	assert.EqualError(t, err, "fabric error: status_code=Internal, error_code=5, message=identity error, reason=could not retrieve user identification")
}

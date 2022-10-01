/*
Package testutil is providing helper functions for test.
*/
package testutil

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/anzx/pkg/jwtauth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2/jwt"
)

// NoResultsAndErrorContains ensures the all results must be nil and error cannot be nil and all expected error messages
// must be contained in error.
func NoResultsAndErrorContains(t *testing.T, expectedErrorContents []string, err error, results ...interface{}) {
	for _, got := range results {
		require.Nil(t, got)
	}
	ErrorContains(t, expectedErrorContents, err)
}

// ErrorContains asserts all the expected key error content are existing in error.
func ErrorContains(t *testing.T, expectedErrorContents []string, err error) {
	require.Error(t, err)
	require.NotEmpty(t, expectedErrorContents)
	for _, e := range expectedErrorContents {
		assert.Contains(t, err.Error(), e)
	}
}

const staticJWT = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImFwaXNpdHRva2VuLmNvcnAuZGV2LmFueiJ9.eyJpc3MiOiJodHRwczovL2RhdGFwb3dlci1zdHMuYW56LmNvbSIsImF1ZCI6ImF1ZHBjbGllbnQwMi5kZXYuYW56Iiwic3ViIjoiYXVkcGNsaWVudDAyLmRldi5hbnoiLCJleHAiOjE1NzgyNTIwMzcuNzUzLCJzY29wZXMiOlsiQVUuUkVUQUlMLkFDQ09VTlQuUFJPRklMRS5SRUFEIl0sImFtciI6WyJwb3AiXSwiYWNyIjoiSUFMMi5BQUwxLkZBTDEifQ.HiSM1dlHwJWpb4sPE7hSriX8nekh8lNV-MnaDE4RL3mrXGHyOBrlQfa3D13Rb_PDBNdbfqzm79E6ajVVIz5U-2G2CCy1CzT1TuiVlBcyd25HJl4JhiBAKcn4aOAwRbnMp88KLYjVbGdEg4egWhfsaPdBBTEX1M5G0KWfBHAfDA5Lesq5dkSTVRGlun0Q9MhpaZSmEI6FYKt-YDEe7wMifjsEFeDF9a_H8qyyYazopFMv0XM6aIjW000nk-XFzRhBYvznwm_LzafQCVGF5tULOp5jYVnv4d7W1GnH2THMnLtC9WtgQYdQOX1eZlK4QrqsLBXrWotM9v4fy8KP06V5lg"

func GetContext(withAuth bool) context.Context {
	ctx := context.Background()

	claims := jwtauth.NewClaims(
		jwtauth.BaseClaims{
			Claims: jwt.Claims{
				Issuer:   "fakerock.sit.fabric.gcpnp.anz",
				Subject:  uuid.New().String(),
				Expiry:   jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
				IssuedAt: jwt.NewNumericDate(time.Now()),
			},
			Persona: &jwtauth.Persona{
				PersonaID: "Persona",
			},
			OCVID:            "OCVID",
			AuthContextClass: "acr",
		},
	)
	ctx = jwtauth.AddClaimsToContext(ctx, claims)
	if withAuth {
		return metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer "+staticJWT))
	}
	return ctx
}

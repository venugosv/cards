package jwtutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/stretchr/testify/assert"
)

func TestAddJWTToOutgoingContext(t *testing.T) {
	const staticJWT = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImFwaXNpdHRva2VuLmNvcnAuZGV2LmFueiJ9.eyJpc3MiOiJodHRwczovL2RhdGFwb3dlci1zdHMuYW56LmNvbSIsImF1ZCI6ImF1ZHBjbGllbnQwMi5kZXYuYW56Iiwic3ViIjoiYXVkcGNsaWVudDAyLmRldi5hbnoiLCJleHAiOjE1NzgyNTIwMzcuNzUzLCJzY29wZXMiOlsiQVUuUkVUQUlMLkFDQ09VTlQuUFJPRklMRS5SRUFEIl0sImFtciI6WyJwb3AiXSwiYWNyIjoiSUFMMi5BQUwxLkZBTDEifQ.HiSM1dlHwJWpb4sPE7hSriX8nekh8lNV-MnaDE4RL3mrXGHyOBrlQfa3D13Rb_PDBNdbfqzm79E6ajVVIz5U-2G2CCy1CzT1TuiVlBcyd25HJl4JhiBAKcn4aOAwRbnMp88KLYjVbGdEg4egWhfsaPdBBTEX1M5G0KWfBHAfDA5Lesq5dkSTVRGlun0Q9MhpaZSmEI6FYKt-YDEe7wMifjsEFeDF9a_H8qyyYazopFMv0XM6aIjW000nk-XFzRhBYvznwm_LzafQCVGF5tULOp5jYVnv4d7W1GnH2THMnLtC9WtgQYdQOX1eZlK4QrqsLBXrWotM9v4fy8KP06V5lg"
	ctx := AddJWTToOutgoingContext(context.Background(), staticJWT)
	got, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	assert.Equal(t, fmt.Sprintf("Bearer %s", staticJWT), got["authorization"][0])
}

func TestPopulateJWTToOutgoingContext(t *testing.T) {
	const staticJWT = "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImFwaXNpdHRva2VuLmNvcnAuZGV2LmFueiJ9.eyJpc3MiOiJodHRwczovL2RhdGFwb3dlci1zdHMuYW56LmNvbSIsImF1ZCI6ImF1ZHBjbGllbnQwMi5kZXYuYW56Iiwic3ViIjoiYXVkcGNsaWVudDAyLmRldi5hbnoiLCJleHAiOjE1NzgyNTIwMzcuNzUzLCJzY29wZXMiOlsiQVUuUkVUQUlMLkFDQ09VTlQuUFJPRklMRS5SRUFEIl0sImFtciI6WyJwb3AiXSwiYWNyIjoiSUFMMi5BQUwxLkZBTDEifQ.HiSM1dlHwJWpb4sPE7hSriX8nekh8lNV-MnaDE4RL3mrXGHyOBrlQfa3D13Rb_PDBNdbfqzm79E6ajVVIz5U-2G2CCy1CzT1TuiVlBcyd25HJl4JhiBAKcn4aOAwRbnMp88KLYjVbGdEg4egWhfsaPdBBTEX1M5G0KWfBHAfDA5Lesq5dkSTVRGlun0Q9MhpaZSmEI6FYKt-YDEe7wMifjsEFeDF9a_H8qyyYazopFMv0XM6aIjW000nk-XFzRhBYvznwm_LzafQCVGF5tULOp5jYVnv4d7W1GnH2THMnLtC9WtgQYdQOX1eZlK4QrqsLBXrWotM9v4fy8KP06V5lg"
	md := metadata.MD{
		"authorization": []string{staticJWT},
	}
	t.Run("happy path", func(t *testing.T) {
		incomingContext := metadata.NewIncomingContext(context.Background(), md)
		outgoingContext := populateJWTToOutgoingContext(incomingContext)
		got, ok := metadata.FromOutgoingContext(outgoingContext)
		require.True(t, ok)
		assert.Equal(t, staticJWT, got["authorization"][0])
	})
	t.Run("unhappy path", func(t *testing.T) {
		og := metadata.NewOutgoingContext(context.Background(), md)
		outgoingContext := populateJWTToOutgoingContext(og)
		_, ok := metadata.FromIncomingContext(outgoingContext)
		require.False(t, ok)
	})
}

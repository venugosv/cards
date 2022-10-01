package forgerock

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/util/jwtutil"
)

const staticJWT = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImFwaXNpdHRva2VuLmNvcnAuZGV2LmFueiJ9.eyJpc3MiOiJodHRwczovL2RhdGFwb3dlci1zdHMuYW56LmNvbSIsImF1ZCI6ImF1ZHBjbGllbnQwMi5kZXYuYW56Iiwic3ViIjoiYXVkcGNsaWVudDAyLmRldi5hbnoiLCJleHAiOjE1NzgyNTIwMzcuNzUzLCJzY29wZXMiOlsiQVUuUkVUQUlMLkFDQ09VTlQuUFJPRklMRS5SRUFEIl0sImFtciI6WyJwb3AiXSwiYWNyIjoiSUFMMi5BQUwxLkZBTDEifQ.HiSM1dlHwJWpb4sPE7hSriX8nekh8lNV-MnaDE4RL3mrXGHyOBrlQfa3D13Rb_PDBNdbfqzm79E6ajVVIz5U-2G2CCy1CzT1TuiVlBcyd25HJl4JhiBAKcn4aOAwRbnMp88KLYjVbGdEg4egWhfsaPdBBTEX1M5G0KWfBHAfDA5Lesq5dkSTVRGlun0Q9MhpaZSmEI6FYKt-YDEe7wMifjsEFeDF9a_H8qyyYazopFMv0XM6aIjW000nk-XFzRhBYvznwm_LzafQCVGF5tULOp5jYVnv4d7W1GnH2THMnLtC9WtgQYdQOX1eZlK4QrqsLBXrWotM9v4fy8KP06V5lg"

type StubClient struct {
	Err error
}

func NewStubClient() StubClient {
	return StubClient{}
}

func (s StubClient) SystemJWT(ctx context.Context, _ ...string) (context.Context, error) {
	if s.Err != nil {
		return ctx, s.Err
	}

	return jwtutil.AddJWTToOutgoingContext(ctx, getJwt(staticJWT).AccessToken), nil
}

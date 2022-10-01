package echidna

import (
	"fmt"
	"testing"

	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func Test_GetGRPCError(t *testing.T) {
	tests := map[int]codes.Code{
		0:          codes.OK,
		55:         codes.InvalidArgument,
		75:         codes.ResourceExhausted,
		1010:       codes.Internal,
		1011:       codes.Internal,
		1012:       codes.Unavailable,
		1013:       codes.DeadlineExceeded,
		1014:       codes.Unavailable,
		1015:       codes.Unavailable,
		1234567890: codes.Unknown,
	}
	for key, value := range tests {
		name := fmt.Sprintf("Code: %d == %s", key, value)
		t.Run(name, func(t *testing.T) {
			got := GetGRPCError(key)
			assert.Equal(t, value, got)
		})
	}
}

func Test_GetANZError(t *testing.T) {
	tests := map[int]anzcodes.Code{
		55:         anzcodes.ValidationFailure,
		75:         anzcodes.RateLimitExhausted,
		1010:       anzcodes.DownstreamFailure,
		1011:       anzcodes.DownstreamFailure,
		1012:       anzcodes.DownstreamFailure,
		1013:       anzcodes.DownstreamFailure,
		1014:       anzcodes.DownstreamFailure,
		1015:       anzcodes.DownstreamFailure,
		1234567890: anzcodes.Unknown,
	}
	for key, value := range tests {
		name := fmt.Sprintf("Code: %d == %v", key, value)
		t.Run(name, func(t *testing.T) {
			got := GetANZError(key)
			assert.Equal(t, value, got)
		})
	}
}

func Test_GetErrorMsg(t *testing.T) {
	tests := map[int]string{
		55:         "Incorrect PIN",
		75:         "Maximum PIN tries exceeded",
		1010:       "Operation failed due to internal error",
		1011:       "Operation failed due to internal error",
		1012:       "Service unavailable",
		1013:       "Operation has timed out",
		1014:       "Information is unavailable within the PIN service.",
		1015:       "Operation failed due to service error",
		1234567890: "Unknown",
	}
	for key, value := range tests {
		name := fmt.Sprintf("Code: %d == %s", key, value)
		t.Run(name, func(t *testing.T) {
			got := GetErrorMsg(key)
			assert.Equal(t, value, got)
		})
	}
}

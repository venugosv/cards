package cards

import (
	"testing"

	"github.com/anzx/fabric-cards/test/data"

	"github.com/anzx/fabric-cards/test/fixtures"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	wrappingKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArI6WKTJMLVfpaG+Mkaj4IVX3/2dbtHvacI9sKfutMsg5It6pEvFf9oYoIWMQkxFARf14ds0+1t83sm6foPHm4HZ0oP2GX0iiFdALEZr3C6C2FXAoQQXYMGeczoeta0IwF75B3Pr6VETQjf7niL00MF0n/McsE9tu9VTOFjq6LkvZgOnBe9wG+f0nvdx29FAPzIjdpBoZ27Ingmtnmtk2T9oadY5vXE2ruIhjU2rL/8aPPN8LtvlWrcV0y+YW2l4EMGenAFYMu4jh6R5deNfartmNotJgbzHFcD7EpXJivzYgdMvea2Dy7AjlC5cic4ijcna750HhfMoFFNqf6T7psQIDAQAB"
)

func TestCards_GetWrappingKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		personaID string
		builder   *fixtures.ServerBuilder
		want      *cpb.GetWrappingKeyResponse
		wantErr   error
	}{
		{
			name:    "EchidnaClient Call fails with error code 1014 Public key information is unavailable within the RemotePIN service, return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1014),
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=failed to get wrapping key, reason=Information is unavailable within the PIN service"),
		},
		{
			name:    "EchidnaClient Call fails with error code 1015 Public key information is unavailable due to RemotePIN service, return error",
			builder: fixtures.AServer().WithData(data.AUserWithACard()).WithEchidnaErrorCode(1015),
			wantErr: errors.New("fabric error: status_code=Unavailable, error_code=2, message=failed to get wrapping key, reason=Operation failed due to service error"),
		},
		{
			name:    "EchidnaClient Call succeeds, return status true",
			builder: fixtures.AServer().WithData(data.AUserWithACard()),
			want: &cpb.GetWrappingKeyResponse{
				EncodedKey: wrappingKey,
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			s := buildCardServer(tt.builder)
			got, err := s.GetWrappingKey(fixtures.GetTestContext(), &cpb.GetWrappingKeyRequest{})
			if test.wantErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

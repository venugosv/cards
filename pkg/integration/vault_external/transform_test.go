package vault_external

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type FakeVaultAPI struct {
	runResponse   []byte
	runError      string
	loginResponse *Secret
	loginError    string
	countRun      int
	countLogin    int
}

func (f *FakeVaultAPI) run(_ context.Context, _ string, _ string, _ string, _ []byte) ([]byte, error) {
	f.countRun += 1
	if f.runError != "" {
		return nil, errors.New(f.runError)
	}
	return f.runResponse, nil
}

func (f *FakeVaultAPI) vaultLogin(_ context.Context, _ string, _ string) (*Secret, error) {
	f.countLogin += 1
	if f.loginError != "" {
		return nil, errors.New(f.loginError)
	}
	return f.loginResponse, nil
}

func TestClient_Transform(t *testing.T) {
	validResponse := BatchTransformResponse{
		Data: &BatchTransformResults{
			Results: []*TransformResult{
				{
					EncodedValue: "xyz",
				},
			},
		},
	}

	validResponseBytes, _ := json.Marshal(validResponse)

	emptyResponse := BatchTransformResponse{}

	emptyResponseBytes, _ := json.Marshal(emptyResponse)

	aGoodRequest := []*TransformRequest{
		{
			Value: "abc",
		},
	}

	tests := []struct {
		name               string
		vaultResponseBytes []byte
		fixedToken         string
		request            []*TransformRequest
		operation          TransformKind
		role               string
		restError          string
		expectedError      string
		expectedRunCount   int
	}{
		{
			name:               "happy path encode",
			vaultResponseBytes: validResponseBytes,
			fixedToken:         "foo",
			request: []*TransformRequest{
				{
					Value: "abc",
				},
			},
			operation:        TransformEncode,
			role:             "something",
			expectedRunCount: 1,
		},
		{
			name:               "error vault response",
			vaultResponseBytes: []byte("abcdefg"),
			fixedToken:         "foo",
			request:            aGoodRequest,
			operation:          TransformEncode,
			role:               "something",
			expectedError:      "fabric error: status_code=Internal, error_code=2, message=transform failed, reason=could not unmarshal JSON from transform response: invalid character 'a' looking for beginning of value",
		},
		{
			name:               "error nil response data",
			vaultResponseBytes: emptyResponseBytes,
			fixedToken:         "foo",
			request:            aGoodRequest,
			operation:          TransformDecode,
			role:               "something",
			expectedError:      "fabric error: status_code=Internal, error_code=2, message=transform failed, reason=transform response JSON missing result data",
		},
		{
			name:          "error rest client",
			fixedToken:    "foo",
			request:       aGoodRequest,
			operation:     TransformEncode,
			role:          "something",
			restError:     "failed",
			expectedError: "fabric error: status_code=Unknown, error_code=0, message=transform failed, reason=UNKNOWN",
		},
		{
			name:               "with auth path",
			vaultResponseBytes: validResponseBytes,
			request:            aGoodRequest,
			operation:          TransformEncode,
			role:               "something",
			expectedRunCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5)
			fakeRest := &FakeVaultAPI{
				runResponse: tt.vaultResponseBytes,
				runError:    tt.restError,
			}
			c := client{
				api: fakeRest,
				jwtSigner: &FixedSignedJwt{
					jwt: ".",
					key: "m",
				},
				config: &Config{
					OverrideServiceEmail: "fabric@anz.com",
				},
				auth: auth{
					renewed:   make(chan interface{}),
					blockTime: time.Duration(100) * time.Millisecond,
					until:     time.Now().Add(999 * time.Hour),
					token:     "foobar",
				},
			}
			resp, err := c.Transform(ctx, tt.operation, tt.role, tt.request)
			if tt.expectedError != "" {
				require.Error(t, err, "transform did not throw expected error")
				require.Equal(t, tt.expectedError, err.Error(), "error does not match")
			} else {
				require.NoError(t, err, "unexpected error from Transform call %v", err)
				require.NotNil(t, resp)
				require.Equal(t, tt.expectedRunCount, fakeRest.countRun, "run called incorrect number of times")
			}
			go func() {
				time.Sleep(time.Second * 5)
				cancel()
			}()
		})
	}
}

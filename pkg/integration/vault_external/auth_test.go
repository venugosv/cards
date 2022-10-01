package vault_external

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
)

func TestCalculateDelayWithBuffer(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		buffer   int64
		expected int64
	}{
		{
			name:     "buffer = 0",
			duration: 123,
			buffer:   0,
			expected: 123,
		},
		{
			name:     "duration > buffer",
			duration: 300,
			buffer:   100,
			expected: 200,
		},
		{
			name:     "duration < buffer",
			duration: 300,
			buffer:   500,
			expected: 300,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			result := calculateDelayWithBuffer(
				time.Duration(test.duration),
				time.Duration(test.buffer),
			)
			require.Equal(t, test.expected, result.Nanoseconds())
		})
	}
}

func TestKeepAuthValid(t *testing.T) {
	tests := []struct {
		name           string
		configLifetime int
		lifetime       int
		token          string
		loginError     string
		waitTime       int
		minLoginCount  int
		maxLoginCount  int
	}{
		{
			name:           "happy path",
			token:          "foo",
			lifetime:       1,
			configLifetime: 100,
			waitTime:       2 * 1000,
			minLoginCount:  2,
			maxLoginCount:  2,
		},
		{
			name:           "backoff",
			loginError:     "foo",
			configLifetime: 10,
			waitTime:       100,
			minLoginCount:  7,
			maxLoginCount:  13,
			lifetime:       5,
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			fakeAPI := &FakeVaultAPI{
				loginResponse: &Secret{
					Auth: &SecretAuth{
						ClientToken:   test.token,
						LeaseDuration: test.lifetime,
					},
				},
				loginError: test.loginError,
			}

			c := client{
				config: &Config{
					OverrideServiceEmail: "foo@local",
					TokenLifetime:        time.Duration(test.configLifetime) * time.Millisecond,
				},
				jwtSigner: &FixedSignedJwt{
					jwt: ".",
					key: ".",
				},
				api: fakeAPI,
				backoff: &backoff.ExponentialBackOff{
					InitialInterval:     10 * time.Millisecond,
					MaxElapsedTime:      10 * time.Minute,
					Multiplier:          1,
					RandomizationFactor: 0,
					MaxInterval:         10 * time.Millisecond,
					Clock:               backoff.SystemClock,
					Stop:                -1,
				},
			}

			c.backoff.Reset()

			cancel := keepAuthValid(context.Background(), &c)

			time.Sleep(time.Duration(test.waitTime) * time.Millisecond)

			cancel()

			require.GreaterOrEqual(t, fakeAPI.countLogin, test.minLoginCount)
			require.LessOrEqual(t, fakeAPI.countLogin, test.maxLoginCount)
		})
	}
}

func TestAuth_AwaitValidToken_WithManyWaiting(t *testing.T) {
	fakeAPI := &FakeVaultAPI{
		loginResponse: &Secret{
			Auth: &SecretAuth{
				ClientToken:   "12345",
				LeaseDuration: 900,
			},
		},
	}
	c := client{
		config: &Config{
			OverrideServiceEmail: "foo@local",
			TokenLifetime:        time.Duration(10) * time.Millisecond,
		},
		jwtSigner: &FixedSignedJwt{
			jwt: ".",
			key: ".",
		},
		api: fakeAPI,
		auth: auth{
			until:     time.Now().Add(-10 * time.Minute),
			blockTime: time.Duration(10) * time.Minute,
			renewed:   make(chan interface{}),
		},
		backoff: backoff.NewExponentialBackOff(),
	}

	waiters := int64(10)
	var counter int64
	var wg sync.WaitGroup

	for i := int64(0); i < waiters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.auth.awaitValidToken()
			atomic.AddInt64(&counter, 1)
		}()
	}

	cancel := keepAuthValid(context.Background(), &c)

	wg.Wait()

	cancel()

	require.Equal(t, waiters, counter)
}

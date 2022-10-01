package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/anzx/pkg/gsm"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/go-redis/redis_rate/v9"

	"github.com/go-redis/redis/v8"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anzx/fabric-cards/pkg/util/testutil"
)

type mockSecretManager struct {
	accessSecretVersionFunc func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

func (m mockSecretManager) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return m.accessSecretVersionFunc(ctx, req, opts...)
}

func TestRateLimit_RedisUnavailable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	config := &Config{
		Redis: RedisConfig{
			Addr:     "testinghost",
			DB:       0,
			Password: "redispassword",
		},
		Limits: nil,
		Prefix: "st",
	}
	opts := &redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	}

	redisClient := redis.NewClient(opts)

	client := &RedisRateLimit{
		Prefix:  config.Prefix,
		Limits:  config.Limits,
		Limiter: redis_rate.NewLimiter(redisClient),
	}

	err := client.Allow(ctx, "activate")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fabric error: status_code=ResourceExhausted, error_code=7, message=rate limit check failed, reason=service unavailable")
}

func TestRateLimit_RedisAvailable(t *testing.T) {
	t.Parallel()
	// Using a mocked redis server
	s, err := miniredis.Run()
	require.NoError(t, err)
	s.RequireAuth("redispassword")
	defer s.Close()

	config := &Config{
		Redis: RedisConfig{
			Addr:     s.Addr(),
			DB:       0,
			Password: "redispassword",
		},
		Limits: map[Domain]LimitConfig{
			Activate: {
				Rate:   1,
				Period: 2 * time.Second,
			},
		},
	}

	gsmClient := &gsm.Client{
		SM: mockSecretManager{
			accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
				return &secretmanagerpb.AccessSecretVersionResponse{
					Name:    "testName",
					Payload: &secretmanagerpb.SecretPayload{Data: []byte("redispassword")},
				}, nil
			},
		},
	}

	client, _ := NewClient(context.Background(), config, gsmClient)

	// First time okay
	err = client.Allow(testutil.GetContext(false), "activate")
	assert.NoError(t, err)

	// Second time return error
	err = client.Allow(testutil.GetContext(false), "activate")
	assert.Error(t, err)
}

package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/pkg/gsm"
	"github.com/pkg/errors"

	"github.com/go-redis/redis_rate/v9"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/pkg/identity"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
)

// Domain define a rate limit category
type Domain string

const (
	Activate  Domain = "activate"
	VerifyPIN Domain = "verifypin"
)

type Config struct {
	Redis  RedisConfig            `json:"redis" yaml:"redis" mapstructure:"redis" validate:"required"`
	Limits map[Domain]LimitConfig `json:"limits" yaml:"limits" mapstructure:"limits" validate:"required"`
	// Prefix to be added to every cache key, can be empty
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty" mapstructure:"prefix"`
}

type LimitConfig struct {
	// Number of requests allowed in the given period
	Rate int `json:"rate" yaml:"rate" mapstructure:"rate" validate:"gte=1"`
	// Period in time.duration
	Period time.Duration `json:"period" yaml:"period" mapstructure:"period" validate:"gte=0"`
}

type RateLimit interface {
	Allow(ctx context.Context, domain Domain) error
}

func NewClient(ctx context.Context, config *Config, gsmClient *gsm.Client) (RateLimit, error) {
	if config == nil {
		logf.Debug(ctx, "ratelimit config not provided %v", config)
		return nil, nil
	}

	if err := config.Redis.GetSecrets(ctx, gsmClient); err != nil {
		logf.Error(ctx, err, "ratelimit: failed to get redis secret")
		return nil, errors.Wrap(err, "unable to access secret")
	}

	redisClient, err := newRedisClient(ctx, config.Redis)
	if err != nil {
		return nil, err
	}

	return &RedisRateLimit{
		Prefix:  config.Prefix,
		Limits:  config.Limits,
		Limiter: redis_rate.NewLimiter(redisClient),
	}, nil
}

// Check return error if not pass the check, otherwise return nil
func (r *RedisRateLimit) Allow(ctx context.Context, domain Domain) error {
	limitConfig, ok := r.Limits[domain]
	if !ok {
		logf.Debug(ctx, "unable to find rate limit domain: %v", domain)
		return anzerrors.New(codes.ResourceExhausted, "rate limit check failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.RateLimitExhausted, "service unavailable"))
	}

	limit := redis_rate.Limit{
		Rate:   limitConfig.Rate,
		Burst:  limitConfig.Rate,
		Period: limitConfig.Period,
	}

	id, err := identity.Get(ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%s:%s", r.Prefix, domain, id.PersonaID)

	result, err := r.Limiter.Allow(ctx, key, limit)
	if err != nil {
		logf.Error(ctx, err, "rate limiter failed with key: %v", key)
		return anzerrors.Wrap(err, codes.Internal, "rate limit check failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.DownstreamFailure, "rate limit check failed"))
	}

	if result.Allowed == 0 {
		return anzerrors.New(codes.ResourceExhausted, "rate limit check failed",
			anzerrors.NewErrorInfo(ctx, anzcodes.RateLimitExhausted, "over rate limit"),
			anzerrors.WithRetryDelay(result.RetryAfter))
	}

	return nil
}

func (c *Config) Byte() []byte {
	out, _ := json.Marshal(c)
	return out
}

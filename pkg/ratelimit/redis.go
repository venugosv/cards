package ratelimit

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/pkg/errors"

	"github.com/anzx/pkg/gsm"
)

type RedisConfig struct {
	Addr      string      `json:"addr" yaml:"addr" mapstructure:"addr" validate:"required"` // Redis server address
	DB        int         `json:"db" yaml:"db" mapstructure:"db" validate:"gte=0"`
	Password  string      `json:"-" yaml:"-"`
	SecretID  string      `json:"secretId,omitempty" yaml:"secretId,omitempty" mapstructure:"secretId" validate:"required"`
	TLSCertID string      `json:"tlsCertId,omitempty" yaml:"tlsCertId,omitempty" mapstructure:"tlsCertId"`
	TlsConfig *tls.Config `json:"-" yaml:"-"`
}

func (c *RedisConfig) GetSecrets(ctx context.Context, secrets *gsm.Client) error {
	password, err := secrets.AccessSecret(ctx, c.SecretID)
	if err != nil {
		return err
	}
	c.Password = password

	if c.TLSCertID != "" {
		// get certificate of the Certificate Authorities (CA) that signed
		caCert, err := secrets.AccessSecret(ctx, c.TLSCertID)
		if err != nil {
			return err
		}

		// create certificate pool and then add the CA’s certificate to the pool
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM([]byte(caCert)) {
			return fmt.Errorf("cannot add caCert to pool")
		}
		c.TlsConfig = &tls.Config{
			RootCAs:    caPool,
			MinVersion: tls.VersionTLS12,
		}
	}
	return nil
}

type RedisRateLimit struct {
	Prefix  string
	Limits  map[Domain]LimitConfig
	Limiter *redis_rate.Limiter
}

func newRedisClient(ctx context.Context, config RedisConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:      config.Addr,
		Password:  config.Password,
		DB:        config.DB,
		TLSConfig: config.TlsConfig,
	}

	client := redis.NewClient(opts).WithContext(ctx)

	ping, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to ping redis: %s", ping))
	}

	return client, nil
}

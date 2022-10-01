package vault_external

import (
	"context"
	"sync"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
)

// auth represents the state of a auth with its value and lifetime
type auth struct {
	sync.Mutex
	token     string
	until     time.Time
	blockTime time.Duration
	renewed   chan interface{}
}

func (t *auth) getToken() string {
	t.Lock()
	defer t.Unlock()
	return t.token
}

func (t *auth) isValid() bool {
	t.Lock()
	defer t.Unlock()
	return time.Now().Before(t.until)
}

// awaitValidToken will wait for a valid token, but times out after 5 seconds. it returns true if there is a new
// valid token to get
func (t *auth) awaitValidToken() bool {
	timeout := time.NewTimer(t.blockTime)
	select {
	case <-t.renewed:
		// drain the timer channel if needed
		if !timeout.Stop() {
			<-timeout.C
		}
		return true
	case <-timeout.C:
		return false
	}
}

func (t *auth) set(value string, duration time.Duration) {
	t.Lock()
	defer t.Unlock()
	t.token = value
	t.until = time.Now().Add(duration)
}

func keepAuthValid(ctx context.Context, c *client) func() {
	initialDelay := calculateDelayWithBuffer(c.config.TokenLifetime, c.config.TokenRenewBuffer)
	timer := time.NewTimer(initialDelay)

	doneChannel := make(chan bool)

	go func() {
		for {
			select {
			case <-timer.C:
				break
			case <-doneChannel:
				logf.Info(ctx, "exited keep auth valid loop")
				return
			}
			newAuth, err := c.login(ctx)
			if err != nil {
				logf.Debug(ctx, "got error from Vault login, backing off")
				delay := c.backoff.NextBackOff()
				logf.Error(ctx, err, "failed to renew vault API auth, retrying in %f s", delay.Seconds())
				timer.Reset(delay)
				continue
			}

			leaseDuration := time.Second * time.Duration(newAuth.LeaseDuration)
			if leaseDuration == 0 {
				leaseDuration = c.config.TokenLifetime
			}
			renewDelay := calculateDelayWithBuffer(leaseDuration, c.config.TokenRenewBuffer)
			timer.Reset(renewDelay)

			logf.Info(ctx, "successfully refreshed Vault auth, lease duration is %s, renew in %s", leaseDuration.String(), renewDelay.String())
			c.backoff.Reset()
			c.auth.set(newAuth.ClientToken, leaseDuration)

			// Ping everything receiving on our "renewed" channel
			hasListeners := true
			for hasListeners {
				select {
				case c.auth.renewed <- true:
					continue
				default:
					// This branch is taken when we cannot send, eg nobody is listening.
					// So this is where we end the loop
					hasListeners = false
				}
			}
		}
	}()

	return func() {
		doneChannel <- true
	}
}

func calculateDelayWithBuffer(duration time.Duration, renewBuffer time.Duration) time.Duration {
	if duration < renewBuffer {
		return duration
	} else {
		return duration - renewBuffer
	}
}

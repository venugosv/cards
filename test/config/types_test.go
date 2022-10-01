package config

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	t.Run("", func(t *testing.T) {
		cfg := Config{
			Timeout: 0,
			Callback: Callback{
				Service: Service{
					Insecure: true,
					BaseURL:  "1234567890",
					Auth:     AuthConfig{},
					Headers:  nil,
					TearDown: false,
					Scheme:   "https",
				},
				CurrentCard:   "4622393000000173",
				PubsubTimeout: time.Second,
			},
			Cards:         Service{},
			CardControls:  Service{},
			CommandCentre: CommandCentre{},
			Vault:         nil,
		}
		fmt.Println(cfg.String())
	})
}

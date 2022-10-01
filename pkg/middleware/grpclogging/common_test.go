package grpclogging

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestDefaultDeciderMethod(t *testing.T) {
	t.Run("default method returns true", func(t *testing.T) {
		assert.True(t, DefaultDeciderMethod("", nil))
	})
}

func TestDefaultErrorToCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{
			name: "random error returns unknown",
			err:  errors.New("oh no"),
			want: codes.Unknown,
		},
		{
			name: "nil error returns ok",
			err:  nil,
			want: codes.OK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultErrorToCode(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

var deciderTests = []struct {
	name string
	req  string
	want bool
}{
	{
		name: "decider returns true successfully",
		req:  "this",
		want: true,
	},
	{
		name: "decider returns true with caps condition successfully",
		req:  "THIS",
		want: true,
	},
	{
		name: "decider returns false successfully",
		req:  "that",
		want: false,
	},
	{
		name: "decider returns false by default",
		req:  "something else",
		want: false,
	},
}

func TestClientDecider(t *testing.T) {
	for _, test := range deciderTests {
		t.Run(test.name, func(t *testing.T) {
			cfg := PayloadLoggingDecider{
				Client: map[string]bool{
					"this": true,
					"that": false,
				},
			}
			got := ClientPayloadDecider(cfg)
			assert.Equal(t, test.want, got(context.Background(), test.req))
		})
	}
}

func TestServerPayloadDecider(t *testing.T) {
	for _, test := range deciderTests {
		t.Run(test.name, func(t *testing.T) {
			cfg := PayloadLoggingDecider{
				Server: map[string]bool{
					"this": true,
					"that": false,
				},
			}
			got := ServerPayloadDecider(cfg)
			assert.Equal(t, test.want, got(context.Background(), test.req, nil))
		})
	}
}

func TestServerDecider(t *testing.T) {
	for _, test := range deciderTests {
		t.Run(test.name, func(t *testing.T) {
			cfg := PayloadLoggingDecider{
				Server: map[string]bool{
					"this": true,
					"that": false,
				},
			}
			got := ServerDecider(cfg)
			assert.Equal(t, test.want, got(test.req, nil))
		})
	}
}

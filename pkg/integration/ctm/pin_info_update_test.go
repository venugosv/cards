package ctm

import (
	"errors"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"github.com/anzx/fabric-cards/pkg/util/testutil"
	testUtil "github.com/anzx/fabric-cards/test/util"
	"github.com/stretchr/testify/assert"
)

func Test_client_UpdatePINInfo(t *testing.T) {
	tests := []struct {
		name     string
		req      string
		want     bool
		wantErr  bool
		mockAPIc apic.Clienter
	}{
		{
			name:     "true returned on downstream 204",
			req:      tokenizedCardNumber,
			want:     true,
			mockAPIc: testUtil.MockAPIcer{},
		},
		{
			name:     "false returned on downstream 500",
			req:      tokenizedCardNumber,
			want:     false,
			wantErr:  true,
			mockAPIc: testUtil.MockAPIcer{ResponseErr: errors.New("failed request")},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				apicClient: tt.mockAPIc,
			}

			got, err := c.UpdatePINInfo(testutil.GetContext(true), tt.req)

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

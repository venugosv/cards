package ctm

import (
	"context"
	"errors"
	"testing"

	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestGetAddress(t *testing.T) {
	tests := []struct {
		name    string
		args    *sspb.Address
		want    MailingAddress
		wantErr error
	}{
		{
			name: "2 blank lines",
			args: &sspb.Address{LineOne: "15 station street", LineTwo: "", LineThree: "", City: "Reservoir", State: "AU-VI", PostalCode: "3011", Country: "AUS"},
			want: MailingAddress{
				AddressLine1: "15 station street",
				AddressLine2: "Reservoir VIC 3011",
				PostCode:     "3011",
			},
			wantErr: nil,
		},
		{
			name: "1 blank lines",
			args: &sspb.Address{LineOne: "Unit 4", LineTwo: "15 station street", LineThree: "", City: "Reservoir", State: "AU-VI", PostalCode: "3011", Country: "AUS"},
			want: MailingAddress{
				AddressLine1: "Unit 4",
				AddressLine2: "15 station street",
				AddressLine3: "Reservoir VIC 3011",
				PostCode:     "3011",
			},
			wantErr: nil,
		},
		{
			name: "0 blank lines",
			args: &sspb.Address{LineOne: "Unit 4", LineTwo: "Block B", LineThree: "15 station street", City: "Reservoir", State: "AU-VI", PostalCode: "3011", Country: "AUS"},
			want: MailingAddress{
				AddressLine1: "Unit 4",
				AddressLine2: "Block B 15 station street",
				AddressLine3: "Reservoir VIC 3011",
				PostCode:     "3011",
			},
			wantErr: nil,
		},
		{
			name: "Length > 36 on line 2 truncation test",
			args: &sspb.Address{LineOne: "Unit 4", LineTwo: "Block B that is really really long and should be less than 36", LineThree: "15 station street", City: "Reservoir", State: "AU-VI", PostalCode: "3011", Country: "AUS"},
			want: MailingAddress{
				AddressLine1: "Unit 4",
				AddressLine2: "Block B that is really really long a",
				AddressLine3: "Reservoir VIC 3011",
				PostCode:     "3011",
			},
			wantErr: nil,
		},

		{
			name: "Length = 36 on line 2 truncation test",
			args: &sspb.Address{LineOne: "Unit 4", LineTwo: "Block B equal 36", LineThree: "15 station street .", City: "Reservoir", State: "AU-VI", PostalCode: "3011", Country: "AUS"},
			want: MailingAddress{
				AddressLine1: "Unit 4",
				AddressLine2: "Block B equal 36 15 station street .",
				AddressLine3: "Reservoir VIC 3011",
				PostCode:     "3011",
			},
			wantErr: nil,
		},
		{
			name: "typical residential address",
			args: &sspb.Address{City: "Footscray", LineOne: "310/244 Barkly Street", PostalCode: "3011", Country: "AUS", State: "VIC"},
			want: MailingAddress{
				AddressLine1: "310/244 Barkly Street",
				AddressLine2: "Footscray VIC 3011",
				AddressLine3: "",
				PostCode:     "3011",
			},
			wantErr: nil,
		},
		{
			name:    "handle non australian address",
			args:    &sspb.Address{City: "Boston", LineOne: "213 Derrick Street", PostalCode: "02130", Country: "USA", State: ""},
			wantErr: errors.New("fabric error: status_code=FailedPrecondition, error_code=20003, message=Invalid Address, reason=address on customer profile is international"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := GetAddress(ctx, tt.args)
			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want.AddressLine1, got.AddressLine1, "GetAddress() produced an incorrect Line1 got = %v, want %v", got.AddressLine1, tt.want.AddressLine1)
				assert.LessOrEqual(t, len(got.AddressLine1), maxCardMailingAddressLength)

				assert.Equal(t, tt.want.AddressLine2, got.AddressLine2, "GetAddress() produced an incorrect Line1 got = %v, want %v", got.AddressLine2, tt.want.AddressLine2)
				assert.LessOrEqual(t, len(got.AddressLine2), maxCardMailingAddressLength)

				assert.Equal(t, tt.want.AddressLine3, got.AddressLine3, "GetAddress() produced an incorrect Line1 got = %v, want %v", got.AddressLine3, tt.want.AddressLine3)
				assert.LessOrEqual(t, len(got.AddressLine3), maxCardMailingAddressLength)

				assert.Equal(t, tt.want.PostCode, got.PostCode, "GetAddress() produced an incorrect Line1 got = %v, want %v", got.PostCode, tt.want.PostCode)
			}
		})
	}
}

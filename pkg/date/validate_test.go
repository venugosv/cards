package date

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pb "github.com/anzx/fabricapis/pkg/fabric/type"
)

func TestDate_IsValid(t *testing.T) {
	tests := []struct {
		name string
		d    *Date
		want bool
	}{
		{
			name: "2006/6/1",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			want: true,
		},
		{
			name: "6/1",
			d: &Date{
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			want: true,
		},
		{
			name: "2006/6",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
			},
			want: true,
		},
		{
			name: "20061/6/1",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 20061},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			want: false,
		},
		{
			name: "6/1",
			d: &Date{
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 100},
			},
			want: false,
		},
		{
			name: "2006/6",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 60},
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.d.IsValid())
		})
	}
}

func TestDate_IsValidMonth(t *testing.T) {
	tests := map[int32]bool{
		0:  false,
		1:  true,
		2:  true,
		3:  true,
		4:  true,
		5:  true,
		6:  true,
		7:  true,
		8:  true,
		9:  true,
		10: true,
		11: true,
		12: true,
		13: false,
		32: false,
	}
	for in, want := range tests {
		t.Run(time.Month(in).String(), func(t *testing.T) {
			d := &Date{
				Month: &pb.OptionalInt32{Value: in},
			}
			assert.Equal(t, want, d.IsValidMonth())
		})
	}
	t.Run("nil check", func(t *testing.T) {
		d := &Date{}
		assert.False(t, d.IsValidMonth())
	})
}

func TestDate_IsValidDay(t *testing.T) {
	tests := map[int32]bool{
		0:  false,
		1:  true,
		2:  true,
		3:  true,
		4:  true,
		5:  true,
		6:  true,
		7:  true,
		8:  true,
		9:  true,
		10: true,
		11: true,
		12: true,
		13: true,
		14: true,
		15: true,
		16: true,
		17: true,
		18: true,
		19: true,
		20: true,
		21: true,
		22: true,
		23: true,
		24: true,
		25: true,
		26: true,
		27: true,
		28: true,
		29: true,
		30: true,
		31: true,
		32: false,
	}
	for in, want := range tests {
		t.Run(fmt.Sprintf("%v is %v", in, want), func(t *testing.T) {
			d := &Date{
				Day: &pb.OptionalInt32{Value: in},
			}
			assert.Equal(t, want, d.IsValidDay())
		})
	}
	t.Run("nil check", func(t *testing.T) {
		d := &Date{}
		assert.False(t, d.IsValidDay())
	})
}

func TestDate_IsValidYear(t *testing.T) {
	tests := map[int32]bool{
		0:     false,
		38:    false,
		1938:  true,
		1970:  true,
		2021:  true,
		2038:  true,
		2096:  true,
		99999: false,
	}
	for in, want := range tests {
		t.Run(fmt.Sprintf("%v is %v", in, want), func(t *testing.T) {
			d := &Date{
				Year: &pb.OptionalInt32{Value: in},
			}
			assert.Equal(t, want, d.IsValidYear())
		})
	}
	t.Run("nil check", func(t *testing.T) {
		d := &Date{}
		assert.False(t, d.IsValidYear())
	})
}

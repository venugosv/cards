package date

import (
	"testing"

	pb "github.com/anzx/fabricapis/pkg/fabric/type"
	"github.com/stretchr/testify/assert"
)

func TestDate_String(t *testing.T) {
	tests := []struct {
		name string
		d    *Date
		want string
	}{
		{
			name: "2021-03-04",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: int32(2021)},
				Month: &pb.OptionalInt32{Value: int32(3)},
				Day:   &pb.OptionalInt32{Value: int32(4)},
			},
			want: "2021-03-04",
		},
		{
			name: "03-04",
			d: &Date{
				Month: &pb.OptionalInt32{Value: int32(3)},
				Day:   &pb.OptionalInt32{Value: int32(4)},
			},
			want: "03-04",
		},
		{
			name: "2021-03",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: int32(2021)},
				Month: &pb.OptionalInt32{Value: int32(3)},
			},
			want: "2021-03",
		},
		{
			name: "nil",
			d:    &Date{},
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.d.String()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestDate_Format(t *testing.T) {
	tests := []struct {
		name string
		d    *Date
		want Format
	}{
		{
			name: "2021-3-4",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: int32(2021)},
				Month: &pb.OptionalInt32{Value: int32(3)},
				Day:   &pb.OptionalInt32{Value: int32(4)},
			},
			want: YMD,
		},
		{
			name: "2021-3-4",
			d: &Date{
				Month: &pb.OptionalInt32{Value: int32(3)},
				Day:   &pb.OptionalInt32{Value: int32(4)},
			},
			want: MD,
		},
		{
			name: "2021-3",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: int32(2021)},
				Month: &pb.OptionalInt32{Value: int32(3)},
			},
			want: YM,
		},
		{
			name: "nil",
			d:    &Date{},
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.d.Format()
			assert.Equal(t, test.want, got)
		})
	}
}

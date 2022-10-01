package date

import (
	"testing"
	"time"

	"github.com/anz-bank/equals"

	"github.com/stretchr/testify/require"

	pb "github.com/anzx/fabricapis/pkg/fabric/type"
)

func TestDate_ToTime(t *testing.T) {
	tests := []struct {
		name string
		d    *Date
		want time.Time
	}{
		{
			name: "2006/6/1",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			want: time.Date(2006, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "2021/12/12",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2021},
				Month: &pb.OptionalInt32{Value: 12},
				Day:   &pb.OptionalInt32{Value: 12},
			},
			want: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "6/1",
			d: &Date{
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			want: time.Date(0, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "2006/6",
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
			},
			want: time.Date(2006, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			got, err := test.d.ToTime()
			require.NoError(t, err)
			theTime := test.want
			equals.AssertJson(t, theTime, got)
		})
	}
}

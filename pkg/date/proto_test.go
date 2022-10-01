package date

import (
	"fmt"
	"testing"
	"time"

	"github.com/anz-bank/equals"

	pb "github.com/anzx/fabricapis/pkg/fabric/type"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want *Date
	}{
		{
			name: "2006/6/1",
			want: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			in: time.Date(2006, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "2021/12/12",
			want: &Date{
				Year:  &pb.OptionalInt32{Value: 2021},
				Month: &pb.OptionalInt32{Value: 12},
				Day:   &pb.OptionalInt32{Value: 12},
			},
			in: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "6/1",
			want: &Date{
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			in: time.Date(0, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "2006/6",
			want: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			in: time.Date(2006, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			equals.AssertJson(t, test.want, New(test.in))
		})
	}
}

func TestCloneDate(t *testing.T) {
	tests := []struct {
		name string
		want *Date
	}{
		{
			name: "2006/6/1",
			want: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
		},
		{
			name: "2021/12/12",
			want: &Date{
				Year:  &pb.OptionalInt32{Value: 2021},
				Month: &pb.OptionalInt32{Value: 12},
				Day:   &pb.OptionalInt32{Value: 12},
			},
		},
		{
			name: "6/1",
			want: &Date{
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
		},
		{
			name: "2006/6",
			want: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			equals.AssertJson(t, test.want, CloneDate(test.want))
		})
	}
}

func TestToProtoAndToDate(t *testing.T) {
	tests := []struct {
		name string
		p    *pb.Date
		d    *Date
	}{
		{
			name: "2006/6/1",
			p: &pb.Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
		},
		{
			name: "2021/12/12",
			p: &pb.Date{
				Year:  &pb.OptionalInt32{Value: 2021},
				Month: &pb.OptionalInt32{Value: 12},
				Day:   &pb.OptionalInt32{Value: 12},
			},
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2021},
				Month: &pb.OptionalInt32{Value: 12},
				Day:   &pb.OptionalInt32{Value: 12},
			},
		},
		{
			name: "6/1",
			p: &pb.Date{
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			d: &Date{
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
		},
		{
			name: "2006/6",
			p: &pb.Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
			d: &Date{
				Year:  &pb.OptionalInt32{Value: 2006},
				Month: &pb.OptionalInt32{Value: 6},
				Day:   &pb.OptionalInt32{Value: 1},
			},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("proto to date %s", test.name), func(t *testing.T) {
			equals.AssertJson(t, test.d, ProtoToDate(test.p))
		})
		t.Run(fmt.Sprintf("date to proto %s", test.name), func(t *testing.T) {
			equals.AssertJson(t, test.p, test.d.ToProto())
		})
	}
}

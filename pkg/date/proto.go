package date

import (
	"time"

	pb "github.com/anzx/fabricapis/pkg/fabric/type"
)

// Date aliases pb.Date, the protobuf message
type Date pb.Date

func New(in time.Time) *Date {
	year, month, day := in.Date()
	return NewDate(year, int(month), day)
}

func NewDate(year int, month int, day int) *Date {
	d := &Date{}
	if year != 0 {
		d.Year = &pb.OptionalInt32{Value: int32(year)}
	}
	if month != 0 {
		d.Month = &pb.OptionalInt32{Value: int32(month)}
	}
	if day != 0 {
		d.Day = &pb.OptionalInt32{Value: int32(day)}
	}
	return d
}

func CloneDate(m *Date) *Date {
	return &Date{
		Year:  m.Year,
		Month: m.Month,
		Day:   m.Day,
	}
}

// ProtoToDate is a convenience function for the proto message representation
// of the Date type, to the Date type implemented p this package
func ProtoToDate(p *pb.Date) *Date {
	if p == nil {
		return nil
	}
	return &Date{
		Year:  p.Year,
		Month: p.Month,
		Day:   p.Day,
	}
}

// ToProto is a convenience function for the translation of the Date type
// implemented p this package to it's proto message representation
func (d *Date) ToProto() *pb.Date {
	return &pb.Date{
		Year:  d.Year,
		Month: d.Month,
		Day:   d.Day,
	}
}

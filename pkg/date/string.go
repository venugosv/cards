package date

import (
	"context"
	"fmt"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	pbtype "github.com/anzx/fabricapis/pkg/fabric/type"
)

type Format string

const (
	YMD      Format = "2006-1-2"
	YM       Format = "2006-1"
	MD       Format = "1-2"
	YYYYMMDD Format = "2006-01-02"
	YYMM     Format = "0601"
	YYYYMM   Format = "200601"
)

func (d *Date) Format() Format {
	if d == nil {
		return ""
	}
	switch {
	case d.Year == nil && (d.Month != nil && d.Day != nil):
		return MD
	case d.Day == nil && (d.Month != nil && d.Year != nil):
		return YM
	case d.Day != nil && d.Month != nil && d.Year != nil:
		return YMD
	default:
		return ""
	}
}

func (d *Date) String() string {
	switch d.Format() {
	case MD:
		return fmt.Sprintf("%02d-%02d", d.Month.GetValue(), d.Day.GetValue())
	case YM:
		return fmt.Sprintf("%d-%02d", d.Year.GetValue(), d.Month.GetValue())
	case YMD:
		return fmt.Sprintf("%d-%02d-%02d", d.Year.GetValue(), d.Month.GetValue(), d.Day.GetValue())
	default:
		return ""
	}
}

func GetDate(ctx context.Context, format Format, input string) *pbtype.Date {
	if input == "" {
		return nil
	}

	d, err := time.Parse(string(format), input)
	if err != nil {
		logf.Err(ctx, fmt.Errorf("unable to parse date %v", input))
		return nil
	}

	if format != YYYYMMDD {
		return NewDate(d.Year(), int(d.Month()), 0).ToProto()
	}

	return New(d).ToProto()
}

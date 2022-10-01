package date

import (
	"time"
)

func (d *Date) ToTime() (time.Time, error) {
	return time.Parse(string(d.Format()), d.String())
}

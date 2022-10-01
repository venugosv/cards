package date

import "time"

func (d *Date) IsValid() bool {
	switch {
	case d.Year == nil:
		return d.IsValidDay() && d.IsValidMonth()
	case d.Day == nil:
		return d.IsValidMonth() && d.IsValidYear()
	default:
		return d.IsValidDay() && d.IsValidMonth() && d.IsValidYear()
	}
}

func (d *Date) IsValidYear() bool {
	y := d.Year.GetValue()
	return 1000 <= y && y <= 9999
}

func (d *Date) IsValidMonth() bool {
	m := time.Month(d.Month.GetValue())
	return time.January <= m && m <= time.December
}

func (d *Date) IsValidDay() bool {
	day := d.Day.GetValue()
	return 1 <= day && day <= 31
}

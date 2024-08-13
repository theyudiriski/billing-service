package billing

import "time"

const (
	LocalTimezone = "Asia/Jakarta"
)

func CurrentLocalTime() time.Time {
	loc, _ := time.LoadLocation(LocalTimezone)
	return time.Now().In(loc)
}

func LocalTime(in time.Time) time.Time {
	loc, _ := time.LoadLocation(LocalTimezone)
	return in.In(loc)
}

package utils

import "time"

func EndOfDayMillis() int64 {
	now := time.Now()
	endOfDay := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		23,
		59,
		59,
		0,
		now.Location(),
	)
	duration := endOfDay.Sub(now)
	return duration.Milliseconds()
}

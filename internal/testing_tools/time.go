package testing_tools

import "time"

var (
	DayBefore  = Inception.AddDate(0, 0, -1)
	HourBefore = Inception.Add(-time.Hour)
	Inception  = time.Date(2024, 1, 3, 12, 3, 42, 0, time.UTC)
	HourAfter  = Inception.Add(time.Hour)
	DayAfter   = Inception.AddDate(0, 0, 1)
)

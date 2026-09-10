package repository

import "time"

// sqlite default timestamp has a special format
func parseSQLiteTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

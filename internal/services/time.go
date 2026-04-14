package services

import "time"

// timeNow returns the current date as "2006-01-02".
// Extracted as a package-level variable for testability.
var timeNow = func() string {
	return time.Now().Format("2006-01-02")
}

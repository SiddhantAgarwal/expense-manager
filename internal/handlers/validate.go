package handlers

import (
	"regexp"
	"time"
)

func validCurrency(currency string) bool {
	for _, c := range currencies {
		if c == currency {
			return true
		}
	}

	return false
}

func validFrequency(freq string) bool {
	for _, f := range frequencies {
		if f == freq {
			return true
		}
	}

	return false
}

func validDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

var numberFormats = []string{"us", "indian", "european"}

func validNumberFormat(format string) bool {
	for _, f := range numberFormats {
		if f == format {
			return true
		}
	}

	return false
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)

func validUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

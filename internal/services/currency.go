package services

import "github.com/siddhantagarwal/expense-manager/internal/models"

// ConvertToBase converts an amount from the given currency to the user's default currency
// using their manually-set exchange rates. Returns the original amount if no rate is found.
func ConvertToBase(amount float64, currency string, user models.User) float64 {
	if currency == user.DefaultCurrency {
		return amount
	}

	rate, ok := user.ExchangeRates[currency]
	if !ok {
		return amount
	}

	return amount * rate
}

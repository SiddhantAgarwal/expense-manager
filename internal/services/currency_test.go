package services

import (
	"math"
	"testing"

	"github.com/siddhantagarwal/expense-manager/internal/models"
)

func TestConvertToBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		amount   float64
		currency string
		user     models.User
		want     float64
	}{
		{
			name:     "same currency returns original amount",
			amount:   100.0,
			currency: "USD",
			user: models.User{
				DefaultCurrency: "USD",
				ExchangeRates:   map[string]float64{"EUR": 0.92},
			},
			want: 100.0,
		},
		{
			name:     "converts using exchange rate",
			amount:   50.0,
			currency: "EUR",
			user: models.User{
				DefaultCurrency: "USD",
				ExchangeRates:   map[string]float64{"EUR": 1.087},
			},
			want: 54.35,
		},
		{
			name:     "no rate returns original amount",
			amount:   100.0,
			currency: "JPY",
			user: models.User{
				DefaultCurrency: "USD",
				ExchangeRates:   map[string]float64{"EUR": 0.92},
			},
			want: 100.0,
		},
		{
			name:     "empty exchange rates returns original amount",
			amount:   75.0,
			currency: "GBP",
			user: models.User{
				DefaultCurrency: "USD",
				ExchangeRates:   map[string]float64{},
			},
			want: 75.0,
		},
		{
			name:     "zero amount",
			amount:   0.0,
			currency: "EUR",
			user: models.User{
				DefaultCurrency: "USD",
				ExchangeRates:   map[string]float64{"EUR": 1.10},
			},
			want: 0.0,
		},
		{
			name:     "large amount conversion",
			amount:   10000.0,
			currency: "INR",
			user: models.User{
				DefaultCurrency: "USD",
				ExchangeRates:   map[string]float64{"INR": 0.012},
			},
			want: 120.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ConvertToBase(tt.amount, tt.currency, tt.user)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("ConvertToBase(%v, %q, %+v) = %v, want %v", tt.amount, tt.currency, tt.user, got, tt.want)
			}
		})
	}
}

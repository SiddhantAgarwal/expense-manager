package models

import "time"

type User struct {
	Username        string             `json:"username"`
	PasswordHash    string             `json:"password_hash"`
	DefaultCurrency string             `json:"default_currency"`
	ExchangeRates   map[string]float64 `json:"exchange_rates"`
	NumberFormat    string             `json:"number_format"`
	CreatedAt       time.Time          `json:"created_at"`
}

type Expense struct {
	ID                string    `json:"id"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	AmountBase        float64   `json:"amount_base"`
	Category          string    `json:"category"`
	Description       string    `json:"description"`
	Date              string    `json:"date"`
	IsRecurring       bool      `json:"is_recurring"`
	RecurringParentID *string   `json:"recurring_parent_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type Budget struct {
	ID           string    `json:"id"`
	Category     string    `json:"category"`
	MonthlyLimit float64   `json:"monthly_limit"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
}

type RecurringExpense struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Frequency   string    `json:"frequency"` // weekly, monthly, yearly
	NextDate    string    `json:"next_date"`
	DayOfMonth  int       `json:"day_of_month"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserData struct {
	Expenses          []Expense          `json:"expenses"`
	Budgets           []Budget           `json:"budgets"`
	RecurringExpenses []RecurringExpense `json:"recurring_expenses"`
	Categories        []string           `json:"categories"`
}

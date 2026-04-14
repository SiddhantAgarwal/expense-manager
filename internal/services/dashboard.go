package services

import (
	"github.com/siddhantagarwal/expense-manager/internal/models"
)

type BudgetStatus struct {
	Category   string
	Spent      float64
	Limit      float64
	Percent    float64
	Remaining  float64
	OverBudget bool
	Alert      bool
	OverAmount float64
}

func ComputeBudgetStatuses(expenses []models.Expense, budgets []models.Budget, defaultCurrency string) []BudgetStatus {
	now := timeNow()
	currentMonth := now[:7] // "2006-01"

	// Sum spending per category this month in base currency
	spending := make(map[string]float64)

	for _, e := range expenses {
		if len(e.Date) >= 7 && e.Date[:7] == currentMonth {
			spending[e.Category] += e.AmountBase
		}
	}

	statuses := make([]BudgetStatus, 0, len(budgets))

	for _, b := range budgets {
		spent := spending[b.Category]
		limit := b.MonthlyLimit

		percent := 0.0
		if limit > 0 {
			percent = (spent / limit) * 100
		}

		status := BudgetStatus{
			Category:   b.Category,
			Spent:      spent,
			Limit:      limit,
			Percent:    percent,
			Remaining:  max(limit-spent, 0),
			OverBudget: percent > 100,
			Alert:      percent >= 80 && percent <= 100,
		}
		if status.OverBudget {
			status.OverAmount = spent - limit
		}

		statuses = append(statuses, status)
	}

	return statuses
}

func MonthlyTotal(expenses []models.Expense) float64 {
	now := timeNow()
	currentMonth := now[:7]

	var total float64

	for _, e := range expenses {
		if len(e.Date) >= 7 && e.Date[:7] == currentMonth {
			total += e.AmountBase
		}
	}

	return total
}

func RecentExpenses(expenses []models.Expense, n int) []models.Expense {
	if len(expenses) < n {
		n = len(expenses)
	}
	// Return the n most recent (expenses should already be sorted by date desc)
	recent := make([]models.Expense, n)
	copy(recent, expenses[:n])

	return recent
}

package services

import (
	"sort"

	"github.com/siddhantagarwal/expense-manager/internal/models"
)

type CategoryBreakdown struct {
	Category string
	Total    float64
	Percent  float64
	Count    int
}

type MonthSummary struct {
	Month string
	Total float64
	Count int
}

func filterByDateRange(expenses []models.Expense, from, to string) []models.Expense {
	var result []models.Expense

	for _, e := range expenses {
		if from != "" && e.Date < from {
			continue
		}

		if to != "" && e.Date > to {
			continue
		}

		result = append(result, e)
	}

	return result
}

func CategoryBreakdowns(expenses []models.Expense, from, to string) []CategoryBreakdown {
	filtered := filterByDateRange(expenses, from, to)

	totals := make(map[string]float64)
	counts := make(map[string]int)

	for _, e := range filtered {
		totals[e.Category] += e.AmountBase
		counts[e.Category]++
	}

	var grandTotal float64

	for _, t := range totals {
		grandTotal += t
	}

	breakdowns := make([]CategoryBreakdown, 0, len(totals))

	for cat, total := range totals {
		pct := 0.0

		if grandTotal > 0 {
			pct = (total / grandTotal) * 100
		}

		breakdowns = append(breakdowns, CategoryBreakdown{
			Category: cat,
			Total:    total,
			Percent:  pct,
			Count:    counts[cat],
		})
	}

	sort.Slice(breakdowns, func(i, j int) bool {
		return breakdowns[i].Total > breakdowns[j].Total
	})

	return breakdowns
}

func MonthlyTotals(expenses []models.Expense, from, to string) []MonthSummary {
	filtered := filterByDateRange(expenses, from, to)

	monthTotals := make(map[string]float64)
	monthCounts := make(map[string]int)

	for _, e := range filtered {
		if len(e.Date) < 7 {
			continue
		}

		month := e.Date[:7]
		monthTotals[month] += e.AmountBase
		monthCounts[month]++
	}

	totals := make([]MonthSummary, 0, len(monthTotals))

	for month, total := range monthTotals {
		totals = append(totals, MonthSummary{
			Month: month,
			Total: total,
			Count: monthCounts[month],
		})
	}

	sort.Slice(totals, func(i, j int) bool {
		return totals[i].Month > totals[j].Month
	})

	return totals
}

func TotalSpent(expenses []models.Expense, from, to string) float64 {
	filtered := filterByDateRange(expenses, from, to)

	var total float64

	for _, e := range filtered {
		total += e.AmountBase
	}

	return total
}

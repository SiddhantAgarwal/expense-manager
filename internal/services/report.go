package services

import (
	"sort"
	"strconv"
	"strings"

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

type MonthlyCategoryDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor string    `json:"backgroundColor"`
}

type MonthlyCategoryChart struct {
	Labels   []string                 `json:"labels"`
	Datasets []MonthlyCategoryDataset `json:"datasets"`
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

func MonthlyCategoryTotals(expenses []models.Expense, year string) MonthlyCategoryChart {
	colors := []string{
		"#4f46e5", "#06b6d4", "#16a34a", "#f59e0b",
		"#dc2626", "#8b5cf6", "#ec4899", "#14b8a6",
		"#f97316", "#6366f1", "#84cc16", "#e11d48",
	}

	labels := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	prefix := year + "-"

	catMonthTotals := make(map[string][12]float64)
	catTotals := make(map[string]float64)

	for _, e := range expenses {
		if !strings.HasPrefix(e.Date, prefix) {
			continue
		}

		if len(e.Date) < 7 {
			continue
		}

		monthStr := e.Date[5:7]

		monthIdx, err := strconv.Atoi(monthStr)
		if err != nil || monthIdx < 1 || monthIdx > 12 {
			continue
		}

		var arr [12]float64
		if existing, ok := catMonthTotals[e.Category]; ok {
			arr = existing
		}

		arr[monthIdx-1] += e.AmountBase
		catMonthTotals[e.Category] = arr
		catTotals[e.Category] += e.AmountBase
	}

	type catTotal struct {
		Category string
		Total    float64
	}

	cats := make([]catTotal, 0, len(catTotals))
	for cat, total := range catTotals {
		cats = append(cats, catTotal{Category: cat, Total: total})
	}

	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Total > cats[j].Total
	})

	datasets := make([]MonthlyCategoryDataset, 0, len(cats))
	for i, ct := range cats {
		data := make([]float64, 12)
		if arr, ok := catMonthTotals[ct.Category]; ok {
			copy(data, arr[:])
		}

		datasets = append(datasets, MonthlyCategoryDataset{
			Label:           ct.Category,
			Data:            data,
			BackgroundColor: colors[i%len(colors)],
		})
	}

	return MonthlyCategoryChart{
		Labels:   labels,
		Datasets: datasets,
	}
}

func TotalSpent(expenses []models.Expense, from, to string) float64 {
	filtered := filterByDateRange(expenses, from, to)

	var total float64

	for _, e := range filtered {
		total += e.AmountBase
	}

	return total
}

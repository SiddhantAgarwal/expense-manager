package services

import (
	"math"
	"testing"

	"github.com/siddhantagarwal/expense-manager/internal/models"
)

func TestFilterByDateRange(t *testing.T) {
	t.Parallel()

	expenses := []models.Expense{
		{ID: "1", AmountBase: 10.0, Category: "Food", Date: "2026-01-15"},
		{ID: "2", AmountBase: 20.0, Category: "Transport", Date: "2026-02-10"},
		{ID: "3", AmountBase: 30.0, Category: "Food", Date: "2026-03-05"},
		{ID: "4", AmountBase: 40.0, Category: "Housing", Date: "2026-03-20"},
		{ID: "5", AmountBase: 50.0, Category: "Food", Date: "2026-04-01"},
	}

	tests := []struct {
		name    string
		from    string
		to      string
		wantIDs []string
	}{
		{
			name:    "both bounds inclusive",
			from:    "2026-02-01",
			to:      "2026-03-20",
			wantIDs: []string{"2", "3", "4"},
		},
		{
			name:    "from only",
			from:    "2026-03-01",
			to:      "",
			wantIDs: []string{"3", "4", "5"},
		},
		{
			name:    "to only",
			from:    "",
			to:      "2026-02-10",
			wantIDs: []string{"1", "2"},
		},
		{
			name:    "no filters returns all",
			from:    "",
			to:      "",
			wantIDs: []string{"1", "2", "3", "4", "5"},
		},
		{
			name:    "range with no matches",
			from:    "2026-05-01",
			to:      "2026-06-01",
			wantIDs: nil,
		},
		{
			name:    "exact date match",
			from:    "2026-03-05",
			to:      "2026-03-05",
			wantIDs: []string{"3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := filterByDateRange(expenses, tt.from, tt.to)
			if len(got) != len(tt.wantIDs) {
				t.Errorf("filterByDateRange() returned %d items, want %d", len(got), len(tt.wantIDs))
				return
			}
			for i := range got {
				if got[i].ID != tt.wantIDs[i] {
					t.Errorf("filterByDateRange()[%d].ID = %q, want %q", i, got[i].ID, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestCategoryBreakdowns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expenses []models.Expense
		from     string
		to       string
		want     []CategoryBreakdown
	}{
		{
			name:     "empty expenses",
			expenses: []models.Expense{},
			from:     "",
			to:       "",
			want:     []CategoryBreakdown{},
		},
		{
			name: "single category",
			expenses: []models.Expense{
				{AmountBase: 100.0, Category: "Food", Date: "2026-04-01"},
				{AmountBase: 50.0, Category: "Food", Date: "2026-04-10"},
			},
			from: "",
			to:   "",
			want: []CategoryBreakdown{
				{Category: "Food", Total: 150.0, Percent: 100.0, Count: 2},
			},
		},
		{
			name: "multiple categories sorted by total descending",
			expenses: []models.Expense{
				{AmountBase: 200.0, Category: "Housing", Date: "2026-04-01"},
				{AmountBase: 100.0, Category: "Food", Date: "2026-04-05"},
				{AmountBase: 50.0, Category: "Transport", Date: "2026-04-10"},
			},
			from: "",
			to:   "",
			want: []CategoryBreakdown{
				{Category: "Housing", Total: 200.0, Percent: 57.14, Count: 1},
				{Category: "Food", Total: 100.0, Percent: 28.57, Count: 1},
				{Category: "Transport", Total: 50.0, Percent: 14.29, Count: 1},
			},
		},
		{
			name: "with date range filter",
			expenses: []models.Expense{
				{AmountBase: 100.0, Category: "Food", Date: "2026-03-01"},
				{AmountBase: 200.0, Category: "Housing", Date: "2026-04-01"},
			},
			from: "2026-04-01",
			to:   "2026-04-30",
			want: []CategoryBreakdown{
				{Category: "Housing", Total: 200.0, Percent: 100.0, Count: 1},
			},
		},
		{
			name:     "nil expenses",
			expenses: nil,
			from:     "",
			to:       "",
			want:     []CategoryBreakdown{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CategoryBreakdowns(tt.expenses, tt.from, tt.to)
			if len(got) != len(tt.want) {
				t.Errorf("CategoryBreakdowns() returned %d items, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Category != tt.want[i].Category {
					t.Errorf("breakdown[%d].Category = %q, want %q", i, got[i].Category, tt.want[i].Category)
				}
				if math.Abs(got[i].Total-tt.want[i].Total) > 0.01 {
					t.Errorf("breakdown[%d].Total = %v, want %v", i, got[i].Total, tt.want[i].Total)
				}
				if math.Abs(got[i].Percent-tt.want[i].Percent) > 0.01 {
					t.Errorf("breakdown[%d].Percent = %v, want %v", i, got[i].Percent, tt.want[i].Percent)
				}
				if got[i].Count != tt.want[i].Count {
					t.Errorf("breakdown[%d].Count = %d, want %d", i, got[i].Count, tt.want[i].Count)
				}
			}
		})
	}
}

func TestMonthlyTotals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expenses []models.Expense
		from     string
		to       string
		want     []MonthSummary
	}{
		{
			name:     "empty expenses",
			expenses: []models.Expense{},
			from:     "",
			to:       "",
			want:     []MonthSummary{},
		},
		{
			name: "groups by month sorted descending",
			expenses: []models.Expense{
				{AmountBase: 100.0, Date: "2026-01-15"},
				{AmountBase: 200.0, Date: "2026-01-20"},
				{AmountBase: 150.0, Date: "2026-02-10"},
				{AmountBase: 300.0, Date: "2026-03-05"},
			},
			from: "",
			to:   "",
			want: []MonthSummary{
				{Month: "2026-03", Total: 300.0, Count: 1},
				{Month: "2026-02", Total: 150.0, Count: 1},
				{Month: "2026-01", Total: 300.0, Count: 2},
			},
		},
		{
			name: "with date range filter",
			expenses: []models.Expense{
				{AmountBase: 100.0, Date: "2026-01-15"},
				{AmountBase: 200.0, Date: "2026-02-10"},
				{AmountBase: 300.0, Date: "2026-03-05"},
			},
			from: "2026-02-01",
			to:   "2026-02-28",
			want: []MonthSummary{
				{Month: "2026-02", Total: 200.0, Count: 1},
			},
		},
		{
			name: "date range spanning multiple months",
			expenses: []models.Expense{
				{AmountBase: 50.0, Date: "2026-01-10"},
				{AmountBase: 100.0, Date: "2026-02-10"},
				{AmountBase: 150.0, Date: "2026-03-10"},
			},
			from: "2026-02-01",
			to:   "2026-03-31",
			want: []MonthSummary{
				{Month: "2026-03", Total: 150.0, Count: 1},
				{Month: "2026-02", Total: 100.0, Count: 1},
			},
		},
		{
			name:     "nil expenses",
			expenses: nil,
			from:     "",
			to:       "",
			want:     []MonthSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MonthlyTotals(tt.expenses, tt.from, tt.to)
			if len(got) != len(tt.want) {
				t.Errorf("MonthlyTotals() returned %d items, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Month != tt.want[i].Month {
					t.Errorf("summary[%d].Month = %q, want %q", i, got[i].Month, tt.want[i].Month)
				}
				if math.Abs(got[i].Total-tt.want[i].Total) > 0.01 {
					t.Errorf("summary[%d].Total = %v, want %v", i, got[i].Total, tt.want[i].Total)
				}
				if got[i].Count != tt.want[i].Count {
					t.Errorf("summary[%d].Count = %d, want %d", i, got[i].Count, tt.want[i].Count)
				}
			}
		})
	}
}

func TestTotalSpent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expenses []models.Expense
		from     string
		to       string
		want     float64
	}{
		{
			name:     "empty expenses",
			expenses: []models.Expense{},
			from:     "",
			to:       "",
			want:     0.0,
		},
		{
			name: "sums all expenses no filter",
			expenses: []models.Expense{
				{AmountBase: 100.0, Date: "2026-01-10"},
				{AmountBase: 200.0, Date: "2026-02-15"},
				{AmountBase: 300.0, Date: "2026-03-20"},
			},
			from: "",
			to:   "",
			want: 600.0,
		},
		{
			name: "with date range filter",
			expenses: []models.Expense{
				{AmountBase: 100.0, Date: "2026-01-10"},
				{AmountBase: 200.0, Date: "2026-02-15"},
				{AmountBase: 300.0, Date: "2026-03-20"},
			},
			from: "2026-02-01",
			to:   "2026-02-28",
			want: 200.0,
		},
		{
			name: "from only includes later dates",
			expenses: []models.Expense{
				{AmountBase: 50.0, Date: "2026-01-10"},
				{AmountBase: 150.0, Date: "2026-03-10"},
			},
			from: "2026-02-01",
			to:   "",
			want: 150.0,
		},
		{
			name: "to only includes earlier dates",
			expenses: []models.Expense{
				{AmountBase: 50.0, Date: "2026-01-10"},
				{AmountBase: 150.0, Date: "2026-03-10"},
			},
			from: "",
			to:   "2026-01-31",
			want: 50.0,
		},
		{
			name: "no matching dates returns zero",
			expenses: []models.Expense{
				{AmountBase: 100.0, Date: "2026-01-10"},
			},
			from: "2026-06-01",
			to:   "2026-06-30",
			want: 0.0,
		},
		{
			name:     "nil expenses",
			expenses: nil,
			from:     "",
			to:       "",
			want:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TotalSpent(tt.expenses, tt.from, tt.to)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("TotalSpent() = %v, want %v", got, tt.want)
			}
		})
	}
}

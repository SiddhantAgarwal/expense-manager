package services

import (
	"math"
	"testing"

	"github.com/siddhantagarwal/expense-manager/internal/models"
)

func TestComputeBudgetStatuses(t *testing.T) {
	tests := []struct {
		name            string
		expenses        []models.Expense
		budgets         []models.Budget
		defaultCurrency string
		want            []BudgetStatus
	}{
		{
			name:            "under budget",
			expenses:        []models.Expense{{AmountBase: 50.0, Category: "Food", Date: "2026-04-10"}},
			budgets:         []models.Budget{{Category: "Food", MonthlyLimit: 200.0}},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 50.0, Limit: 200.0, Percent: 25.0, Remaining: 150.0, OverBudget: false, Alert: false, OverAmount: 0},
			},
		},
		{
			name: "at 80 percent triggers alert",
			expenses: []models.Expense{
				{AmountBase: 160.0, Category: "Food", Date: "2026-04-10"},
			},
			budgets:         []models.Budget{{Category: "Food", MonthlyLimit: 200.0}},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 160.0, Limit: 200.0, Percent: 80.0, Remaining: 40.0, OverBudget: false, Alert: true, OverAmount: 0},
			},
		},
		{
			name: "at exactly 100 percent triggers alert not over budget",
			expenses: []models.Expense{
				{AmountBase: 200.0, Category: "Food", Date: "2026-04-10"},
			},
			budgets:         []models.Budget{{Category: "Food", MonthlyLimit: 200.0}},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 200.0, Limit: 200.0, Percent: 100.0, Remaining: 0, OverBudget: false, Alert: true, OverAmount: 0},
			},
		},
		{
			name: "over budget sets over amount",
			expenses: []models.Expense{
				{AmountBase: 250.0, Category: "Food", Date: "2026-04-10"},
			},
			budgets:         []models.Budget{{Category: "Food", MonthlyLimit: 200.0}},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 250.0, Limit: 200.0, Percent: 125.0, Remaining: 0, OverBudget: true, Alert: false, OverAmount: 50.0},
			},
		},
		{
			name:            "no spending zero percent",
			expenses:        []models.Expense{},
			budgets:         []models.Budget{{Category: "Food", MonthlyLimit: 500.0}},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 0, Limit: 500.0, Percent: 0, Remaining: 500.0, OverBudget: false, Alert: false, OverAmount: 0},
			},
		},
		{
			name: "expenses from other months are excluded",
			expenses: []models.Expense{
				{AmountBase: 300.0, Category: "Food", Date: "2026-03-15"},
				{AmountBase: 100.0, Category: "Food", Date: "2026-04-10"},
			},
			budgets:         []models.Budget{{Category: "Food", MonthlyLimit: 200.0}},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 100.0, Limit: 200.0, Percent: 50.0, Remaining: 100.0, OverBudget: false, Alert: false, OverAmount: 0},
			},
		},
		{
			name: "multiple budgets with mixed status",
			expenses: []models.Expense{
				{AmountBase: 180.0, Category: "Food", Date: "2026-04-05"},
				{AmountBase: 400.0, Category: "Transport", Date: "2026-04-10"},
			},
			budgets: []models.Budget{
				{Category: "Food", MonthlyLimit: 200.0},
				{Category: "Transport", MonthlyLimit: 300.0},
			},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 180.0, Limit: 200.0, Percent: 90.0, Remaining: 20.0, OverBudget: false, Alert: true, OverAmount: 0},
				{Category: "Transport", Spent: 400.0, Limit: 300.0, Percent: 133.33, Remaining: 0, OverBudget: true, Alert: false, OverAmount: 100.0},
			},
		},
		{
			name:            "empty budgets returns empty",
			expenses:        []models.Expense{{AmountBase: 50.0, Category: "Food", Date: "2026-04-10"}},
			budgets:         []models.Budget{},
			defaultCurrency: "USD",
			want:            []BudgetStatus{},
		},
		{
			name: "spending in category with no budget is ignored",
			expenses: []models.Expense{
				{AmountBase: 100.0, Category: "Entertainment", Date: "2026-04-10"},
			},
			budgets:         []models.Budget{{Category: "Food", MonthlyLimit: 200.0}},
			defaultCurrency: "USD",
			want: []BudgetStatus{
				{Category: "Food", Spent: 0, Limit: 200.0, Percent: 0, Remaining: 200.0, OverBudget: false, Alert: false, OverAmount: 0},
			},
		},
	}

	originalTimeNow := timeNow

	t.Cleanup(func() { timeNow = originalTimeNow })

	timeNow = func() string { return "2026-04-19" }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeBudgetStatuses(tt.expenses, tt.budgets, tt.defaultCurrency)
			if len(got) != len(tt.want) {
				t.Errorf("ComputeBudgetStatuses() returned %d statuses, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Category != tt.want[i].Category {
					t.Errorf("status[%d].Category = %q, want %q", i, got[i].Category, tt.want[i].Category)
				}

				if math.Abs(got[i].Spent-tt.want[i].Spent) > 0.01 {
					t.Errorf("status[%d].Spent = %v, want %v", i, got[i].Spent, tt.want[i].Spent)
				}

				if math.Abs(got[i].Limit-tt.want[i].Limit) > 0.01 {
					t.Errorf("status[%d].Limit = %v, want %v", i, got[i].Limit, tt.want[i].Limit)
				}

				if math.Abs(got[i].Percent-tt.want[i].Percent) > 0.01 {
					t.Errorf("status[%d].Percent = %v, want %v", i, got[i].Percent, tt.want[i].Percent)
				}

				if math.Abs(got[i].Remaining-tt.want[i].Remaining) > 0.01 {
					t.Errorf("status[%d].Remaining = %v, want %v", i, got[i].Remaining, tt.want[i].Remaining)
				}

				if got[i].OverBudget != tt.want[i].OverBudget {
					t.Errorf("status[%d].OverBudget = %v, want %v", i, got[i].OverBudget, tt.want[i].OverBudget)
				}

				if got[i].Alert != tt.want[i].Alert {
					t.Errorf("status[%d].Alert = %v, want %v", i, got[i].Alert, tt.want[i].Alert)
				}

				if math.Abs(got[i].OverAmount-tt.want[i].OverAmount) > 0.01 {
					t.Errorf("status[%d].OverAmount = %v, want %v", i, got[i].OverAmount, tt.want[i].OverAmount)
				}
			}
		})
	}
}

func TestRecentExpenses(t *testing.T) {
	allExpenses := []models.Expense{
		{ID: "1", Description: "first", Date: "2026-04-01"},
		{ID: "2", Description: "second", Date: "2026-04-10"},
		{ID: "3", Description: "third", Date: "2026-04-15"},
		{ID: "4", Description: "fourth", Date: "2026-04-18"},
		{ID: "5", Description: "fifth", Date: "2026-04-19"},
	}

	tests := []struct {
		name     string
		expenses []models.Expense
		n        int
		want     []models.Expense
	}{
		{
			name:     "returns first n expenses",
			expenses: allExpenses,
			n:        3,
			want:     allExpenses[:3],
		},
		{
			name:     "n larger than slice returns all",
			expenses: allExpenses,
			n:        10,
			want:     allExpenses,
		},
		{
			name:     "n equals slice length returns all",
			expenses: allExpenses,
			n:        5,
			want:     allExpenses,
		},
		{
			name:     "n is zero returns empty",
			expenses: allExpenses,
			n:        0,
			want:     []models.Expense{},
		},
		{
			name:     "empty expenses returns empty",
			expenses: []models.Expense{},
			n:        5,
			want:     []models.Expense{},
		},
		{
			name:     "nil expenses returns empty",
			expenses: nil,
			n:        3,
			want:     []models.Expense{},
		},
		{
			name:     "single expense n=1",
			expenses: []models.Expense{{ID: "only", Date: "2026-04-01"}},
			n:        1,
			want:     []models.Expense{{ID: "only", Date: "2026-04-01"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RecentExpenses(tt.expenses, tt.n)
			if len(got) != len(tt.want) {
				t.Errorf("RecentExpenses() returned %d items, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("RecentExpenses()[%d].ID = %q, want %q", i, got[i].ID, tt.want[i].ID)
				}
			}
		})
	}
}

func TestMonthlyTotal(t *testing.T) {
	tests := []struct {
		name     string
		expenses []models.Expense
		want     float64
	}{
		{
			name:     "empty expenses returns zero",
			expenses: []models.Expense{},
			want:     0.0,
		},
		{
			name: "sums current month expenses",
			expenses: []models.Expense{
				{AmountBase: 50.0, Date: "2026-04-10"},
				{AmountBase: 30.0, Date: "2026-04-15"},
				{AmountBase: 20.0, Date: "2026-04-18"},
			},
			want: 100.0,
		},
		{
			name: "excludes expenses from other months",
			expenses: []models.Expense{
				{AmountBase: 100.0, Date: "2026-03-31"},
				{AmountBase: 200.0, Date: "2026-02-15"},
				{AmountBase: 50.0, Date: "2026-04-01"},
			},
			want: 50.0,
		},
		{
			name: "excludes expenses from other years",
			expenses: []models.Expense{
				{AmountBase: 75.0, Date: "2025-04-15"},
				{AmountBase: 25.0, Date: "2026-04-15"},
			},
			want: 25.0,
		},
		{
			name:     "nil expenses returns zero",
			expenses: nil,
			want:     0.0,
		},
		{
			name: "handles single current month expense",
			expenses: []models.Expense{
				{AmountBase: 999.99, Date: "2026-04-01"},
			},
			want: 999.99,
		},
		{
			name: "all expenses from other months returns zero",
			expenses: []models.Expense{
				{AmountBase: 50.0, Date: "2026-03-31"},
				{AmountBase: 75.0, Date: "2026-01-10"},
			},
			want: 0.0,
		},
	}

	originalTimeNow := timeNow

	t.Cleanup(func() { timeNow = originalTimeNow })

	timeNow = func() string { return "2026-04-19" }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MonthlyTotal(tt.expenses)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("MonthlyTotal() = %v, want %v", got, tt.want)
			}
		})
	}
}

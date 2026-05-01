package handlers

import (
	"log"
	"net/http"
	"sort"

	"github.com/siddhantagarwal/expense-manager/internal/middleware"
	"github.com/siddhantagarwal/expense-manager/internal/models"
	"github.com/siddhantagarwal/expense-manager/internal/services"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	budgetStatuses := services.ComputeBudgetStatuses(ud.Expenses, ud.Budgets, user.DefaultCurrency)
	monthlyTotal := services.MonthlyTotal(ud.Expenses)

	// Sort expenses by date descending for recent list
	sort.Slice(ud.Expenses, func(i, j int) bool {
		return ud.Expenses[i].Date > ud.Expenses[j].Date
	})
	recent := services.RecentExpenses(ud.Expenses, 5)

	budgetsOnTrack := 0

	for _, bs := range budgetStatuses {
		if !bs.OverBudget && !bs.Alert {
			budgetsOnTrack++
		}
	}

	data := struct {
		Username        string
		DefaultCurrency string
		NumberFormat    string
		MonthlyTotal    float64
		RecentExpenses  []models.Expense
		BudgetStatuses  []services.BudgetStatus
		BudgetsOnTrack  int
		TotalBudgets    int
	}{
		Username:        username,
		DefaultCurrency: user.DefaultCurrency,
		NumberFormat:    user.NumberFormat,
		MonthlyTotal:    monthlyTotal,
		RecentExpenses:  recent,
		BudgetStatuses:  budgetStatuses,
		BudgetsOnTrack:  budgetsOnTrack,
		TotalBudgets:    len(ud.Budgets),
	}

	if err := h.templates["dashboard"].ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Println(err)
	}
}

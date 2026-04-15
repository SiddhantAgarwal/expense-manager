package handlers

import (
	"log"
	"net/http"

	"github.com/siddhantagarwal/expense-manager/internal/middleware"
	"github.com/siddhantagarwal/expense-manager/internal/services"
)

type reportListData struct {
	Username           string
	DefaultCurrency    string
	CategoryBreakdowns []services.CategoryBreakdown
	MonthlyTotals      []services.MonthSummary
	TotalSpent         float64
	FilterFrom         string
	FilterTo           string
	HasData            bool
}

func (h *Handlers) ReportList(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	categoryBreakdowns := services.CategoryBreakdowns(ud.Expenses, from, to)
	monthlyTotals := services.MonthlyTotals(ud.Expenses, from, to)
	totalSpent := services.TotalSpent(ud.Expenses, from, to)

	data := reportListData{
		Username:           username,
		DefaultCurrency:    user.DefaultCurrency,
		CategoryBreakdowns: categoryBreakdowns,
		MonthlyTotals:      monthlyTotals,
		TotalSpent:         totalSpent,
		FilterFrom:         from,
		FilterTo:           to,
		HasData:            len(ud.Expenses) > 0,
	}

	if err := h.templates["reports"].ExecuteTemplate(w, "reports.html", data); err != nil {
		log.Println(err)
	}
}

package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/siddhantagarwal/expense-manager/internal/middleware"
	"github.com/siddhantagarwal/expense-manager/internal/models"
	"github.com/siddhantagarwal/expense-manager/internal/services"
)

type budgetListData struct {
	Username        string
	DefaultCurrency string
	NumberFormat    string
	Budgets         []models.Budget
	BudgetStatuses  []services.BudgetStatus
	Categories      []string
	Currencies      []string
}

func (h *Handlers) BudgetList(w http.ResponseWriter, r *http.Request) {
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
	budgetStatuses := services.ComputeBudgetStatuses(ud.Expenses, ud.Budgets, user.DefaultCurrency)

	data := budgetListData{
		Username:        username,
		DefaultCurrency: user.DefaultCurrency,
		NumberFormat:    user.NumberFormat,
		Budgets:         ud.Budgets,
		BudgetStatuses:  budgetStatuses,
		Categories:      ud.Categories,
		Currencies:      currencies,
	}

	if err := h.templates["budgets"].ExecuteTemplate(w, "budgets.html", data); err != nil {
		log.Println(err)
	}
}

func (h *Handlers) BudgetCreate(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())

	limit, err := strconv.ParseFloat(r.FormValue("monthly_limit"), 64)
	if err != nil || limit <= 0 {
		http.Error(w, "invalid monthly limit", http.StatusBadRequest)
		return
	}

	category := r.FormValue("category")
	if category == "" {
		http.Error(w, "category is required", http.StatusBadRequest)
		return
	}

	currency := r.FormValue("currency")
	if !validCurrency(currency) {
		http.Error(w, "invalid currency", http.StatusBadRequest)
		return
	}

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Check if budget already exists for this category — update if so
	for i, b := range ud.Budgets {
		if b.Category == category {
			ud.Budgets[i].MonthlyLimit = limit
			ud.Budgets[i].Currency = currency

			if err := h.store.SaveUserData(username, ud); err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/budgets", http.StatusSeeOther)

			return
		}
	}

	budget := models.Budget{
		ID:           services.NewID(),
		Category:     category,
		MonthlyLimit: limit,
		Currency:     currency,
		CreatedAt:    time.Now(),
	}

	ud.Budgets = append(ud.Budgets, budget)

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/budgets", http.StatusSeeOther)
}

func (h *Handlers) BudgetDelete(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	vars := mux.Vars(r)
	id := vars["id"]

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	filtered := make([]models.Budget, 0, len(ud.Budgets))
	found := false

	for _, b := range ud.Budgets {
		if b.ID == id {
			found = true
			continue
		}

		filtered = append(filtered, b)
	}

	if !found {
		http.Error(w, "budget not found", http.StatusNotFound)
		return
	}

	ud.Budgets = filtered

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

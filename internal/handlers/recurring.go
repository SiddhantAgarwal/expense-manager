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

var frequencies = []string{"weekly", "monthly", "yearly"}

type recurringListData struct {
	Username          string
	DefaultCurrency   string
	NumberFormat      string
	RecurringExpenses []models.RecurringExpense
	Categories        []string
	Currencies        []string
	Frequencies       []string
}

func (h *Handlers) RecurringList(w http.ResponseWriter, r *http.Request) {
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

	data := recurringListData{
		Username:          username,
		DefaultCurrency:   users[username].DefaultCurrency,
		NumberFormat:      users[username].NumberFormat,
		RecurringExpenses: ud.RecurringExpenses,
		Categories:        ud.Categories,
		Currencies:        currencies,
		Frequencies:       frequencies,
	}

	if err := h.templates["recurring"].ExecuteTemplate(w, "recurring.html", data); err != nil {
		log.Println(err)
	}
}

func (h *Handlers) RecurringCreate(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	currency := r.FormValue("currency")
	if !validCurrency(currency) {
		http.Error(w, "invalid currency", http.StatusBadRequest)
		return
	}

	category := r.FormValue("category")
	if category == "" {
		http.Error(w, "category is required", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")

	frequency := r.FormValue("frequency")
	if !validFrequency(frequency) {
		http.Error(w, "invalid frequency", http.StatusBadRequest)
		return
	}

	startDate := r.FormValue("start_date")
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	} else if !validDate(startDate) {
		http.Error(w, "invalid start date", http.StatusBadRequest)
		return
	}

	dayOfMonthStr := r.FormValue("day_of_month")

	dayOfMonth := 0
	if dayOfMonthStr != "" {
		dayOfMonth, err = strconv.Atoi(dayOfMonthStr)
		if err != nil || dayOfMonth < 1 || dayOfMonth > 31 {
			http.Error(w, "day of month must be 1-31", http.StatusBadRequest)
			return
		}
	}

	if dayOfMonth == 0 {
		t, _ := time.Parse("2006-01-02", startDate)
		dayOfMonth = t.Day()
	}

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	recurring := models.RecurringExpense{
		ID:          services.NewID(),
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		Description: description,
		Frequency:   frequency,
		NextDate:    startDate,
		DayOfMonth:  dayOfMonth,
		Active:      true,
		CreatedAt:   time.Now(),
	}

	ud.RecurringExpenses = append(ud.RecurringExpenses, recurring)

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/recurring", http.StatusSeeOther)
}

func (h *Handlers) RecurringUpdate(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	vars := mux.Vars(r)
	id := vars["id"]

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	found := false

	for i, re := range ud.RecurringExpenses {
		if re.ID == id {
			ud.RecurringExpenses[i].Active = !ud.RecurringExpenses[i].Active
			found = true

			break
		}
	}

	if !found {
		http.Error(w, "recurring expense not found", http.StatusNotFound)
		return
	}

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/recurring")
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) RecurringDelete(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	vars := mux.Vars(r)
	id := vars["id"]

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	filtered := make([]models.RecurringExpense, 0, len(ud.RecurringExpenses))
	found := false

	for _, re := range ud.RecurringExpenses {
		if re.ID == id {
			found = true
			continue
		}

		filtered = append(filtered, re)
	}

	if !found {
		http.Error(w, "recurring expense not found", http.StatusNotFound)
		return
	}

	ud.RecurringExpenses = filtered

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

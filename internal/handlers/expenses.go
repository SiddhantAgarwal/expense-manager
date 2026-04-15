package handlers

import (
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/siddhantagarwal/expense-manager/internal/middleware"
	"github.com/siddhantagarwal/expense-manager/internal/models"
	"github.com/siddhantagarwal/expense-manager/internal/services"
)

var currencies = []string{"USD", "EUR", "GBP", "INR", "JPY", "CAD", "AUD"}

type expenseListData struct {
	Username        string
	Expenses        []models.Expense
	Categories      []string
	DefaultCurrency string
	FilterFrom      string
	FilterTo        string
	FilterCategory  string
}

func (h *Handlers) ExpenseList(w http.ResponseWriter, r *http.Request) {
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
	category := r.URL.Query().Get("category")

	filtered := filterExpenses(ud.Expenses, from, to, category)

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date > filtered[j].Date
	})

	data := expenseListData{
		Username:        username,
		Expenses:        filtered,
		Categories:      ud.Categories,
		DefaultCurrency: user.DefaultCurrency,
		FilterFrom:      from,
		FilterTo:        to,
		FilterCategory:  category,
	}

	if err := h.templates["expenses"].ExecuteTemplate(w, "expenses.html", data); err != nil {
		log.Println(err)
	}
}

func filterExpenses(expenses []models.Expense, from, to, category string) []models.Expense {
	var result []models.Expense

	for _, e := range expenses {
		if from != "" && e.Date < from {
			continue
		}

		if to != "" && e.Date > to {
			continue
		}

		if category != "" && e.Category != category {
			continue
		}

		result = append(result, e)
	}

	return result
}

type expenseFormData struct {
	Username        string
	IsEdit          bool
	Expense         models.Expense
	Categories      []string
	Currencies      []string
	DefaultCurrency string
	Today           string
}

func (h *Handlers) ExpenseNew(w http.ResponseWriter, r *http.Request) {
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

	data := expenseFormData{
		Username:        username,
		IsEdit:          false,
		Categories:      ud.Categories,
		Currencies:      currencies,
		DefaultCurrency: users[username].DefaultCurrency,
		Today:           time.Now().Format("2006-01-02"),
	}

	if err := h.templates["expense_form"].ExecuteTemplate(w, "expense_form.html", data); err != nil {
		log.Println(err)
	}
}

func (h *Handlers) ExpenseCreate(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	currency := r.FormValue("currency")
	category := r.FormValue("category")
	date := r.FormValue("date")
	description := r.FormValue("description")

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]

	amountBase := services.ConvertToBase(amount, currency, user)

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	expense := models.Expense{
		ID:          services.NewID(),
		Amount:      amount,
		Currency:    currency,
		AmountBase:  amountBase,
		Category:    category,
		Description: description,
		Date:        date,
		CreatedAt:   time.Now(),
	}

	ud.Expenses = append(ud.Expenses, expense)

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *Handlers) ExpenseEdit(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	vars := mux.Vars(r)
	id := vars["id"]

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var expense models.Expense

	found := false

	for _, e := range ud.Expenses {
		if e.ID == id {
			expense = e
			found = true

			break
		}
	}

	if !found {
		http.Error(w, "expense not found", http.StatusNotFound)
		return
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username        string
		Expense         models.Expense
		Categories      []string
		Currencies      []string
		DefaultCurrency string
	}{
		Username:        username,
		Expense:         expense,
		Categories:      ud.Categories,
		Currencies:      currencies,
		DefaultCurrency: users[username].DefaultCurrency,
	}

	if err := h.templates["expense_edit_partial"].ExecuteTemplate(w, "expense_edit_partial", data); err != nil {
		log.Println(err)
	}
}

func (h *Handlers) ExpenseUpdate(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	vars := mux.Vars(r)
	id := vars["id"]

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	currency := r.FormValue("currency")
	category := r.FormValue("category")
	date := r.FormValue("date")
	description := r.FormValue("description")

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]
	amountBase := services.ConvertToBase(amount, currency, user)

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	found := false

	for i, e := range ud.Expenses {
		if e.ID == id {
			ud.Expenses[i] = models.Expense{
				ID:                e.ID,
				Amount:            amount,
				Currency:          currency,
				AmountBase:        amountBase,
				Category:          category,
				Description:       description,
				Date:              date,
				IsRecurring:       e.IsRecurring,
				RecurringParentID: e.RecurringParentID,
				CreatedAt:         e.CreatedAt,
			}
			found = true

			break
		}
	}

	if !found {
		http.Error(w, "expense not found", http.StatusNotFound)
		return
	}

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/expenses")
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) ExpenseDelete(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	vars := mux.Vars(r)
	id := vars["id"]

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	filtered := make([]models.Expense, 0, len(ud.Expenses))
	found := false

	for _, e := range ud.Expenses {
		if e.ID == id {
			found = true
			continue
		}

		filtered = append(filtered, e)
	}

	if !found {
		http.Error(w, "expense not found", http.StatusNotFound)
		return
	}

	ud.Expenses = filtered
	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// redirectSafe redirects to a URL, ensuring it's a relative path
// nolint:unused
func redirectSafe(w http.ResponseWriter, r *http.Request, rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host != "" {
		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, u.String(), http.StatusSeeOther)
}

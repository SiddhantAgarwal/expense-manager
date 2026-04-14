package handlers

import (
	"html/template"
	"log"
	"net/http"
	"sort"

	"github.com/siddhantagarwal/expense-manager/internal/auth"
	"github.com/siddhantagarwal/expense-manager/internal/middleware"
	"github.com/siddhantagarwal/expense-manager/internal/models"
	"github.com/siddhantagarwal/expense-manager/internal/services"
	"github.com/siddhantagarwal/expense-manager/internal/store"
)

type Handlers struct {
	store     *store.Store
	auth      *auth.Auth
	templates map[string]*template.Template
}

func New(st *store.Store, au *auth.Auth, tmplDir string) (*Handlers, error) {
	// Each page template defines the same block names ("title", "content").
	// ParseGlob would load them into one set and the last definition wins,
	// causing every page to render the same content. Instead, parse each
	// page template into its own isolated template set with base.html.
	pageTemplates := []string{
		"login", "signup", "dashboard", "expenses", "expense_form",
	}

	templates := make(map[string]*template.Template)

	for _, page := range pageTemplates {
		t, err := template.ParseFiles(
			tmplDir+"/base.html",
			tmplDir+"/"+page+".html",
		)
		if err != nil {
			return nil, err
		}

		templates[page] = t
	}

	// Partial templates that don't extend base.html
	t, err := template.ParseFiles(tmplDir + "/expense_edit_partial.html")
	if err != nil {
		return nil, err
	}

	templates["expense_edit_partial"] = t

	return &Handlers{
		store:     st,
		auth:      au,
		templates: templates,
	}, nil
}

type authData struct {
	Username string
	Error    string
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := h.templates["login"].ExecuteTemplate(w, "login.html", authData{})
		log.Println(err)

		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, ok := users[username]
	if !ok || !auth.CheckPassword(password, user.PasswordHash) {
		_ = h.templates["login"].ExecuteTemplate(w, "login.html", authData{Error: "Invalid username or password"})
		return
	}

	token := h.auth.CreateSession(username)
	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_ = h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	currency := r.FormValue("currency")

	if len(username) < 3 {
		_ = h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: "Username must be at least 3 characters"})
		return
	}

	if len(password) < 6 {
		_ = h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: "Password must be at least 6 characters"})
		return
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, exists := users[username]; exists {
		_ = h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: "Username already taken"})
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	users[username] = models.User{
		Username:        username,
		PasswordHash:    hash,
		DefaultCurrency: currency,
		ExchangeRates:   map[string]float64{},
	}

	if err := h.store.SaveUsers(users); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Initialize user data file
	userData := &models.UserData{
		Expenses:          []models.Expense{},
		Budgets:           []models.Budget{},
		RecurringExpenses: []models.RecurringExpense{},
		Categories:        []string{"Food", "Transport", "Housing", "Entertainment", "Health", "Other"},
	}
	if err := h.store.SaveUserData(username, userData); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token := h.auth.CreateSession(username)
	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := auth.GetSessionToken(r)
	if token != "" {
		h.auth.DeleteSession(token)
	}

	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

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
		MonthlyTotal    float64
		RecentExpenses  []models.Expense
		BudgetStatuses  []services.BudgetStatus
		BudgetsOnTrack  int
		TotalBudgets    int
	}{
		Username:        username,
		DefaultCurrency: user.DefaultCurrency,
		MonthlyTotal:    monthlyTotal,
		RecentExpenses:  recent,
		BudgetStatuses:  budgetStatuses,
		BudgetsOnTrack:  budgetsOnTrack,
		TotalBudgets:    len(ud.Budgets),
	}

	_ = h.templates["dashboard"].ExecuteTemplate(w, "dashboard.html", data)
}

package handlers

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"

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
		"login", "signup", "dashboard", "expenses", "expense_form", "budgets", "recurring", "reports", "settings",
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

	t, err = template.ParseFiles(tmplDir + "/error.html")
	if err != nil {
		return nil, err
	}

	templates["error"] = t

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

func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) error {
	return h.templates["error"].ExecuteTemplate(w, "error.html", map[string]string{
		"Title":   "Page Not Found",
		"Message": "The page you're looking for doesn't exist.",
	})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := h.templates["login"].ExecuteTemplate(w, "login.html", authData{}); err != nil {
			log.Println(err)
		}

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
		if err := h.templates["login"].ExecuteTemplate(w, "login.html", authData{Error: "Invalid username or password"}); err != nil {
			log.Println(err)
		}
		return
	}

	token := h.auth.CreateSession(username)
	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{}); err != nil {
			log.Println(err)
		}
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	currency := r.FormValue("currency")

	if !validUsername(username) {
		if err := h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: "Username must be 3-50 characters (letters, numbers, underscores)"}); err != nil {
			log.Println(err)
		}
		return
	}

	if len(password) < 6 {
		if err := h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: "Password must be at least 6 characters"}); err != nil {
			log.Println(err)
		}
		return
	}

	if len(password) > 72 {
		if err := h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: "Password must be at most 72 characters"}); err != nil {
			log.Println(err)
		}
		return
	}

	if !validCurrency(currency) {
		currency = "USD"
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, exists := users[username]; exists {
		if err := h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: "Username already taken"}); err != nil {
			log.Println(err)
		}
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
		CreatedAt:       time.Now(),
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

	if err := h.templates["dashboard"].ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Println(err)
	}
}

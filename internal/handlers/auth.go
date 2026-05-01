package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/siddhantagarwal/expense-manager/internal/auth"
	"github.com/siddhantagarwal/expense-manager/internal/models"
)

type authData struct {
	Username string
	Error    string
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

	token := h.auth.Store.Create(username)
	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handlers) signupError(w http.ResponseWriter, errMsg string) {
	if err := h.templates["signup"].ExecuteTemplate(w, "signup.html", authData{Error: errMsg}); err != nil {
		log.Println(err)
	}
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.signupError(w, "")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	currency := r.FormValue("currency")

	if !validUsername(username) {
		h.signupError(w, "Username must be 3-50 characters (letters, numbers, underscores)")
		return
	}

	if len(password) < 6 {
		h.signupError(w, "Password must be at least 6 characters")
		return
	}

	if len(password) > 72 {
		h.signupError(w, "Password must be at most 72 characters")
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
		h.signupError(w, "Username already taken")
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
		NumberFormat:    "us",
		CreatedAt:       time.Now(),
	}

	if err := h.store.SaveUsers(users); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

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

	token := h.auth.Store.Create(username)
	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := auth.GetSessionToken(r)
	if token != "" {
		h.auth.Store.Delete(token)
	}

	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

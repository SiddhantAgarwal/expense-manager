package handlers

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/siddhantagarwal/expense-manager/internal/auth"
	"github.com/siddhantagarwal/expense-manager/internal/middleware"
)

type settingsData struct {
	Username        string
	DefaultCurrency string
	NumberFormat    string
	NumberFormats   []numberFormatOption
	ExchangeRates   map[string]float64
	SortedRates     []exchangeRateEntry
	Currencies      []string
	Categories      []string
	Error           string
	Success         string
}

type exchangeRateEntry struct {
	Currency string
	Rate     float64
}

type numberFormatOption struct {
	Value      string
	Label      string
	IsSelected bool
}

func (h *Handlers) SettingsPage(w http.ResponseWriter, r *http.Request) {
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

	// Sort exchange rates for consistent display
	sortedRates := make([]exchangeRateEntry, 0, len(user.ExchangeRates))
	for cur, rate := range user.ExchangeRates {
		sortedRates = append(sortedRates, exchangeRateEntry{Currency: cur, Rate: rate})
	}

	sort.Slice(sortedRates, func(i, j int) bool {
		return sortedRates[i].Currency < sortedRates[j].Currency
	})

	currentFormat := user.NumberFormat
	if currentFormat == "" {
		currentFormat = "us"
	}

	allFormats := []numberFormatOption{
		{Value: "us", Label: "US (1,234,567.89)"},
		{Value: "indian", Label: "Indian (12,34,567.89)"},
		{Value: "european", Label: "European (1.234.567,89)"},
	}
	for i := range allFormats {
		allFormats[i].IsSelected = allFormats[i].Value == currentFormat
	}

	data := settingsData{
		Username:        username,
		DefaultCurrency: user.DefaultCurrency,
		NumberFormat:    user.NumberFormat,
		NumberFormats:   allFormats,
		ExchangeRates:   user.ExchangeRates,
		SortedRates:     sortedRates,
		Currencies:      currencies,
		Categories:      ud.Categories,
		Error:           r.URL.Query().Get("error"),
		Success:         r.URL.Query().Get("success"),
	}

	if err := h.templates["settings"].ExecuteTemplate(w, "settings.html", data); err != nil {
		log.Println(err)
	}
}

func (h *Handlers) SettingsUpdate(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	action := r.FormValue("action")

	switch action {
	case "currency":
		h.updateCurrency(w, r, username)
	case "number_format":
		h.updateNumberFormat(w, r, username)
	case "exchange_rate":
		h.addExchangeRate(w, r, username)
	case "add_category":
		h.addCategory(w, r, username)
	case "password":
		h.updatePassword(w, r, username)
	default:
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	}
}

func (h *Handlers) SettingsDeleteRate(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	h.deleteExchangeRate(w, r, username)
}

func (h *Handlers) SettingsDeleteCategory(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.FromContext(r.Context())
	h.deleteCategory(w, r, username)
}

func (h *Handlers) updateCurrency(w http.ResponseWriter, r *http.Request, username string) {
	currency := r.FormValue("default_currency")
	if !validCurrency(currency) {
		http.Redirect(w, r, "/settings?error=invalid_currency", http.StatusSeeOther)
		return
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]
	user.DefaultCurrency = currency
	users[username] = user

	if err := h.store.SaveUsers(users); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?success=currency", http.StatusSeeOther)
}

func (h *Handlers) updateNumberFormat(w http.ResponseWriter, r *http.Request, username string) {
	format := r.FormValue("number_format")
	if !validNumberFormat(format) {
		http.Redirect(w, r, "/settings?error=invalid_number_format", http.StatusSeeOther)
		return
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]
	user.NumberFormat = format
	users[username] = user

	if err := h.store.SaveUsers(users); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?success=number_format", http.StatusSeeOther)
}

func (h *Handlers) addExchangeRate(w http.ResponseWriter, r *http.Request, username string) {
	currency := r.FormValue("currency")
	if !validCurrency(currency) {
		http.Redirect(w, r, "/settings?error=invalid_currency", http.StatusSeeOther)
		return
	}

	rate, err := strconv.ParseFloat(r.FormValue("rate"), 64)
	if err != nil || rate <= 0 {
		http.Redirect(w, r, "/settings?error=invalid_rate", http.StatusSeeOther)
		return
	}

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]
	if user.ExchangeRates == nil {
		user.ExchangeRates = make(map[string]float64)
	}

	user.ExchangeRates[currency] = rate
	users[username] = user

	if err := h.store.SaveUsers(users); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?success=rate", http.StatusSeeOther)
}

func (h *Handlers) deleteExchangeRate(w http.ResponseWriter, r *http.Request, username string) {
	vars := mux.Vars(r)
	currency := vars["currency"]

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]
	delete(user.ExchangeRates, currency)
	users[username] = user

	if err := h.store.SaveUsers(users); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?success=rate_deleted", http.StatusSeeOther)
}

func (h *Handlers) addCategory(w http.ResponseWriter, r *http.Request, username string) {
	category := strings.TrimSpace(r.FormValue("category"))
	if category == "" {
		http.Redirect(w, r, "/settings?error=empty_category", http.StatusSeeOther)
		return
	}

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	for _, c := range ud.Categories {
		if strings.EqualFold(c, category) {
			http.Redirect(w, r, "/settings?error=duplicate_category", http.StatusSeeOther)
			return
		}
	}

	ud.Categories = append(ud.Categories, category)

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?success=category", http.StatusSeeOther)
}

func (h *Handlers) deleteCategory(w http.ResponseWriter, r *http.Request, username string) {
	vars := mux.Vars(r)
	category := vars["category"]

	ud, err := h.store.LoadUserData(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Check if category is in use
	for _, e := range ud.Expenses {
		if e.Category == category {
			http.Redirect(w, r, "/settings?error=category_in_use", http.StatusSeeOther)
			return
		}
	}

	for i, c := range ud.Categories {
		if c == category {
			ud.Categories = append(ud.Categories[:i], ud.Categories[i+1:]...)
			break
		}
	}

	if err := h.store.SaveUserData(username, ud); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?success=category_deleted", http.StatusSeeOther)
}

func (h *Handlers) updatePassword(w http.ResponseWriter, r *http.Request, username string) {
	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	users, err := h.store.LoadUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := users[username]

	if !auth.CheckPassword(currentPassword, user.PasswordHash) {
		http.Redirect(w, r, "/settings?error=wrong_password", http.StatusSeeOther)
		return
	}

	if len(newPassword) < 6 {
		http.Redirect(w, r, "/settings?error=short_password", http.StatusSeeOther)
		return
	}

	if newPassword != confirmPassword {
		http.Redirect(w, r, "/settings?error=password_mismatch", http.StatusSeeOther)
		return
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = hash
	users[username] = user

	if err := h.store.SaveUsers(users); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?success=password", http.StatusSeeOther)
}

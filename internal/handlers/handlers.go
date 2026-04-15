package handlers

import (
	"html/template"
	"net/http"

	"github.com/siddhantagarwal/expense-manager/internal/auth"
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

func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) error {
	return h.templates["error"].ExecuteTemplate(w, "error.html", map[string]string{
		"Title":   "Page Not Found",
		"Message": "The page you're looking for doesn't exist.",
	})
}

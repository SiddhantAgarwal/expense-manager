package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/siddhantagarwal/expense-manager/internal/auth"
	"github.com/siddhantagarwal/expense-manager/internal/embedded"
	"github.com/siddhantagarwal/expense-manager/internal/handlers"
	"github.com/siddhantagarwal/expense-manager/internal/middleware"
	"github.com/siddhantagarwal/expense-manager/internal/services"
	"github.com/siddhantagarwal/expense-manager/internal/store"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	st := store.New(dataDir)
	au := auth.New()

	h, err := handlers.New(st, au, embedded.Templates)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	r := mux.NewRouter()

	// Static files (embedded)
	staticSub, err := fs.Sub(embedded.Static, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub-filesystem: %v", err)
	}

	staticFS := http.FileServer(http.FS(staticSub))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFS))

	// Public routes
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	}).Methods("GET")
	r.HandleFunc("/login", h.Login).Methods("GET", "POST")
	r.HandleFunc("/signup", h.Signup).Methods("GET", "POST")
	r.HandleFunc("/logout", h.Logout).Methods("POST")

	// Protected routes
	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware(au))
	protected.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
	protected.HandleFunc("/expenses", h.ExpenseList).Methods("GET")
	protected.HandleFunc("/expenses/export", h.ExpenseExportCSV).Methods("GET")
	protected.HandleFunc("/expenses/new", h.ExpenseNew).Methods("GET")
	protected.HandleFunc("/expenses", h.ExpenseCreate).Methods("POST")
	protected.HandleFunc("/expenses/{id}", h.ExpenseUpdate).Methods("PUT")
	protected.HandleFunc("/expenses/{id}", h.ExpenseDelete).Methods("DELETE")
	protected.HandleFunc("/expenses/{id}/edit", h.ExpenseEdit).Methods("GET")
	protected.HandleFunc("/budgets", h.BudgetList).Methods("GET")
	protected.HandleFunc("/budgets", h.BudgetCreate).Methods("POST")
	protected.HandleFunc("/budgets/{id}", h.BudgetDelete).Methods("DELETE")

	// Recurring expense routes
	protected.HandleFunc("/recurring", h.RecurringList).Methods("GET")
	protected.HandleFunc("/recurring", h.RecurringCreate).Methods("POST")
	protected.HandleFunc("/recurring/{id}", h.RecurringUpdate).Methods("PUT")
	protected.HandleFunc("/recurring/{id}", h.RecurringDelete).Methods("DELETE")

	// Report routes
	protected.HandleFunc("/reports", h.ReportList).Methods("GET")

	// Settings routes
	protected.HandleFunc("/settings", h.SettingsPage).Methods("GET")
	protected.HandleFunc("/settings", h.SettingsUpdate).Methods("POST")
	protected.HandleFunc("/settings/rates/{currency}", h.SettingsDeleteRate).Methods("DELETE")
	protected.HandleFunc("/settings/categories/{category}", h.SettingsDeleteCategory).Methods("DELETE")

	// Custom 404 handler
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = h.NotFound(w, r)
	})

	// Start recurring expense processor background goroutine
	recurringSvc := services.NewRecurringProcessor(st)
	recurringSvc.Start(ctx)

	log.Printf("Starting expense manager %s (commit %s, built %s) on :%s", version, commit, date, port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server forced to shutdown: %s\n", err)
	}

	log.Println("Expense manager exited.")
}

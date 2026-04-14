package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/siddhantagarwal/expense-manager/internal/auth"
	"github.com/siddhantagarwal/expense-manager/internal/handlers"
	"github.com/siddhantagarwal/expense-manager/internal/middleware"
	"github.com/siddhantagarwal/expense-manager/internal/store"
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

	st := store.New(dataDir)
	au := auth.New()

	h, err := handlers.New(st, au, "templates")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	r := mux.NewRouter()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

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
	protected.HandleFunc("/expenses/new", h.ExpenseNew).Methods("GET")
	protected.HandleFunc("/expenses", h.ExpenseCreate).Methods("POST")
	protected.HandleFunc("/expenses/{id}", h.ExpenseUpdate).Methods("PUT")
	protected.HandleFunc("/expenses/{id}", h.ExpenseDelete).Methods("DELETE")
	protected.HandleFunc("/expenses/{id}/edit", h.ExpenseEdit).Methods("GET")

	log.Printf("Starting expense manager on :%s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

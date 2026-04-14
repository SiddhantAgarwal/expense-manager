package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/siddhantagarwal/expense-manager/internal/models"
)

type Store struct {
	mu       sync.Mutex
	dataDir  string
	userFile string
}

func New(dataDir string) *Store {
	return &Store{
		dataDir:  dataDir,
		userFile: filepath.Join(dataDir, "users.json"),
	}
}

func (s *Store) LoadUsers() (map[string]models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.userFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]models.User), nil
		}

		return nil, err
	}

	var users map[string]models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Store) SaveUsers(users map[string]models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.userFile, data, 0644)
}

func (s *Store) LoadUserData(username string) (*models.UserData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dataDir, username+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.UserData{
				Expenses:          []models.Expense{},
				Budgets:           []models.Budget{},
				RecurringExpenses: []models.RecurringExpense{},
				Categories:        defaultCategories(),
			}, nil
		}

		return nil, err
	}

	var ud models.UserData
	if err := json.Unmarshal(data, &ud); err != nil {
		return nil, err
	}

	return &ud, nil
}

func (s *Store) SaveUserData(username string, ud *models.UserData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(ud, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dataDir, username+".json")

	return os.WriteFile(path, data, 0644)
}

func defaultCategories() []string {
	return []string{"Food", "Transport", "Housing", "Entertainment", "Health", "Other"}
}

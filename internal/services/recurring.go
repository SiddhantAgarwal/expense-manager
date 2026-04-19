package services

import (
	"context"
	"log"
	"time"

	"github.com/siddhantagarwal/expense-manager/internal/models"
	"github.com/siddhantagarwal/expense-manager/internal/store"
)

// RecurringProcessor checks for due recurring expenses and auto-creates
// expense entries. It runs as a background goroutine.
type RecurringProcessor struct {
	store *store.Store
}

// NewRecurringProcessor creates a new processor backed by the given store.
func NewRecurringProcessor(st *store.Store) *RecurringProcessor {
	return &RecurringProcessor{store: st}
}

// Start launches the background goroutine that processes recurring expenses.
func (rp *RecurringProcessor) Start(ctx context.Context) {
	go rp.run(ctx)
}

func (rp *RecurringProcessor) run(ctx context.Context) {
	// Check immediately on start, then every hour.
	rp.processAll()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("stopping recurring processor")
			return
		case <-ticker.C:
			rp.processAll()
		}
	}
}

func (rp *RecurringProcessor) processAll() {
	users, err := rp.store.LoadUsers()
	if err != nil {
		log.Printf("recurring: failed to load users: %v", err)
		return
	}

	today := time.Now().Format("2006-01-02")

	for username := range users {
		rp.processUser(username, today)
	}
}

func (rp *RecurringProcessor) processUser(username, today string) {
	ud, err := rp.store.LoadUserData(username)
	if err != nil {
		log.Printf("recurring: failed to load data for %s: %v", username, err)
		return
	}

	dirty := false

	for i, re := range ud.RecurringExpenses {
		if !re.Active || re.NextDate > today {
			continue
		}

		// Auto-create expense linked to this recurring parent
		expense := models.Expense{
			ID:                NewID(),
			Amount:            re.Amount,
			Currency:          re.Currency,
			AmountBase:        0, // will be set below
			Category:          re.Category,
			Description:       re.Description,
			Date:              re.NextDate,
			IsRecurring:       true,
			RecurringParentID: new(re.ID),
			CreatedAt:         time.Now(),
		}

		// Load user for currency conversion
		users, err := rp.store.LoadUsers()
		if err != nil {
			log.Printf("recurring: failed to load user %s for currency conversion: %v", username, err)
			continue
		}

		user := users[username]
		expense.AmountBase = ConvertToBase(expense.Amount, expense.Currency, user)

		ud.Expenses = append(ud.Expenses, expense)

		// Advance next_date
		ud.RecurringExpenses[i].NextDate = computeNextDate(re.NextDate, re.Frequency, re.DayOfMonth)

		dirty = true

		log.Printf("recurring: created expense for %s (%s, %s)", username, re.Description, re.NextDate)
	}

	if dirty {
		if err := rp.store.SaveUserData(username, ud); err != nil {
			log.Printf("recurring: failed to save data for %s: %v", username, err)
		}
	}
}

// computeNextDate returns the next occurrence date based on frequency.
func computeNextDate(currentDate, frequency string, dayOfMonth int) string {
	t, err := time.Parse("2006-01-02", currentDate)
	if err != nil {
		return currentDate
	}

	switch frequency {
	case "weekly":
		return t.Add(7 * 24 * time.Hour).Format("2006-01-02")
	case "monthly":
		next := t.AddDate(0, 1, 0)
		if dayOfMonth > 0 && dayOfMonth <= 31 {
			// Use the configured day of month, clamped to the actual max days in the next month
			maxDay := daysInMonth(next.Year(), next.Month())
			d := min(dayOfMonth, maxDay)

			return time.Date(next.Year(), next.Month(), d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		}

		return next.Format("2006-01-02")
	case "yearly":
		return t.AddDate(1, 0, 0).Format("2006-01-02")
	default:
		return t.AddDate(0, 1, 0).Format("2006-01-02")
	}
}

func daysInMonth(year int, month time.Month) int {
	// Day 32 of any month rolls over to the next month; subtract the overflow to get last day.
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

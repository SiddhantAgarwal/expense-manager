# Expense Manager — Specification

## Overview
A self-hosted, web-based expense manager with a Go backend, Go template frontend (HTMX), and flat-file JSON storage. Designed for multi-user local deployment.

---

## Architecture

### Stack
- **Backend**: Go (net/http) with gorilla/mux or chi router
- **Frontend**: Go `html/template` + HTMX for interactivity + minimal CSS framework (PicoCSS or similar)
- **Storage**: Flat JSON files — one file per user under `data/<username>.json`
- **Auth**: Session-based auth with cookies; bcrypt-hashed passwords stored in `data/users.json`

### Project Structure
```
expense-manager/
├── cmd/
│   └── server/
│       └── main.go            # Entry point
├── internal/
│   ├── auth/                  # Login, signup, session management
│   ├── models/                # Data structures (User, Expense, Budget, etc.)
│   ├── store/                 # JSON file read/write operations
│   ├── handlers/              # HTTP handlers (routes)
│   ├── services/              # Business logic (recurring, reports, currency)
│   └── middleware/            # Auth middleware, logging
├── templates/
│   ├── base.html              # Layout template
│   ├── login.html
│   ├── signup.html
│   ├── dashboard.html
│   ├── expenses.html
│   ├── budgets.html
│   ├── reports.html
│   └── settings.html
├── static/
│   ├── css/
│   └── js/                    # HTMX, minimal custom JS
├── data/                      # JSON storage directory (gitignored)
│   ├── users.json
│   └── <username>.json
├── go.mod
└── SPEC.md
```

---

## Data Models

### User (stored in `data/users.json`)
```json
{
  "username": "alice",
  "password_hash": "bcrypt_hash",
  "default_currency": "USD",
  "exchange_rates": {
    "EUR": 0.92,
    "GBP": 0.79,
    "INR": 83.1
  },
  "created_at": "2026-04-14T10:00:00Z"
}
```

### User Data (stored in `data/<username>.json`)
```json
{
  "expenses": [
    {
      "id": "uuid",
      "amount": 50.00,
      "currency": "EUR",
      "amount_base": 54.35,
      "category": "Food",
      "description": "Groceries",
      "date": "2026-04-14",
      "is_recurring": false,
      "recurring_parent_id": null,
      "created_at": "2026-04-14T10:00:00Z"
    }
  ],
  "budgets": [
    {
      "id": "uuid",
      "category": "Food",
      "monthly_limit": 500.00,
      "currency": "USD",
      "created_at": "2026-04-14T10:00:00Z"
    }
  ],
  "recurring_expenses": [
    {
      "id": "uuid",
      "amount": 1200.00,
      "currency": "USD",
      "category": "Housing",
      "description": "Rent",
      "frequency": "monthly",
      "next_date": "2026-05-01",
      "day_of_month": 1,
      "active": true,
      "created_at": "2026-04-14T10:00:00Z"
    }
  ],
  "categories": ["Food", "Transport", "Housing", "Entertainment", "Health", "Other"]
}
```

---

## Features

### 1. Authentication
- Signup with username + password (bcrypt hash)
- Login → session cookie (secure, HttpOnly)
- Logout
- Auth middleware protects all routes except `/login` and `/signup`
- Each user's data is isolated to their own JSON file

### 2. Expense CRUD
- **Create**: amount, currency, category, description, date
- **Read**: paginated list with filters (date range, category, currency)
- **Update**: edit any field
- **Delete**: remove expense (soft consideration — confirm dialog via HTMX)
- Amounts stored in original currency AND converted to user's default currency using their exchange rates

### 3. Budgets
- Set monthly budget per category
- Dashboard shows spending vs. budget (progress bar)
- Alert when spending reaches 80% and 100% of budget (visual indicator on dashboard)

### 4. Reports
- **Monthly summary**: total spent, broken down by category (bar chart or table)
- **Category breakdown**: pie chart or table of spending by category for a date range
- **Trend**: month-over-month comparison (simple table or sparkline)
- All report amounts in user's default currency

### 5. Recurring Expenses
- Create recurring expense with frequency (weekly, monthly, yearly) and day
- Background goroutine checks every hour and auto-creates expense entries when `next_date` arrives
- Auto-created expenses are linked to parent via `recurring_parent_id`
- Can pause/resume or delete recurring expenses
- Editing a recurring expense does NOT change already-created entries

### 6. Multi-Currency
- User selects a default currency at signup
- User can manually set exchange rates in Settings
- When logging an expense in a foreign currency, amount is auto-converted to default currency using stored rates
- Reports always show in default currency
- Exchange rates displayed and editable in Settings page

### 7. Settings
- Change default currency
- Update exchange rates
- Add/remove custom categories
- Change password

---

## Pages & Routes

| Route | Method | Description |
|---|---|---|
| `/` | GET | Redirect to `/dashboard` or `/login` |
| `/signup` | GET/POST | Signup page |
| `/login` | GET/POST | Login page |
| `/logout` | POST | Logout and redirect |
| `/dashboard` | GET | Overview: recent expenses, budget status, monthly total |
| `/expenses` | GET | Expense list with filters |
| `/expenses` | POST | Create new expense |
| `/expenses/{id}` | PUT | Update expense (HTMX) |
| `/expenses/{id}` | DELETE | Delete expense (HTMX) |
| `/expenses/new` | GET | New expense form |
| `/budgets` | GET | Budget list |
| `/budgets` | POST | Create/update budget |
| `/budgets/{id}` | DELETE | Delete budget |
| `/reports` | GET | Reports with date range picker |
| `/recurring` | GET | Recurring expenses list |
| `/recurring` | POST | Create recurring expense |
| `/recurring/{id}` | PUT | Update (pause/resume) |
| `/recurring/{id}` | DELETE | Delete recurring expense |
| `/settings` | GET | Settings page |
| `/settings` | POST | Update settings (currency, rates, categories, password) |

---

## Key Design Decisions

1. **HTMX over SPA**: Server renders HTML fragments; HTMX swaps them in. No build step, no JS framework.
2. **JSON files over SQLite**: No CGO dependency, human-readable files, trivial backup (copy the `data/` dir).
3. **File locking**: `sync.Mutex` per user file to prevent concurrent write corruption.
4. **Session storage**: In-memory map with TTL (server restart = re-login). Simple and sufficient for local use.
5. **No external API calls**: Exchange rates are manual. No internet dependency.

---

## Implementation Order

1. Project scaffolding + Go module setup
2. Data models + JSON store (read/write with file locking)
3. Auth (signup, login, logout, middleware, sessions)
4. Expense CRUD + list page
5. Dashboard
6. Budgets
7. Recurring expenses (background goroutine)
8. Reports
9. Multi-currency + Settings
10. Polish (CSS, error handling, validation)

## Implemented so far
1. Project scaffolding + Go module setup
2. Data models + JSON store (read/write with file locking)
3. Auth (signup, login, logout, middleware, sessions)
4. Expense CRUD + list page
5. Dashboard
6. Budgets
7. Recurring expenses (background goroutine)

## Phases remaining
1. Reports
2. Multi-currency + Settings
3. Polish (CSS, error handling, validation)
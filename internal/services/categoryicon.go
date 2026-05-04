package services

import "strings"

type IconOption struct {
	ID    string
	Label string
}

var AvailableIcons = []IconOption{
	{"icon-food", "Food"},
	{"icon-transport", "Transport"},
	{"icon-housing", "Housing"},
	{"icon-entertainment", "Entertainment"},
	{"icon-health", "Health"},
	{"icon-other", "Tag"},
	{"icon-education", "Education"},
	{"icon-shopping", "Shopping"},
	{"icon-gift", "Gift"},
	{"icon-pets", "Pets"},
	{"icon-fitness", "Fitness"},
	{"icon-dining", "Dining"},
	{"icon-books", "Books"},
	{"icon-travel", "Travel"},
	{"icon-music", "Music"},
	{"icon-savings", "Savings"},
}

var defaultCategoryIcons = map[string]string{
	"food":          "icon-food",
	"transport":     "icon-transport",
	"housing":       "icon-housing",
	"entertainment": "icon-entertainment",
	"health":        "icon-health",
	"other":         "icon-other",
}

func CategoryIconID(category string, customIcons map[string]string) string {
	if customIcons != nil {
		if id, ok := customIcons[category]; ok {
			return id
		}
	}

	key := strings.ToLower(category)
	if id, ok := defaultCategoryIcons[key]; ok {
		return id
	}

	return "icon-other"
}

func IsValidIconID(id string) bool {
	for _, opt := range AvailableIcons {
		if opt.ID == id {
			return true
		}
	}

	return false
}

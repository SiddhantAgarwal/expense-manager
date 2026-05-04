package services

import "strings"

func CategoryIconID(category string) string {
	mapping := map[string]string{
		"food":          "icon-food",
		"transport":     "icon-transport",
		"housing":       "icon-housing",
		"entertainment": "icon-entertainment",
		"health":        "icon-health",
		"other":         "icon-other",
	}

	key := strings.ToLower(category)
	if id, ok := mapping[key]; ok {
		return id
	}

	return "icon-other"
}

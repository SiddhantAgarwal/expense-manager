package services

import (
	"strconv"
	"strings"
)

func FormatAmount(amount float64, format string) string {
	if format == "" {
		format = "us"
	}

	s := strconv.FormatFloat(amount, 'f', 2, 64)

	parts := strings.SplitN(s, ".", 2)
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) == 2 {
		decimalPart = parts[1]
	}

	neg := false
	if strings.HasPrefix(integerPart, "-") {
		neg = true
		integerPart = integerPart[1:]
	}

	var formatted string
	switch format {
	case "indian":
		formatted = formatIndian(integerPart)
	case "european":
		formatted = formatEuropean(integerPart)
	default:
		formatted = formatUS(integerPart)
	}

	if neg {
		formatted = "-" + formatted
	}

	if decimalPart != "" {
		switch format {
		case "european":
			formatted += "," + decimalPart
		default:
			formatted += "." + decimalPart
		}
	}

	return formatted
}

func FormatAmountInput(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

func formatUS(integer string) string {
	return groupDigits(integer, 3, ",")
}

func formatIndian(integer string) string {
	if len(integer) <= 3 {
		return integer
	}
	result := integer[len(integer)-3:]
	rest := integer[:len(integer)-3]
	for len(rest) > 2 {
		result = rest[len(rest)-2:] + "," + result
		rest = rest[:len(rest)-2]
	}
	if len(rest) > 0 {
		result = rest + "," + result
	}
	return result
}

func formatEuropean(integer string) string {
	return groupDigits(integer, 3, ".")
}

func groupDigits(integer string, groupSize int, sep string) string {
	if len(integer) <= groupSize {
		return integer
	}
	var result string
	for i := len(integer); i > 0; i -= groupSize {
		start := i - groupSize
		if start < 0 {
			start = 0
		}
		part := integer[start:i]
		if result == "" {
			result = part
		} else {
			result = part + sep + result
		}
	}
	return result
}

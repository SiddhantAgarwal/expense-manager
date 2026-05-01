package services

import "testing"

func TestFormatAmountUS(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{0, "0.00"},
		{1.5, "1.50"},
		{999.99, "999.99"},
		{1000, "1,000.00"},
		{12345.67, "12,345.67"},
		{123456.78, "123,456.78"},
		{1234567.89, "1,234,567.89"},
		{12345678.9, "12,345,678.90"},
		{-1234.56, "-1,234.56"},
	}

	for _, tt := range tests {
		got := FormatAmount(tt.amount, "us")
		if got != tt.want {
			t.Errorf("FormatAmount(%v, \"us\") = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestFormatAmountIndian(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{0, "0.00"},
		{1.5, "1.50"},
		{999.99, "999.99"},
		{1000, "1,000.00"},
		{12345.67, "12,345.67"},
		{123456.78, "1,23,456.78"},
		{1234567.89, "12,34,567.89"},
		{12345678.9, "1,23,45,678.90"},
		{-1234.56, "-1,234.56"},
	}

	for _, tt := range tests {
		got := FormatAmount(tt.amount, "indian")
		if got != tt.want {
			t.Errorf("FormatAmount(%v, \"indian\") = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestFormatAmountEuropean(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{0, "0,00"},
		{1.5, "1,50"},
		{999.99, "999,99"},
		{1000, "1.000,00"},
		{12345.67, "12.345,67"},
		{123456.78, "123.456,78"},
		{1234567.89, "1.234.567,89"},
		{12345678.9, "12.345.678,90"},
		{-1234.56, "-1.234,56"},
	}

	for _, tt := range tests {
		got := FormatAmount(tt.amount, "european")
		if got != tt.want {
			t.Errorf("FormatAmount(%v, \"european\") = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestFormatAmountEmptyDefault(t *testing.T) {
	got := FormatAmount(1234.56, "")
	if got != "1,234.56" {
		t.Errorf("FormatAmount with empty format should default to US, got %q", got)
	}
}

func TestFormatAmountInput(t *testing.T) {
	got := FormatAmountInput(1234.56)
	if got != "1234.56" {
		t.Errorf("FormatAmountInput(1234.56) = %q, want \"1234.56\"", got)
	}
}

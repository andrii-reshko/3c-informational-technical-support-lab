package web

import (
	"testing"
)

func TestRiskClass(t *testing.T) {
	// Таблиця сценаріїв: маппінг статусу ризику на CSS клас
	tests := []struct {
		name     string // Назва тесту
		risk     string // Вхідний статус
		expected string // Очікуваний CSS клас
	}{
		{
			name:     "Критичний статус",
			risk:     "CRITICAL",
			expected: "text-bg-danger",
		},
		{
			name:     "Статус Warning",
			risk:     "WARNING",
			expected: "text-bg-success", // За поточною логікою усе окрім CRITICAL є success
		},
		{
			name:     "Безпечний статус",
			risk:     "SAFE",
			expected: "text-bg-success",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := riskClass(tc.risk)
			if result != tc.expected {
				t.Errorf("Для %q очікували %q, отримали %q", tc.risk, tc.expected, result)
			}
		})
	}
}

func TestNormalizeCoveragePercent(t *testing.T) {
	// Таблиця сценаріїв: нормалізація відсотків
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "Більше 100% обрізається до 100",
			input:    105.5,
			expected: 100.0,
		},
		{
			name:     "Менше 0% обрізається до 0",
			input:    -5.0,
			expected: 0.0,
		},
		{
			name:     "Нормальне значення залишається",
			input:    85.5,
			expected: 85.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeCoveragePercent(tc.input)
			if result != tc.expected {
				t.Errorf("Вхід %v: очікували %v, отримали %v", tc.input, tc.expected, result)
			}
		})
	}
}

func TestSafeBoundaryOrDefault(t *testing.T) {
	// Таблиця сценаріїв: перевірка дефолтних меж
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "Нульове значення змінюється на дефолт (75)",
			input:    0.0,
			expected: 75.0,
		},
		{
			name:     "Коректне значення залишається незмінним",
			input:    80.0,
			expected: 80.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := safeBoundaryOrDefault(tc.input)
			if result != tc.expected {
				t.Errorf("Вхід %v: очікували %v, отримали %v", tc.input, tc.expected, result)
			}
		})
	}
}

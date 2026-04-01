package bootstrap

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Встановлюємо тестову змінну оточення
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	// Таблиця тестових сценаріїв
	tests := []struct {
		name     string // Назва тесту
		key      string // Ключ змінної оточення
		fallback string // Значення за замовчуванням
		expected string // Очікуваний результат
	}{
		{
			name:     "Якщо змінна існує, повертається її значення",
			key:      "TEST_KEY",
			fallback: "default",
			expected: "test_value",
		},
		{
			name:     "Якщо змінна відсутня, повертається fallback",
			key:      "MISSING_KEY",
			fallback: "default",
			expected: "default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getEnv(tc.key, tc.fallback)
			if result != tc.expected {
				t.Errorf("Очікували %q, але отримали %q", tc.expected, result)
			}
		})
	}
}

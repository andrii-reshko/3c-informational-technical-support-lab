package app

import (
	"testing"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
)

// Тестуємо FeatureEngineer - це чиста логіка підготовки даних.
// Її дуже легко тестувати, бо вона не залежить від баз даних чи мережі.
func TestPrepareVector(t *testing.T) {
	// Конфіг моделі, який каже, що у нас 16 ознак (features),
	// і дає масиви Scale та Min для нормалізації, щоб уникнути index out of range.
	cfg := domain.ModelConfig{
		Features: make([]string, 16),
		Scale:    make([]float64, 16),
		Min:      make([]float64, 16),
	}
	fe := NewFeatureEngineer(cfg)

	tests := []struct {
		name        string // Назва сценарію
		metricsSize int    // Скільки метрик передаємо на вхід
		expectedNil bool   // Чи очікуємо, що результат буде nil (через нестачу даних)
	}{
		{
			name:        "Недостатньо даних (менше 61)",
			metricsSize: 10,
			expectedNil: true, // Очікуємо nil, бо для вікна 60 хв треба 61 точку
		},
		{
			name:        "Точно достатньо даних (61)",
			metricsSize: 61,
			expectedNil: false, // Очікуємо нормальний масив float32
		},
		{
			name:        "Даних більше ніж треба (100)",
			metricsSize: 100,
			expectedNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Створюємо масив пустих метрик потрібного розміру з валідними часовими мітками
			metrics := make([]domain.Metric, tc.metricsSize)
			for i := range metrics {
				metrics[i] = domain.Metric{
					Timestamp:  time.Now(),
					CPUPercent: float64(i),
				}
			}

			// Викликаємо метод, який тестуємо
			result := fe.PrepareVector(metrics)

			// Перевіряємо результат
			if tc.expectedNil && result != nil {
				t.Errorf("Очікували nil, але отримали вектор")
			}
			if !tc.expectedNil && result == nil {
				t.Errorf("Очікували вектор, але отримали nil")
			}
			// Додаткова перевірка: якщо масив повернувся, його розмір має дорівнювати 16
			if !tc.expectedNil && len(result) != 16 {
				t.Errorf("Очікували вектор розміром 16, отримали %d", len(result))
			}
		})
	}
}

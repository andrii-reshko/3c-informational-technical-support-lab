package app

import (
	"context"
	"testing"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
)

// Цей блок - це створення "фейкового" репозиторію (Mock) для тестів.
// Так як Collector працює з базою даних, ми не можемо підключати справжню БД у простих тестах.
// Ми створюємо "заглушку", яка просто запам'ятовує факт: чи був викликаний метод DeleteOldMetrics.
type mockMetricsRepo struct {
	// Це трюк в Go, щоб не реалізовувати всі методи інтерфейсу (ми їх "наслідуємо" з пустишкою)
	domain.MetricsRepository
	wasDeleted bool // Прапорець, чи виконалася операція
}

func (m *mockMetricsRepo) DeleteOldMetrics(ctx context.Context, olderThan time.Time) error {
	m.wasDeleted = true // Запам'ятовуємо, що видалення відбулось
	return nil
}

// Тепер тестуємо сам CollectorService, ізольовано від реальної БД.
func TestCollectorService_Cleanup(t *testing.T) {
	// Створюємо нашу заглушку замість реального SQLite-репозиторію
	fakeRepo := &mockMetricsRepo{}

	// Ініціалізуємо сервіс, передаючи йому фейковий репозиторій
	service := NewCollectorService(nil, fakeRepo, nil)

	tests := []struct {
		name         string
		daysToKeep   int  // Скільки днів зберігати
		expectDelete bool // Чи мав сервіс викликати видалення старих даних
	}{
		{
			name:         "Не видаляти, якщо днів <= 0 (захист від видалення всього)",
			daysToKeep:   0,
			expectDelete: false, // Очікуємо, що сервіс проігнорує цей виклик
		},
		{
			name:         "Не видаляти, якщо передано від'ємне число",
			daysToKeep:   -5,
			expectDelete: false,
		},
		{
			name:         "Успішне видалення для коректної кількості днів",
			daysToKeep:   7,
			expectDelete: true, // Очікуємо, що сервіс викличе DeleteOldMetrics
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Очищаємо прапорець перед кожним тестом
			fakeRepo.wasDeleted = false

			// Викликаємо функцію, яку тестуємо
			service.Cleanup(context.Background(), tc.daysToKeep)

			// Перевіряємо, чи збігається поведінка з очікуваною
			if fakeRepo.wasDeleted != tc.expectDelete {
				t.Errorf("Очікували виклик видалення=%v, але отримали=%v", tc.expectDelete, fakeRepo.wasDeleted)
			}
		})
	}
}

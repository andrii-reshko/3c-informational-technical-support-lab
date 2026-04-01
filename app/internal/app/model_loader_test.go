package app

import (
	"os"
	"testing"
)

// Тестуємо завантаження JSON метаданих. Тут ми створюємо справжній тимчасовий
// файл на диску і дивимось, як логіка реагує на нього або на його відсутність.
func TestLoadJsonMeta(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func() string // Функція для підготовки файла перед тестом
		cleanupFunc  func(path string) // Очищення після тесту
		expectError  bool // Чи має бути помилка при зчитуванні
	}{
		{
			name: "Файлу не існує",
			setupFunc: func() string {
				// Просто повертаємо вигадане ім'я
				return "non_existent_file_123.json"
			},
			cleanupFunc: func(path string) {}, // Немає чого видаляти
			expectError: true, // Очікуємо помилку: "file not found"
		},
		{
			name: "Невалідний JSON",
			setupFunc: func() string {
				// Створюємо тимчасовий файл зі зламаним JSON
				file, _ := os.CreateTemp("", "bad_meta_*.json")
				file.WriteString("{ bad json format ]")
				file.Close()
				return file.Name()
			},
			cleanupFunc: func(path string) {
				os.Remove(path) // Видаляємо після тесту
			},
			expectError: true, // Очікуємо помилку: "unmarshal error"
		},
		{
			name: "Успішне зчитування валідного JSON",
			setupFunc: func() string {
				// Створюємо валідний JSON
				file, _ := os.CreateTemp("", "good_meta_*.json")
				file.WriteString(`{"features":["f1","f2"]}`)
				file.Close()
				return file.Name()
			},
			cleanupFunc: func(path string) {
				os.Remove(path) // Видаляємо після тесту
			},
			expectError: false, // Все має пройти успішно
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Готуємо тестове середовище
			path := tc.setupFunc()
			
			// Змінна куди зчитається результат
			var target map[string]interface{}
			
			// Сам тест логіки
			err := loadJsonMeta(path, &target)
			
			// Перевіряємо помилку
			if tc.expectError && err == nil {
				t.Error("Очікували помилку, але отримали успіх")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Не очікували помилки, але отримали: %v", err)
			}
			
			// Очищаємо за собою
			tc.cleanupFunc(path)
		})
	}
}

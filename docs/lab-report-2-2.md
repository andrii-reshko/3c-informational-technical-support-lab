# Інформаційно-технічний супровід програмного забезпечення

Частина 2. Тестування програмного забезпечення (за індивідуальним варіантом)  
Звіт про виконання лабораторної роботи: Розширення системи прогнозування - додавання RAM forecasting

---

## 1. Мета роботи
Метою цієї роботи є розширення функціональності системи прогнозування навантаження серверів шляхом додавання окремої моделі для прогнозування використання RAM (оперативної пам'яті), оновлення UI для відображення прогнозів як для CPU, так і для RAM, та оновлення технічної документації.

## 2. Завдання
1. Додати окрему ML-модель для прогнозування RAM usage.
2. Оновити Go-сервіс для використання двох моделей (CPU + RAM).
3. Оновити UI для відображення прогнозів CPU та RAM окремо.
4. Оновити технічну документацію та User Manual.
5. Оформити звіт про виконану роботу.

## 3. Хід роботи

### 3.1. Розширення ML-пайплайну

#### Створення RAM-моделі
У рамках розширення ML-пайплайну (`./model/pipeline.py`) було додано тренування окремої моделі для прогнозування RAM:

```python
target_resources = ['cpu_usage', 'ram_usage']
```

Це забезпечує:
- Навчання окремих MLP-моделей для кожного типу ресурсу
- Експорт окремих `.tflite` файлів для CPU та RAM
- Генерацію окремих `*_meta.json` та `*_scaler.json` конфігураційних файлів

**Результати тренування:**
| Ресурс | MAE | Coverage | Pinball Loss |
|--------|-----|----------|--------------|
| CPU | 2.71 | 84.8% | 0.685 |
| RAM | 0.73 | 79.1% | 0.153 |

### 3.2. Оновлення Go-сервісу

#### Зміни у структурах даних
Додано окремі поля для CPU та RAM у відповіді прогнозу (`internal/domain/forecast.go`):

```go
type ForecastResponse struct {
    // CPU Forecast
    CPUCurrent    float64
    CPULowerBound float64
    CPUUpperBound float64
    CPUPredicted  float64
    CPUQuality    ModelQuality
    
    // RAM Forecast
    RAMCurrent    float64
    RAMLowerBound float64
    RAMUpperBound float64
    RAMPredicted  float64
    RAMQuality    ModelQuality
    
    // ...
}
```

#### Метрика RAMPercent
Додано поле `RAMPercent` до структури `Metric` для зберігання % використання RAM:

```go
type Metric struct {
    NodeID     string
    Timestamp  time.Time
    CPUPercent float64
    RAMBytes   int64
    RAMPercent float64  // Нове поле
}
```

#### Feature Engineer
Оновлено `FeatureEngineer.PrepareVector()` для підтримки обох типів ресурсів:

```go
func (fe *FeatureEngineer) PrepareVector(metrics []domain.Metric, isRAM bool) []float32
```

Параметр `isRAM` визначає, для якого ресурсу будуть обчислюватись фічі.

#### Prometheus Adapter
Додано кешування лімітів пам'яті та обчислення `RAMPercent`:

```go
type MemLimiter struct {
    mu     sync.RWMutex
    limits map[string]int64
}

// RAMPercent = (RAMBytes / mem_limit) * 100
```

#### Container & Forecaster
Оновлено `BuildForecaster()` для завантаження двох моделей:

```go
cpuModel, err := tflite.NewPredictor(fmt.Sprintf("%s/models/global_trace_cpu_usage_15m_model.tflite", cwd))
ramModel, err := tflite.NewPredictor(fmt.Sprintf("%s/models/global_trace_ram_usage_15m_model.tflite", cwd))
```

### 3.3. Оновлення UI

#### Dashboard (головна сторінка)
Оновлено відображення карток вузлів:
- **Ліворуч**: CPU поточний % → CPU прогноз (lower-upper)
- **Праворуч**: RAM поточний % → RAM прогноз (lower-upper)
- Додано підтримку resize для sparkline-графіків

#### Node Details
Додано окремі графіки для CPU та RAM:
- Два інтерактивних D3.js графіки з live-оновленням через WebSocket
- Окремі Quality Metrics для кожного ресурсу

#### WebSocket
Оновлено формат повідомлень для передачі даних CPU та RAM:

```json
{
  "cpu": 15.5,
  "ram": 42.3,
  "cpu_lower": 10.2,
  "cpu_upper": 25.8,
  "ram_lower": 35.0,
  "ram_upper": 55.0,
  "now_iso": "2026-05-13T08:55:00Z"
}
```

### 3.4. Оновлення документації

#### Технічна документація
Оновлено схеми та описи:
- `tech/01_db.md` - додано колонку `ram_usage_percent`
- `tech/02_classes.md` - оновлено схему класів з двома моделями
- `tech/03_algorithms.md` - оновлено блок-схему для CPU + RAM

#### User Manual
Оновлено розділи:
- `user/02_terms.md` - додано терміни RAM forecasting
- `user/06_ui.md` - оновлено опис UI з новими скріншотами
- `user/07_ui_elements.md` - оновлено огляд елементів UI
- `user/08_api.md` - оновлено API документацію

## 4. Результати тестування

### Функціональне тестування нової функціональності
| № | Тестові дані | Фактичний результат | Очікувані результати | Статус |
|---|---|---|---|---|
| 1 | Dashboard відображає CPU + RAM дані | Відображаються обидва типи | CPU та RAM з прогнозами | ✅ Pass |
| 2 | Node Details показує два графіки | CPU та RAM графіки | Окремі графіки для CPU та RAM | ✅ Pass |
| 3 | WebSocket оновлює CPU та RAM | Обидва типи оновлюються | Live streaming CPU + RAM | ✅ Pass |
| 4 | RAM forecasts мають коректні bounds | lower < current < upper | Інтервальний прогноз | ✅ Pass |
| 5 | Quality metrics для CPU та RAM | Обидва набори метрик | Окремі MAE, Coverage | ✅ Pass |
| 6 | Ресайз вікна - графіки оновлюються | Sparklines перемальовуються | Responsive resize | ✅ Pass |

## 5. Висновки

У результаті виконання лабораторної роботи було успішно розширено систему прогнозування навантаження серверів:

1. **ML-пайплайн**: Додано тренування окремої моделі для RAM forecasting з покращеними показниками якості (MAE: 0.73, Coverage: 79.1%).

2. **Backend (Go)**: 
   - Рефакторено `ForecasterService` для підтримки двох моделей
   - Додано обчислення `RAMPercent` з використанням лімітів пам'яті
   - Оновлено схему БД з новою колонкою

3. **Frontend**:
   - Dashboard показує CPU → CPU pred | RAM → RAM pred
   - Node Details має два окремих графіки для CPU та RAM
   - Додано responsive resize для sparklines

4. **Документація**: Оновлено технічну документацію та User Manual з новими скріншотами.

Система готова до експлуатації з підтримкою прогнозування як CPU, так і RAM ресурсів.
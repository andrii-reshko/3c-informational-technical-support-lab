# Блок-схема алгоритмів

Ця схема показує загальний процес отримання прогнозу через систему (ISO 5807:2016).

```mermaid
flowchart TD
    %% Start
    Start([Початок]) --> A[/"Отримати запит на прогноз \n GET /predict?node_id=id"/]

    %% Process
    A --> B{"Чи існує \n node_id в БД?"}
    B -- Ні --> Error1[/"Повернення помилки 404"/]
    B -- Так --> C["Отримати історичні метрики \n з SQLite БД за останні 60 хв"]

    C --> D{"Чи достатньо метрик \n для вікна?"}
    D -- Ні --> Error2[/"Повернення помилки \n 'Not enough data'"/]
    D -- Так --> E["Feature Engineer: \n нормалізація, додавання лагів, \n rolling mean для CPU та RAM"]

    E --> F1["Inference CPU: \n TFLite(features, isRAM=false)"]
    E --> F2["Inference RAM: \n TFLite(features, isRAM=true)"]

    F1 --> G1["CPU квантильний прогноз: \n cpu_lower, cpu_upper"]
    F2 --> G2["RAM квантильний прогноз: \n ram_lower, ram_upper"]
    
    G1 --> H["Визначення risk_status \n SAFE, WARNING, CRITICAL \n відповідно до max(cpu_upper, ram_upper) \n та hardware-limits"]

    H --> I[/"Формування JSON відповіді \n CPU + RAM forecast"/]
    Error1 --> J([Кінець])
    Error2 --> J
    I --> J
```

### Опис кроків:
1. **Отримання метрик** - запит останніх 61 метрики для формування вікон та лагів
2. **Feature Engineering** - для кожного ресурсу (CPU/RAM) обчислюються:
   - Лаги (lag1, lag2, lag3, lag5, lag10, lag15, lag30)
   - Rolling statistics (mean/max для вікон 5, 15, 60 хв)
   - Часові ознаки (година, день тижня)
   - Нормалізація через MinMaxScaler
3. **Inference** - паралельне виконання для CPU та RAM моделей
4. **Визначення ризику** - максимальний upper bound з обох ресурсів порівнюється з SafeBoundary

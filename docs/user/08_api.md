# 8. АРІ документація

Система надає REST API. Основний формат даних - JSON. Інтерфейси також використовують WebSockets (WSS) для потокової передачі даних.

### Отримання передбачення:
`GET /predict?node_id={node_id}`
*   **Опис:** Робить запит на отримання прогнозу для конкретного сервера (за його ID)
*   **Параметри запиту:**
    *   `node_id` (string) - ідентифікатор вузла.
*   **Приклад відповіді (JSON):**
    ```json
    {
       "node_id": "test-node-01",
       "horizon_minutes": 15,
       "cpu_current_percent": 15.5,
       "cpu_lower_bound": 10.2,
       "cpu_upper_bound": 25.8,
       "cpu_predicted_peak": 25.8,
       "cpu_quality": {
          "mae": 2.71,
          "coverage": 0.848
       },
       "ram_current_percent": 42.3,
       "ram_lower_bound": 35.0,
       "ram_upper_bound": 55.0,
       "ram_predicted_peak": 55.0,
       "ram_quality": {
          "mae": 0.73,
          "coverage": 0.791
       },
       "risk_level": "safe",
       "headroom_percent": 44.7,
       "coverage_low": false,
       "upper_bound_exceeded": false
    }
    ```

### WebSocket підключення:
`WS /ws`
*   **Опис:** Використовується UI-застосунком для отримання метрик та прогнозів для дашборду у режимі реального часу. Дані надсилаються автоматично в форматі JSON.
*   **Параметри:**
    *   `node_id` (string, опціонально) - для стріму конкретного вузла
    *   `interval_sec` (int, опціонально) - інтервал оновлення (за замовчуванням 2 сек)
*   **Формат повідомлення:**
    ```json
    {
       "timestamp": "2026-05-13T08:55:00Z",
       "cpu": 15.5,
       "ram": 42.3,
       "cpu_lower": 10.2,
       "cpu_upper": 25.8,
       "ram_lower": 35.0,
       "ram_upper": 55.0,
       "now_iso": "2026-05-13T08:55:00Z"
    }
    ```

package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
	"github.com/jmoiron/sqlx"
)

type metricsRepo struct {
	db    *sqlx.DB
	table string
}

func NewMetricsRepository(db *sqlx.DB) domain.MetricsRepository {
	return &metricsRepo{
		db:    db,
		table: "metrics",
	}
}

func (r *metricsRepo) SaveMetrics(ctx context.Context, metrics []domain.Metric) error {
	query := `INSERT OR IGNORE INTO %s (node_id, timestamp, cpu_usage_percent, ram_usage_bytes)
            VALUES (:node_id, :timestamp, :cpu_usage_percent, :ram_usage_bytes)`
	query = fmt.Sprintf(query, r.table)
	_, err := r.db.NamedExecContext(ctx, query, metrics)
	return err
}

func (r *metricsRepo) GetLastMetricTimestamp(ctx context.Context, nodeID string) (time.Time, error) {
	var tsStr sql.NullString
	query := `SELECT MAX(timestamp) FROM %s WHERE node_id = $1`
	query = fmt.Sprintf(query, r.table)

	err := r.db.GetContext(ctx, &tsStr, query, nodeID)
	if err != nil {
		return time.Time{}, err
	}

	// 2. Якщо в базі NULL (немає метрик), повертаємо нульовий час
	if !tsStr.Valid || tsStr.String == "" {
		return time.Time{}, nil
	}

	// 3. Парсимо рядок у time.Time.
	// SQLite зазвичай використовує формат "2006-01-02 15:04:05" або ISO8601
	// Спробуємо спочатку стандартний для твого Collector формат
	layouts := []string{
		"2006-01-02 15:04:05.999999-07:00",    // Твій точний формат (6 знаків після крапки)
		"2006-01-02 15:04:05.999999999-07:00", // На випадок наносекунд (9 знаків)
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	var t time.Time
	var parseErr error
	for _, layout := range layouts {
		t, parseErr = time.Parse(layout, tsStr.String)
		if parseErr == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse timestamp '%s': %v", tsStr.String, parseErr)
}

func (r *metricsRepo) GetRangeMetrics(ctx context.Context, nodeID string, start, end time.Time) ([]domain.Metric, error) {
	var metrics []domain.Metric
	query := `
       SELECT node_id, timestamp, cpu_usage_percent, ram_usage_bytes FROM %s 
       WHERE node_id = $1 AND timestamp >= $2 AND timestamp <= $3 
       ORDER BY timestamp ASC`
	query = fmt.Sprintf(query, r.table)

	err := r.db.SelectContext(ctx, &metrics, query, nodeID, start, end)
	return metrics, err
}

func (r *metricsRepo) DeleteOldMetrics(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM %s WHERE timestamp < $1`
	query = fmt.Sprintf(query, r.table)

	_, err := r.db.ExecContext(ctx, query, olderThan)
	return err
}

// / ===
// GetLatestMetricsForNode повертає останні N метрик для конкретного вузла
// у хронологічному порядку (від найстарішої до найновішої).
func (r *metricsRepo) GetLatestMetricsForNode(ctx context.Context, nodeID string, limit int) ([]domain.Metric, error) {
	var metrics []domain.Metric

	// Використовуємо підзапит, щоб отримати останні записи, а потім відсортувати їх ASC.
	// Це набагато швидше, ніж вигрібати всю базу і сортувати в Go.
	query := `
		SELECT node_id, timestamp, cpu_usage_percent, ram_usage_bytes
		FROM (
			SELECT node_id, timestamp, cpu_usage_percent, ram_usage_bytes
			FROM %s
			WHERE node_id = $1
			ORDER BY timestamp DESC
			LIMIT $2
		) sub
		ORDER BY timestamp ASC`
	query = fmt.Sprintf(query, r.table)

	err := r.db.SelectContext(ctx, &metrics, query, nodeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest metrics for node %s: %w", nodeID, err)
	}

	return metrics, nil
}

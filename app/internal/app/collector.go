package app

import (
	"context"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
	"github.com/sirupsen/logrus"
)

type CollectorService struct {
	nodes           domain.NodeRepository
	metrics         domain.MetricsRepository
	provider        domain.MetricProvider
	wasDisconnected bool
}

func NewCollectorService(
	nodes domain.NodeRepository,
	metrics domain.MetricsRepository,
	provider domain.MetricProvider,
) *CollectorService {
	return &CollectorService{
		nodes:           nodes,
		metrics:         metrics,
		provider:        provider,
		wasDisconnected: false,
	}
}

// SyncNodes використовує NodeRepository
func (s *CollectorService) SyncNodes(ctx context.Context) {
	nodes, err := s.provider.GetHardwareFrames(ctx)
	for _, node := range nodes {
		logrus.Debugf("collector: sync node %s, hardware: %.1f CPU %.1f RAM", node.ID, node.CpuCores, node.RamGB)
	}

	if err != nil {
		logrus.Errorf("failed to get hardware frames: %v", err)
		return
	}
	for _, n := range nodes {
		if err := s.nodes.UpsertNode(ctx, n); err != nil {
			logrus.Errorf("failed to upsert node %s: %v", n.ID, err)
		}
	}
}

// Collect використовує MetricsRepository
func (s *CollectorService) Collect(ctx context.Context) {
	metrics, err := s.provider.GetLatestMetrics(ctx)
	if err != nil {
		logrus.Errorf("Metrics collection failed: %v", err)
		s.wasDisconnected = true
		return
	}

	if s.wasDisconnected {
		logrus.Info("collector: connection recovered, triggering backfill")
		s.Backfill(ctx)
		s.wasDisconnected = false
	}

	for _, m := range metrics {
		logrus.Debugf("collector: sync node %s, hardware: %.1f%% CPU %d bytes RAM", m.NodeID, m.CPUPercent, m.RAMBytes)
	}
	if len(metrics) > 0 {
		if err := s.metrics.SaveMetrics(ctx, metrics); err != nil {
			logrus.Errorf("Failed to save metrics: %v", err)
		}
	}
}

func (s *CollectorService) Backfill(ctx context.Context) {
	logrus.Info("service: checking for data gaps")

	nodes, err := s.nodes.GetAllNodes(ctx)
	if err != nil {
		logrus.Errorf("service: failed to get nodes: %v")
		return
	}

	now := time.Now()
	totalRestored := 0

	for _, node := range nodes {
		lastTS, err := s.metrics.GetLastMetricTimestamp(ctx, node.ID)
		if err != nil {
			logrus.Errorf("service: failed to get cursor for node %s: %v", node.ID, err)
			continue
		}

		// Якщо даних взагалі немає, можливо, це новий вузол.
		// Можемо підтягнути дані, наприклад, за останню годину для "розігріву".
		if lastTS.IsZero() {
			lastTS = now.Add(-1 * time.Hour)
			logrus.Infof("service: no history for node %s, fetching last 1h", node.ID)
		}

		// 3. Перевіряємо, чи є розрив (наприклад, більше 15 секунд)
		gap := now.Sub(lastTS)
		if gap < 15*time.Second {
			logrus.Infof("service: node %s is up to date", node.ID)
			continue
		}

		logrus.Infof("service: found %v s gap for node %s, fetching from source", gap.Truncate(time.Second), node.ID)

		// 4. Питаємо Prometheus Range API про цей проміжок
		// Додаємо 1 секунду до lastTS, щоб не дублювати останню точку
		historicalMetrics, err := s.provider.GetRangeMetrics(ctx, node.ID, lastTS.Add(time.Second), now)
		if err != nil {
			logrus.Errorf("service: failed to fetch range from source for node %s: %v", node.ID, err)
			continue
		}

		// 5. Зберігаємо все, що знайшли, в базу
		if len(historicalMetrics) > 0 {
			err = s.metrics.SaveMetrics(ctx, historicalMetrics)
			if err != nil {
				logrus.Errorf("service: failed to save historical data for node %s: %v", node.ID, err)
				continue
			}
			totalRestored += len(historicalMetrics)
			logrus.Infof("service: restored %d points for %s", len(historicalMetrics), node.ID)
		} else {
			logrus.Warnf("service: source returned no data for %s in the requested range", node.ID)
		}
	}

	logrus.Infof("service: backfill completed, %d points restored", totalRestored)
}

func (s *CollectorService) Cleanup(ctx context.Context, days int) {
	if days <= 0 {
		logrus.Warn("service: cleanup, retention days set to 0 or less, skipping cleanup")
		return
	}

	olderThan := time.Now().AddDate(0, 0, -days)
	logrus.Infof("service: cleanup, starting removal of metrics older than %d days (%s)", days, olderThan.Format("2006-01-02"))

	err := s.metrics.DeleteOldMetrics(ctx, olderThan)
	if err != nil {
		logrus.Errorf("service: cleanup, failed to delete old metrics: %v", err)
		return
	}

	logrus.Info("service: cleanup, old metrics removed")
}

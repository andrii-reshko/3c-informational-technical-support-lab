package domain

import (
	"context"
	"time"
)

type NodeRepository interface {
	UpsertNode(ctx context.Context, node Node) error
	GetAllNodes(ctx context.Context) ([]Node, error)
	GetNodeByID(ctx context.Context, id string) (*Node, error)
	UpdateNode(ctx context.Context, node Node) error
}

type MetricsRepository interface {
	SaveMetrics(ctx context.Context, metrics []Metric) error
	GetLastMetricTimestamp(ctx context.Context, nodeID string) (time.Time, error)
	GetRangeMetrics(ctx context.Context, nodeID string, start, end time.Time) ([]Metric, error)
	DeleteOldMetrics(ctx context.Context, olderThan time.Time) error

	// Отримати останні N метрик для конкретного вузла
	GetLatestMetricsForNode(ctx context.Context, nodeID string, limit int) ([]Metric, error)
}

type MetricProvider interface {
	GetHardwareFrames(ctx context.Context) ([]Node, error)
	GetLatestMetrics(ctx context.Context) ([]Metric, error)
	GetRangeMetrics(ctx context.Context, nodeID string, start, end time.Time) ([]Metric, error)
}

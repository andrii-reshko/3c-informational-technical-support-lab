package domain

import (
	"time"
)

type Node struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CpuCores  float64   `db:"cpu_cores"`
	RamGB     float64   `db:"ram_gb"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Metric struct {
	NodeID     string    `db:"node_id"`
	Timestamp  time.Time `db:"timestamp"`
	CPUPercent float64   `db:"cpu_usage_percent"`
	RAMBytes   int64     `db:"ram_usage_bytes"`
}

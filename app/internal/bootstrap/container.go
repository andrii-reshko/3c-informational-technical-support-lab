package bootstrap

import (
	"context"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/db"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/prom"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/app"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Container struct {
	DB          *sqlx.DB
	Config      *Config
	NodeRepo    domain.NodeRepository
	MetricsRepo domain.MetricsRepository
	Prom        domain.MetricProvider
	Collector   *app.CollectorService
}

func NewContainer() *Container {
	cfg := Configure()

	// logging
	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("invalid log level: %v", err)
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	// persistence
	conn, err := db.NewDbConnection(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	// gateways
	nodeRepo := db.NewNodeRepository(conn)
	metricsRepo := db.NewMetricsRepository(conn)
	metricsProvider, err := prom.NewAdapter(cfg.PrometheusURL)
	if err != nil {
		log.Fatalf("failed to create Prometheus adapter: %v", err)
	}

	// service layer
	collector := app.NewCollectorService(nodeRepo, metricsRepo, metricsProvider)

	return &Container{
		Config:      cfg,
		DB:          conn,
		Collector:   collector,
		NodeRepo:    nodeRepo,
		MetricsRepo: metricsRepo,
		Prom:        metricsProvider,
	}
}

func (c *Container) Close(_ context.Context) {
	_ = c.DB.Close()
}

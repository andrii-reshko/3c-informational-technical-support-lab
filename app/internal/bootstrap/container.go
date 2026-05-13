package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/db"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/prom"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/tflite"
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

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("invalid log level: %v", err)
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	conn, err := db.NewDbConnection(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	nodeRepo := db.NewNodeRepository(conn)
	metricsRepo := db.NewMetricsRepository(conn)
	metricsProvider, err := prom.NewAdapter(cfg.PrometheusURL)
	if err != nil {
		log.Fatalf("failed to create Prometheus adapter: %v", err)
	}

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

func (c *Container) BuildForecaster() (*app.ForecasterService, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cpuMeta, err := loadModelMeta(cwd, "cpu_usage")
	if err != nil {
		return nil, fmt.Errorf("failed to load CPU model meta: %w", err)
	}

	ramMeta, err := loadModelMeta(cwd, "ram_usage")
	if err != nil {
		return nil, fmt.Errorf("failed to load RAM model meta: %w", err)
	}

	cpuScalerCfg, err := loadScalerConfig(cwd, "cpu_usage")
	if err != nil {
		return nil, fmt.Errorf("failed to load CPU scaler: %w", err)
	}

	cpuModel, err := tflite.NewPredictor(fmt.Sprintf("%s/models/global_trace_cpu_usage_15m_model.tflite", cwd))
	if err != nil {
		return nil, fmt.Errorf("failed to load CPU model: %w", err)
	}

	ramModel, err := tflite.NewPredictor(fmt.Sprintf("%s/models/global_trace_ram_usage_15m_model.tflite", cwd))
	if err != nil {
		return nil, fmt.Errorf("failed to load RAM model: %w", err)
	}

	fe := app.NewFeatureEngineer(cpuScalerCfg)

	return app.NewForecasterService(c.NodeRepo, c.MetricsRepo, fe, cpuModel, ramModel, cpuMeta, ramMeta), nil
}

func loadModelMeta(cwd, resource string) (*domain.ModelQuality, error) {
	path := fmt.Sprintf("%s/models/global_trace_%s_15m_meta.json", cwd, resource)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta domain.ModelQuality
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func loadScalerConfig(cwd, resource string) (domain.ModelConfig, error) {
	path := fmt.Sprintf("%s/models/global_trace_%s_15m_scaler.json", cwd, resource)
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.ModelConfig{}, err
	}
	var cfg domain.ModelConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.ModelConfig{}, err
	}
	return cfg, nil
}

func (c *Container) Close(_ context.Context) {
	_ = c.DB.Close()
}

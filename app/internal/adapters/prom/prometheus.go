package prom

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusAdapter struct {
	api v1.API
}

// NewAdapter створює новий екземпляр адаптера.
// Він задовольняє інтерфейс ports.MetricProvider.
func NewAdapter(url string) (*PrometheusAdapter, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create prometheus client: %w", err)
	}

	return &PrometheusAdapter{
		api: v1.NewAPI(client),
	}, nil
}

// GetHardwareFrames дістає ліміти контейнерів для формування Hardware Profile
func (a *PrometheusAdapter) GetHardwareFrames(ctx context.Context) ([]domain.Node, error) {
	// Запит ліміту пам'яті як базової метрики "існування" вузла
	val, _, err := a.api.Query(ctx, "docker_container_mem_limit", time.Now())
	if err != nil {
		return nil, err
	}

	vector, ok := val.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected prometheus response type: %T", val)
	}

	var nodes []domain.Node
	for _, sample := range vector {
		name := string(sample.Metric["container_name"])
		if a.isSystemContainer(name) {
			continue
		}

		nodes = append(nodes, domain.Node{
			ID:       name,
			Name:     name,
			CpuCores: 1.0, // Для MVP ставимо 1.0, або можна додати метрику n_cpus
			RamGB:    float64(sample.Value) / 1024 / 1024 / 1024,
		})
	}

	return nodes, nil
}

// GetLatestMetrics збирає поточні показники CPU та RAM
func (a *PrometheusAdapter) GetLatestMetrics(ctx context.Context) ([]domain.Metric, error) {
	now := time.Now()

	// Отримуємо CPU
	cpuVal, _, err := a.api.Query(ctx, "docker_container_cpu_usage_percent", now)
	if err != nil {
		return nil, err
	}

	// Отримуємо RAM
	ramVal, _, err := a.api.Query(ctx, "docker_container_mem_usage", now)
	if err != nil {
		return nil, err
	}

	cpuVector := cpuVal.(model.Vector)
	ramVector := ramVal.(model.Vector)

	// Мапа для мерджу метрик по імені контейнера
	metricsMap := make(map[string]*domain.Metric)

	for _, s := range cpuVector {
		name := string(s.Metric["container_name"])
		if a.isSystemContainer(name) {
			continue
		}
		metricsMap[name] = &domain.Metric{
			NodeID:     name,
			Timestamp:  now,
			CPUPercent: float64(s.Value),
		}
	}

	for _, s := range ramVector {
		name := string(s.Metric["container_name"])
		if m, ok := metricsMap[name]; ok {
			m.RAMBytes = int64(s.Value)
		}
	}

	// Перетворюємо мапу в слайс
	var result []domain.Metric
	for _, m := range metricsMap {
		result = append(result, *m)
	}

	return result, nil
}

// isSystemContainer фільтрує допоміжні сервіси
func (a *PrometheusAdapter) isSystemContainer(name string) bool {
	systemContainers := map[string]bool{
		"":             true,
		"prometheus":   true,
		"telegraf":     true,
		"docker-proxy": true,
	}
	return systemContainers[name]
}

func (a *PrometheusAdapter) GetRangeMetrics(ctx context.Context, nodeID string, start, end time.Time) ([]domain.Metric, error) {
	// 1. Визначаємо крок (Step).
	// Бажано, щоб він збігався з нашим інтервалом збору (5s).
	duration := end.Sub(start)
	step := 5 * time.Second
	if duration > 0 && duration/step > 10000 {
		step = duration / 10000
	}

	queryRange := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	// 2. Формуємо запити з фільтром по конкретному контейнеру
	cpuQuery := fmt.Sprintf(`docker_container_cpu_usage_percent{container_name="%s"}`, nodeID)
	ramQuery := fmt.Sprintf(`docker_container_mem_usage{container_name="%s"}`, nodeID)

	// 3. Виконуємо запити
	cpuRes, _, err := a.api.QueryRange(ctx, cpuQuery, queryRange)
	if err != nil {
		return nil, fmt.Errorf("cpu range query failed: %w", err)
	}

	ramRes, _, err := a.api.QueryRange(ctx, ramQuery, queryRange)
	if err != nil {
		return nil, fmt.Errorf("ram range query failed: %w", err)
	}

	// 4. Обробка результатів (Matrix -> Map для мерджу)
	// Використовуємо unix timestamp як ключ для синхронізації CPU та RAM
	dataMap := make(map[int64]*domain.Metric)

	// Парсимо CPU Matrix
	if matrix, ok := cpuRes.(model.Matrix); ok {
		for _, stream := range matrix {
			for _, sample := range stream.Values {
				ts := sample.Timestamp.Time().Unix()
				dataMap[ts] = &domain.Metric{
					NodeID:     nodeID,
					Timestamp:  sample.Timestamp.Time(),
					CPUPercent: float64(sample.Value),
				}
			}
		}
	}

	// Парсимо RAM Matrix та "підклеюємо" до існуючих точок
	if matrix, ok := ramRes.(model.Matrix); ok {
		for _, stream := range matrix {
			for _, sample := range stream.Values {
				ts := sample.Timestamp.Time().Unix()
				if m, exists := dataMap[ts]; exists {
					m.RAMBytes = int64(sample.Value)
				} else {
					// Якщо раптом для цього TS немає CPU, створюємо новий запис
					dataMap[ts] = &domain.Metric{
						NodeID:    nodeID,
						Timestamp: sample.Timestamp.Time(),
						RAMBytes:  int64(sample.Value),
					}
				}
			}
		}
	}

	// 5. Конвертуємо карту у відсортований слайс
	var result []domain.Metric
	for _, m := range dataMap {
		result = append(result, *m)
	}

	// Сортуємо за часом (опціонально, але корисно для ML)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result, nil
}

package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/tflite"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
)

type ForecasterService struct {
	nodeRepo    domain.NodeRepository
	metricsRepo domain.MetricsRepository
	fe          *FeatureEngineer
	predictor   *tflite.Predictor // Наш TFLite адаптер
	meta        *domain.ModelQuality
}

func NewForecasterService(nr domain.NodeRepository, mr domain.MetricsRepository, fe *FeatureEngineer, p *tflite.Predictor, meta *domain.ModelQuality) *ForecasterService {
	return &ForecasterService{
		nodeRepo:    nr,
		metricsRepo: mr,
		fe:          fe,
		predictor:   p,
		meta:        meta,
	}
}

func (s *ForecasterService) PredictForNode(ctx context.Context, nodeID string) (*domain.ForecastResponse, error) {
	metricCount := 61 // 60 для вікна + 1 для "прогріву"
	// Тепер нам потрібно 61 значення для розрахунку 60-хвилинного вікна.
	metrics, err := s.metricsRepo.GetLatestMetricsForNode(ctx, nodeID, metricCount)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	// 2. ПЕРЕВІРКА: Чи достатньо даних для "прогріву" вікон
	if len(metrics) < metricCount {
		return nil, fmt.Errorf("insufficient data for 60m window: have %d, need 61", len(metrics))
	}

	// 3. ПІДГОТОВКА ВЕКТОРА
	vector := s.fe.PrepareVector(metrics)
	if vector == nil {
		return nil, fmt.Errorf("feature engineer failed to prepare vector")
	}

	// 4. ІНФЕРЕНС
	rawOutputs, err := s.predictor.Predict(vector)
	if err != nil {
		return nil, fmt.Errorf("model prediction failed: %w", err)
	}

	// Перевірка, чи модель повернула хоча б 2 квантилі (q0.1 та q0.9)
	if len(rawOutputs) < 2 {
		return nil, fmt.Errorf("model returned unexpected number of outputs: %d", len(rawOutputs))
	}

	lower := float64(rawOutputs[0]) // q0.1
	upper := float64(rawOutputs[1]) // q0.9
	peak := upper                   // Згідно з ТЗ, ми фокусуємося на очікуваному піковому ресурсі

	// 5. ЛОГІКА ВИЗНАЧЕННЯ РИЗИКУ (використовуємо SafeBoundary з вузла)
	node, err := s.nodeRepo.GetNodeByID(ctx, nodeID)
	safeBoundary := 75.0
	if err == nil && node.SafeBoundary > 0 {
		safeBoundary = node.SafeBoundary
	}

	risk := "safe"
	if upper > safeBoundary {
		risk = "critical"
	}

	// 6. ФОРМУВАННЯ ВІДПОВІДІ (з використанням реальних метаданих якості)
	if s.meta == nil {
		s.meta = &domain.ModelQuality{Coverage: 0.85, MAE: 0, Horizon: 15}
	}
	log.Printf("DEBUG: meta.Coverage=%v MAE=%v Horizon=%v", s.meta.Coverage, s.meta.MAE, s.meta.Horizon)
	coverageLow := s.meta.Coverage < 0.8
	upperBoundExceeded := upper > 100.0
	upperClamped := upper
	if upper > 100.0 {
		upperClamped = 100.0
	}
	return &domain.ForecastResponse{
		NodeID:             nodeID,
		Timestamp:          time.Now().UTC(),
		HorizonMin:         s.meta.Horizon,
		CurrentCPU:         metrics[len(metrics)-1].CPUPercent,
		LowerBound:         lower,
		UpperBound:         upperClamped,
		Predicted:          peak,
		RiskLevel:          risk,
		Headroom:           100.0 - upper,
		Quality:            *s.meta,
		CoverageLow:        coverageLow,
		UpperBoundExceeded: upperBoundExceeded,
	}, nil
}

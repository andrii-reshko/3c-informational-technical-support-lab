package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/tflite"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
)

type resourcePrediction struct {
	current   float64
	lower     float64
	upper     float64
	predicted float64
	quality   domain.ModelQuality
}

type ForecasterService struct {
	nodeRepo    domain.NodeRepository
	metricsRepo domain.MetricsRepository
	fe          *FeatureEngineer
	cpuModel    *tflite.Predictor
	ramModel    *tflite.Predictor
	cpuMeta     *domain.ModelQuality
	ramMeta     *domain.ModelQuality
}

func NewForecasterService(nr domain.NodeRepository, mr domain.MetricsRepository, fe *FeatureEngineer, cpuModel *tflite.Predictor, ramModel *tflite.Predictor, cpuMeta, ramMeta *domain.ModelQuality) *ForecasterService {
	return &ForecasterService{
		nodeRepo:    nr,
		metricsRepo: mr,
		fe:          fe,
		cpuModel:    cpuModel,
		ramModel:    ramModel,
		cpuMeta:     cpuMeta,
		ramMeta:     ramMeta,
	}
}

func (s *ForecasterService) predictResource(ctx context.Context, nodeID string, model *tflite.Predictor, meta *domain.ModelQuality, isRAM bool) (*resourcePrediction, error) {
	metricCount := 61
	metrics, err := s.metricsRepo.GetLatestMetricsForNode(ctx, nodeID, metricCount)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	if len(metrics) < metricCount {
		return nil, fmt.Errorf("insufficient data: have %d, need %d", len(metrics), metricCount)
	}

	vector := s.fe.PrepareVector(metrics, isRAM)
	if vector == nil {
		return nil, fmt.Errorf("feature engineer failed to prepare vector")
	}

	rawOutputs, err := model.Predict(vector)
	if err != nil {
		return nil, fmt.Errorf("model prediction failed: %w", err)
	}

	if len(rawOutputs) < 2 {
		return nil, fmt.Errorf("model returned unexpected number of outputs: %d", len(rawOutputs))
	}

	quality := *meta
	if meta == nil {
		quality = domain.ModelQuality{Coverage: 0.85, MAE: 0, Horizon: 15}
	}
	if isRAM {
		quality.Resource = "ram_usage"
	} else {
		quality.Resource = "cpu_usage"
	}

	lastMetric := metrics[len(metrics)-1]
	var current float64
	if isRAM {
		current = lastMetric.RAMPercent
	} else {
		current = lastMetric.CPUPercent
	}

	return &resourcePrediction{
		current:   current,
		lower:     float64(rawOutputs[0]),
		upper:     float64(rawOutputs[1]),
		predicted: float64(rawOutputs[1]),
		quality:   quality,
	}, nil
}

func (s *ForecasterService) PredictForNode(ctx context.Context, nodeID string) (*domain.ForecastResponse, error) {
	cpuPred, err := s.predictResource(ctx, nodeID, s.cpuModel, s.cpuMeta, false)
	if err != nil {
		return nil, fmt.Errorf("CPU prediction failed: %w", err)
	}

	ramPred, err := s.predictResource(ctx, nodeID, s.ramModel, s.ramMeta, true)
	if err != nil {
		return nil, fmt.Errorf("RAM prediction failed: %w", err)
	}

	node, err := s.nodeRepo.GetNodeByID(ctx, nodeID)
	safeBoundary := 75.0
	if err == nil && node.SafeBoundary > 0 {
		safeBoundary = node.SafeBoundary
	}

	maxUpper := cpuPred.upper
	if ramPred.upper > maxUpper {
		maxUpper = ramPred.upper
	}

	risk := "safe"
	if maxUpper > safeBoundary {
		risk = "critical"
	}

	coverageLow := cpuPred.quality.Coverage < 0.8 || ramPred.quality.Coverage < 0.8
	upperBoundExceeded := maxUpper > 100.0

	headroom := 100.0 - maxUpper
	if headroom < 0 {
		headroom = 0
	}

	horizon := 15
	if s.cpuMeta != nil {
		horizon = s.cpuMeta.Horizon
	}

	log.Printf("DEBUG: CPU Coverage=%v RAM Coverage=%v", cpuPred.quality.Coverage, ramPred.quality.Coverage)

	return &domain.ForecastResponse{
		NodeID:             nodeID,
		Timestamp:          time.Now().UTC(),
		HorizonMin:         horizon,
		CPUCurrent:         cpuPred.current,
		CPULowerBound:      cpuPred.lower,
		CPUUpperBound:      cpuPred.upper,
		CPUPredicted:       cpuPred.predicted,
		CPUQuality:         cpuPred.quality,
		RAMCurrent:         ramPred.current,
		RAMLowerBound:      ramPred.lower,
		RAMUpperBound:      ramPred.upper,
		RAMPredicted:       ramPred.predicted,
		RAMQuality:         ramPred.quality,
		RiskLevel:          risk,
		Headroom:           headroom,
		Quality:            cpuPred.quality,
		CoverageLow:        coverageLow,
		UpperBoundExceeded: upperBoundExceeded,
	}, nil
}

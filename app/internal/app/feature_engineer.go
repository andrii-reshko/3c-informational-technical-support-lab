package app

import (
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
	log "github.com/sirupsen/logrus"
)

type FeatureEngineer struct {
	cfg domain.ModelConfig
}

func NewFeatureEngineer(cfg domain.ModelConfig) *FeatureEngineer {
	return &FeatureEngineer{cfg: cfg}
}

func (fe *FeatureEngineer) PrepareVector(metrics []domain.Metric, isRAM bool) []float32 {
	const minRequired = 61
	if len(metrics) < minRequired {
		log.Warnf("feature engineer: not enough data (%d/%d)", len(metrics), minRequired)
		return nil
	}

	n := len(metrics)
	current := metrics[n-1]
	raw := make([]float64, len(fe.cfg.Features))

	getValue := func(m domain.Metric) float64 {
		if isRAM {
			return m.RAMPercent
		}
		return m.CPUPercent
	}

	raw[domain.IdxResource] = getValue(current)
	raw[domain.IdxLag1] = getValue(metrics[n-2])
	raw[domain.IdxLag2] = getValue(metrics[n-3])
	raw[domain.IdxLag3] = getValue(metrics[n-4])
	raw[domain.IdxLag5] = getValue(metrics[n-6])
	raw[domain.IdxLag10] = getValue(metrics[n-11])
	raw[domain.IdxLag15] = getValue(metrics[n-16])
	raw[domain.IdxLag30] = getValue(metrics[n-31])

	calcRolling := func(window int) (float64, float64) {
		var sum, maxVal float64
		for i := n - window; i < n; i++ {
			val := getValue(metrics[i])
			sum += val
			if val > maxVal {
				maxVal = val
			}
		}
		return sum / float64(window), maxVal
	}

	raw[domain.IdxRollingMean5], raw[domain.IdxRollingMax5] = calcRolling(5)
	raw[domain.IdxRollingMean15], raw[domain.IdxRollingMax15] = calcRolling(15)
	raw[domain.IdxRollingMean60], raw[domain.IdxRollingMax60] = calcRolling(60)

	raw[domain.IdxHour] = float64(current.Timestamp.Hour())
	raw[domain.IdxDayOfWeek] = float64(current.Timestamp.Weekday())

	vector := make([]float32, len(raw))
	for i := range raw {
		norm := raw[i]*fe.cfg.Scale[i] + fe.cfg.Min[i]
		vector[i] = float32(norm)
	}

	return vector
}

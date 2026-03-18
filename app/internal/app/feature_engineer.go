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

func (fe *FeatureEngineer) PrepareVector(metrics []domain.Metric) []float32 {
	// Для вікна 60 хв та лага 30 нам потрібно мінімум 61 метрика
	const minRequired = 61
	if len(metrics) < minRequired {
		log.Warnf("feature engineer: not enough data (%d/%d)", len(metrics), minRequired)
		return nil
	}

	n := len(metrics)
	current := metrics[n-1]
	raw := make([]float64, len(fe.cfg.Features)) // Має бути 16 згідно domain

	// --- 1. Поточне значення та ЛАГИ ---
	raw[domain.IdxResource] = current.CPUPercent
	raw[domain.IdxLag1] = metrics[n-2].CPUPercent
	raw[domain.IdxLag2] = metrics[n-3].CPUPercent
	raw[domain.IdxLag3] = metrics[n-4].CPUPercent
	raw[domain.IdxLag5] = metrics[n-6].CPUPercent
	raw[domain.IdxLag10] = metrics[n-11].CPUPercent
	raw[domain.IdxLag15] = metrics[n-16].CPUPercent
	raw[domain.IdxLag30] = metrics[n-31].CPUPercent

	// --- 2. ROLLING STATISTICS (5, 15, 60 хв) ---
	calcRolling := func(window int) (float64, float64) {
		var sum, maxVal float64
		// Беремо останні 'window' точок
		for i := n - window; i < n; i++ {
			sum += metrics[i].CPUPercent
			if metrics[i].CPUPercent > maxVal {
				maxVal = metrics[i].CPUPercent
			}
		}
		return sum / float64(window), maxVal
	}

	raw[domain.IdxRollingMean5], raw[domain.IdxRollingMax5] = calcRolling(5)
	raw[domain.IdxRollingMean15], raw[domain.IdxRollingMax15] = calcRolling(15)
	raw[domain.IdxRollingMean60], raw[domain.IdxRollingMax60] = calcRolling(60)

	// --- 3. ЧАСОВІ ОЗНАКИ (UTC) ---
	raw[domain.IdxHour] = float64(current.Timestamp.Hour())
	raw[domain.IdxDayOfWeek] = float64(current.Timestamp.Weekday())

	// --- 4. НОРМАЛІЗАЦІЯ (MinMaxScaler) ---
	vector := make([]float32, len(raw))
	for i := range raw {
		// Формула sklearn: X_std = (X - X.min(axis=0)) / (X.max(axis=0) - X.min(axis=0))
		// X_scaled = X_std * (max - min) + min
		// Або спрощено: X * scale + min_ (де min_ це зсув, що включає віднімання мінімуму)
		// У конфігурації Min - це саме sklearn.min_, тому віднімати DataMin ще раз не потрібно.
		norm := raw[i]*fe.cfg.Scale[i] + fe.cfg.Min[i]
		vector[i] = float32(norm)
	}

	return vector
}

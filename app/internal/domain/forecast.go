package domain

import "time"

// ModelConfig зберігає параметри нормалізації MinMaxScaler із Python
type ModelConfig struct {
	Features []string  `json:"features"`
	Scale    []float64 `json:"scale"`
	Min      []float64 `json:"min"`
	DataMin  []float64 `json:"data_min"`
	DataMax  []float64 `json:"data_max"`
}

// Порядок ознак у векторі. МАЄ СУВОРО ВІДПОВІДАТИ ПОРЯДКУ У Python build_features()!
const (
	IdxResource = iota // Поточне значення (cpu_usage або ram_usage)
	IdxLag1
	IdxLag2
	IdxLag3
	IdxLag5
	IdxLag10
	IdxLag15
	IdxLag30 // Лаги згідно ТЗ

	// Rolling stats для трьох вікон: 5, 15, 60 хвилин
	IdxRollingMean5
	IdxRollingMax5
	IdxRollingMean15
	IdxRollingMax15
	IdxRollingMean60
	IdxRollingMax60

	IdxHour
	IdxDayOfWeek
)

// ForecastResponse - повний пакет даних для UI
type ForecastResponse struct {
	NodeID     string    `json:"node_id"`
	Timestamp  time.Time `json:"timestamp"`
	HorizonMin int       `json:"horizon_minutes"` // MVP: 15 хв

	// CPU Forecast
	CPUCurrent    float64      `json:"cpu_current_percent"`
	CPULowerBound float64      `json:"cpu_lower_bound"`
	CPUUpperBound float64      `json:"cpu_upper_bound"`
	CPUPredicted  float64      `json:"cpu_predicted_peak"`
	CPUQuality    ModelQuality `json:"cpu_quality"`

	// RAM Forecast
	RAMCurrent    float64      `json:"ram_current_percent"`
	RAMLowerBound float64      `json:"ram_lower_bound"`
	RAMUpperBound float64      `json:"ram_upper_bound"`
	RAMPredicted  float64      `json:"ram_predicted_peak"`
	RAMQuality    ModelQuality `json:"ram_quality"`

	// Combined risk assessment
	RiskLevel string  `json:"risk_level"` // safe | warning | critical
	Headroom  float64 `json:"headroom_percent"`

	// Метрики якості (дефолтні, якщо CPU/RAM quality недоступні)
	Quality ModelQuality `json:"quality"`

	// Додано для UI попереджень
	CoverageLow        bool `json:"coverage_low"`         // Coverage < 80%
	UpperBoundExceeded bool `json:"upper_bound_exceeded"` // Upper > 100%
}

type ModelQuality struct {
	MAE           float64 `json:"mae"`      // Середня абсолютна помилка
	Coverage      float64 `json:"coverage"` // Відсоток попадання у довірчий інтервал
	Resource      string  `json:"resource"` // "cpu_usage" або "ram_usage"
	Horizon       int     `json:"horizon"`
	FeaturesCount int     `json:"features_count"`
}

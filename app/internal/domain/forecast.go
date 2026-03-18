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

	// Поточні значення
	CurrentCPU float64 `json:"current_cpu_percent"`
	CurrentRAM float64 `json:"current_ram_percent"`

	// Інтервальний прогноз (Quantile Regression 0.1 та 0.9)
	LowerBound float64 `json:"lower_bound"`    // q0.1
	UpperBound float64 `json:"upper_bound"`    // q0.9
	Predicted  float64 `json:"predicted_peak"` // Очікуваний пік

	// Аналіз ризиків для Dashboard
	RiskLevel string  `json:"risk_level"` // safe | warning | critical
	Headroom  float64 `json:"headroom_percent"`

	// Метрики якості для детального перегляду (Node Details)
	Quality ModelQuality `json:"quality"`
}

type ModelQuality struct {
	MAE           float64 `json:"mae"`      // Середня абсолютна помилка
	Coverage      float64 `json:"coverage"` // Відсоток попадання у довірчий інтервал
	Resource      string  `json:"resource"` // "cpu_usage" або "ram_usage"
	Horizon       int     `json:"horizon"`
	FeaturesCount int     `json:"features_count"`
}

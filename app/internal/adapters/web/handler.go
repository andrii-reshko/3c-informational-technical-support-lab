package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/app"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
	"github.com/gorilla/websocket"
)

type Handler struct {
	forecaster  *app.ForecasterService
	nodeRepo    domain.NodeRepository
	metricsRepo domain.MetricsRepository
	tmpl        *template.Template
}

type dashboardNodeView struct {
	ID             string
	Name           string
	Risk           string
	RiskClass      string
	CPU            float64
	RAM            float64
	Predicted      float64
	PredictedLower float64
	PredictedUpper float64
	Horizon        int
	SparkJSON      template.JS
}

type dashboardViewData struct {
	Nodes            []dashboardNodeView
	InitialNodesJSON template.JS
}

type initialNode struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Risk           string      `json:"risk"`
	RiskClass      string      `json:"risk_class"`
	CPU            float64     `json:"cpu"`
	RAM            float64     `json:"ram"`
	Predicted      float64     `json:"predicted"`
	PredictedLower float64     `json:"pred_lower"`
	PredictedUpper float64     `json:"pred_upper"`
	Horizon        int         `json:"horizon"`
	SparkJSON      template.JS `json:"spark_json"`
}

type nodePageData struct {
	NodeID             string
	NodeName           string
	CPUCores           float64
	RAMGB              float64
	SafeBoundary       float64
	CurrentCPU         float64
	CurrentRAM         float64
	LowerBound         float64
	UpperBound         float64
	Predicted          float64
	Headroom           float64
	PredictedCPUCores  float64
	PredictedRAMGB     float64
	HeadroomCPUCores   float64
	HeadroomRAMGB      float64
	Risk               string
	RiskClass          string
	Horizon            int
	QualityMAE         float64
	QualityCoverage    float64
	QualityLowCoverage bool
	ChartJSON          template.JS
	NowISO             string
	// Нові поля для попереджень
	CoverageLow        bool
	UpperBoundExceeded bool
}

type chartPoint struct {
	Timestamp string  `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	Lower     float64 `json:"lower"`
	Upper     float64 `json:"upper"`
}

type wsPoint struct {
	Timestamp string  `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	RAM       float64 `json:"ram"`
	Lower     float64 `json:"lower"`
	Upper     float64 `json:"upper"`
	NowISO    string  `json:"now_iso"`
}

type wsDashboardNode struct {
	ID             string  `json:"id"`
	Risk           string  `json:"risk"`
	RiskClass      string  `json:"risk_class"`
	CPU            float64 `json:"cpu"`
	RAM            float64 `json:"ram"`
	PredictedLower float64 `json:"pred_lower"`
	PredictedUpper float64 `json:"pred_upper"`
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewHandler(f *app.ForecasterService, nr domain.NodeRepository, mr domain.MetricsRepository) (*Handler, error) {
	tmpl, err := template.ParseGlob("internal/adapters/web/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{forecaster: f, nodeRepo: nr, metricsRepo: mr, tmpl: tmpl}, nil
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.nodeRepo.GetAllNodes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	end := time.Now().UTC()
	start := end.Add(-1 * time.Hour)

	items := make([]dashboardNodeView, 0, len(nodes))
	for _, n := range nodes {
		pred, predErr := h.forecaster.PredictForNode(r.Context(), n.ID)
		if predErr != nil {
			continue
		}

		history, _ := h.metricsRepo.GetRangeMetrics(r.Context(), n.ID, start, end)

		sparkPoints := buildSparkPoints(history, pred, end)
		sparkPayload, _ := json.Marshal(sparkPoints)

		name := n.Name
		if strings.TrimSpace(name) == "" {
			name = n.ID
		}

		items = append(items, dashboardNodeView{
			ID:             n.ID,
			Name:           name,
			Risk:           strings.ToUpper(pred.RiskLevel),
			RiskClass:      riskClass(pred.RiskLevel),
			CPU:            pred.CurrentCPU,
			RAM:            pred.CurrentRAM,
			Predicted:      pred.Predicted,
			PredictedLower: pred.LowerBound,
			PredictedUpper: pred.UpperBound,
			Horizon:        pred.HorizonMin,
			SparkJSON:      template.JS(sparkPayload),
		})
	}

	initialNodes := make([]initialNode, 0, len(items))
	for _, item := range items {
		initialNodes = append(initialNodes, initialNode{
			ID:             item.ID,
			Name:           item.Name,
			Risk:           item.Risk,
			RiskClass:      item.RiskClass,
			CPU:            item.CPU,
			RAM:            item.RAM,
			Predicted:      item.Predicted,
			PredictedLower: item.PredictedLower,
			PredictedUpper: item.PredictedUpper,
			Horizon:        item.Horizon,
			SparkJSON:      item.SparkJSON,
		})
	}
	initialJSON, _ := json.Marshal(initialNodes)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "dashboard.html", dashboardViewData{Nodes: items, InitialNodesJSON: template.JS(initialJSON)})
}

func (h *Handler) NodeDetails(w http.ResponseWriter, r *http.Request) {
	nodeID := strings.TrimPrefix(path.Clean(r.URL.Path), "/ui/nodes/")
	if nodeID == "" || nodeID == "." || strings.Contains(nodeID, "/") {
		http.NotFound(w, r)
		return
	}

	n, err := h.nodeRepo.GetNodeByID(r.Context(), nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pred, err := h.forecaster.PredictForNode(r.Context(), nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	end := time.Now().UTC()
	start := end.Add(-1 * time.Hour)
	history, err := h.metricsRepo.GetRangeMetrics(r.Context(), nodeID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	points := buildPOCChartPoints(history, pred, end)
	payload, _ := json.Marshal(points)

	name := n.Name
	if strings.TrimSpace(name) == "" {
		name = n.ID
	}

	predCPUCores := n.CpuCores * pred.Predicted / 100.0
	headroomCPUCores := n.CpuCores - predCPUCores
	if headroomCPUCores < 0 {
		headroomCPUCores = 0
	}

	predRAMGB := n.RamGB * pred.Predicted / 100.0
	headroomRAMGB := n.RamGB - predRAMGB
	if headroomRAMGB < 0 {
		headroomRAMGB = 0
	}

	coveragePct := normalizeCoveragePercent(pred.Quality.Coverage)

	data := nodePageData{
		NodeID:             n.ID,
		NodeName:           name,
		CPUCores:           n.CpuCores,
		RAMGB:              n.RamGB,
		SafeBoundary:       safeBoundaryOrDefault(n.SafeBoundary),
		CurrentCPU:         pred.CurrentCPU,
		CurrentRAM:         pred.CurrentRAM,
		LowerBound:         pred.LowerBound,
		UpperBound:         pred.UpperBound,
		Predicted:          pred.Predicted,
		Headroom:           pred.Headroom,
		PredictedCPUCores:  predCPUCores,
		PredictedRAMGB:     predRAMGB,
		HeadroomCPUCores:   headroomCPUCores,
		HeadroomRAMGB:      headroomRAMGB,
		Risk:               strings.ToUpper(pred.RiskLevel),
		RiskClass:          riskClass(pred.RiskLevel),
		Horizon:            pred.HorizonMin,
		QualityMAE:         pred.Quality.MAE,
		QualityCoverage:    coveragePct,
		QualityLowCoverage: coveragePct < 80.0,
		ChartJSON:          template.JS(payload),
		NowISO:             end.Format(time.RFC3339),
		// Нові поля для попереджень
		CoverageLow:        pred.CoverageLow,
		UpperBoundExceeded: pred.UpperBoundExceeded,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "node.html", data)
}

// WSStream streams live forecast updates via WebSocket.
// Query params: node_id (optional).
// If node_id is provided, streams single node details (wsPoint).
// If node_id is empty, streams dashboard data ([]wsDashboardNode).
func (h *Handler) WSStream(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")

	intervalSec := 2
	if raw := r.URL.Query().Get("interval_sec"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 60 {
			intervalSec = v
		}
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	if strings.TrimSpace(nodeID) == "" {
		h.streamDashboard(r, conn, intervalSec)
	} else {
		h.streamSingleNode(r, conn, nodeID, intervalSec)
	}
}

func (h *Handler) streamSingleNode(r *http.Request, conn *websocket.Conn, nodeID string, intervalSec int) {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	push := func() error {
		pred, err := h.forecaster.PredictForNode(r.Context(), nodeID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		return conn.WriteJSON(wsPoint{
			Timestamp: now.Format(time.RFC3339),
			CPU:       pred.CurrentCPU,
			RAM:       pred.CurrentRAM,
			Lower:     pred.LowerBound,
			Upper:     pred.UpperBound,
			NowISO:    now.Format(time.RFC3339),
		})
	}

	if err := push(); err != nil {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if err := push(); err != nil {
				return
			}
		}
	}
}

func (h *Handler) streamDashboard(r *http.Request, conn *websocket.Conn, intervalSec int) {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	push := func() error {
		nodes, err := h.nodeRepo.GetAllNodes(r.Context())
		if err != nil {
			return err
		}

		payload := make([]wsDashboardNode, 0, len(nodes))
		for _, n := range nodes {
			pred, err := h.forecaster.PredictForNode(r.Context(), n.ID)
			if err != nil {
				continue
			}
			payload = append(payload, wsDashboardNode{
				ID:             n.ID,
				Risk:           pred.RiskLevel,
				RiskClass:      riskClass(pred.RiskLevel),
				CPU:            pred.CurrentCPU,
				RAM:            pred.CurrentRAM,
				PredictedLower: pred.LowerBound,
				PredictedUpper: pred.UpperBound,
			})
		}
		return conn.WriteJSON(payload)
	}

	if err := push(); err != nil {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if err := push(); err != nil {
				return
			}
		}
	}
}

func normalizeCoveragePercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v <= 1.0 {
		return v * 100.0
	}
	if v > 100 {
		return 100
	}
	return v
}

func safeBoundaryOrDefault(v float64) float64 {
	if v <= 0 {
		return 75.0
	}
	return v
}

func buildPOCChartPoints(history []domain.Metric, pred *domain.ForecastResponse, now time.Time) []chartPoint {
	start := now.Add(-1 * time.Hour)
	filtered := make([]domain.Metric, 0, len(history))
	for _, m := range history {
		t := m.Timestamp.UTC()
		if !t.Before(start) && !t.After(now) {
			filtered = append(filtered, m)
		}
	}

	points := make([]chartPoint, 0, len(filtered)+5)
	for _, m := range filtered {
		points = append(points, chartPoint{
			Timestamp: m.Timestamp.UTC().Format(time.RFC3339),
			CPU:       m.CPUPercent,
			Lower:     m.CPUPercent,
			Upper:     m.CPUPercent,
		})
	}

	// anchor at now — bridge between history and forecast
	points = append(points, chartPoint{
		Timestamp: now.Format(time.RFC3339),
		CPU:       pred.CurrentCPU,
		Lower:     pred.CurrentCPU,
		Upper:     pred.CurrentCPU,
	})

	// forecast fan: now+0 .. now+horizon
	steps := []struct{ min int }{
		{pred.HorizonMin / 3},
		{2 * pred.HorizonMin / 3},
		{pred.HorizonMin},
	}
	for _, s := range steps {
		frac := float64(s.min) / float64(pred.HorizonMin)
		t := now.Add(time.Duration(s.min) * time.Minute)
		points = append(points, chartPoint{
			Timestamp: t.Format(time.RFC3339),
			CPU:       pred.CurrentCPU + (pred.Predicted-pred.CurrentCPU)*frac,
			Lower:     pred.CurrentCPU + (pred.LowerBound-pred.CurrentCPU)*frac,
			Upper:     pred.CurrentCPU + (pred.UpperBound-pred.CurrentCPU)*frac,
		})
	}

	return points
}

func buildSparkPoints(history []domain.Metric, pred *domain.ForecastResponse, now time.Time) []chartPoint {
	start := now.Add(-1 * time.Hour)
	points := make([]chartPoint, 0, len(history)+2)
	for _, m := range history {
		t := m.Timestamp.UTC()
		if t.Before(start) || t.After(now) {
			continue
		}
		points = append(points, chartPoint{
			Timestamp: t.Format(time.RFC3339),
			CPU:       m.CPUPercent,
			Lower:     m.CPUPercent,
			Upper:     m.CPUPercent,
		})
	}
	points = append(points, chartPoint{
		Timestamp: now.Format(time.RFC3339),
		CPU:       pred.CurrentCPU,
		Lower:     pred.CurrentCPU,
		Upper:     pred.CurrentCPU,
	})
	points = append(points, chartPoint{
		Timestamp: now.Add(time.Duration(pred.HorizonMin) * time.Minute).Format(time.RFC3339),
		CPU:       pred.Predicted,
		Lower:     pred.LowerBound,
		Upper:     pred.UpperBound,
	})
	return points
}

func riskClass(risk string) string {
	switch strings.ToLower(risk) {
	case "critical":
		return "text-bg-danger"
	default:
		return "text-bg-success"
	}
}

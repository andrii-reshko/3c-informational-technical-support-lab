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
	ID           string
	Name         string
	Risk         string
	RiskClass    string
	CPU          float64
	RAM          float64
	CPUPredLower float64
	CPUPredUpper float64
	RAMPredLower float64
	RAMPredUpper float64
	Horizon      int
	SparkJSON    template.JS
}

type dashboardViewData struct {
	Nodes            []dashboardNodeView
	InitialNodesJSON template.JS
}

type initialNode struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Risk         string      `json:"risk"`
	RiskClass    string      `json:"risk_class"`
	CPU          float64     `json:"cpu"`
	RAM          float64     `json:"ram"`
	CPUPredLower float64     `json:"cpu_pred_lower"`
	CPUPredUpper float64     `json:"cpu_pred_upper"`
	RAMPredLower float64     `json:"ram_pred_lower"`
	RAMPredUpper float64     `json:"ram_pred_upper"`
	Horizon      int         `json:"horizon"`
	SparkJSON    template.JS `json:"spark_json"`
}

type nodePageData struct {
	NodeID       string
	NodeName     string
	CPUCores     float64
	RAMGB        float64
	SafeBoundary float64

	CurrentCPU float64
	CurrentRAM float64

	CPULowerBound float64
	CPUUpperBound float64

	RAMLowerBound float64
	RAMUpperBound float64

	Headroom float64

	PredictedCPUCores float64
	PredictedRAMGB    float64
	HeadroomCPUCores  float64
	HeadroomRAMGB     float64

	Risk      string
	RiskClass string
	Horizon   int

	CPUQualityMAE      float64
	CPUQualityCoverage float64
	RAMQualityMAE      float64
	RAMQualityCoverage float64
	QualityLowCoverage bool
	ChartJSON          template.JS
	RAMChartJSON       template.JS
	NowISO             string
	CoverageLow        bool
	UpperBoundExceeded bool
}

type chartPoint struct {
	Timestamp string  `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	RAM       float64 `json:"ram"`
	Lower     float64 `json:"lower"`
	Upper     float64 `json:"upper"`
}

type wsPoint struct {
	Timestamp string  `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	RAM       float64 `json:"ram"`
	CPULower  float64 `json:"cpu_lower"`
	CPUUpper  float64 `json:"cpu_upper"`
	RAMLower  float64 `json:"ram_lower"`
	RAMUpper  float64 `json:"ram_upper"`
	NowISO    string  `json:"now_iso"`
}

type wsDashboardNode struct {
	ID        string  `json:"id"`
	Risk      string  `json:"risk"`
	RiskClass string  `json:"risk_class"`
	CPU       float64 `json:"cpu"`
	RAM       float64 `json:"ram"`
	CPULower  float64 `json:"cpu_lower"`
	CPUUpper  float64 `json:"cpu_upper"`
	RAMLower  float64 `json:"ram_lower"`
	RAMUpper  float64 `json:"ram_upper"`
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
			ID:           n.ID,
			Name:         name,
			Risk:         strings.ToUpper(pred.RiskLevel),
			RiskClass:    riskClass(pred.RiskLevel),
			CPU:          pred.CPUCurrent,
			RAM:          pred.RAMCurrent,
			CPUPredLower: pred.CPULowerBound,
			CPUPredUpper: pred.CPUUpperBound,
			RAMPredLower: pred.RAMLowerBound,
			RAMPredUpper: pred.RAMUpperBound,
			Horizon:      pred.HorizonMin,
			SparkJSON:    template.JS(sparkPayload),
		})
	}

	initialNodes := make([]initialNode, 0, len(items))
	for _, item := range items {
		initialNodes = append(initialNodes, initialNode{
			ID:           item.ID,
			Name:         item.Name,
			Risk:         item.Risk,
			RiskClass:    item.RiskClass,
			CPU:          item.CPU,
			RAM:          item.RAM,
			CPUPredLower: item.CPUPredLower,
			CPUPredUpper: item.CPUPredUpper,
			RAMPredLower: item.RAMPredLower,
			RAMPredUpper: item.RAMPredUpper,
			Horizon:      item.Horizon,
			SparkJSON:    item.SparkJSON,
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

	cpuPoints := buildPOCChartPoints(history, pred, end, false)
	cpuPayload, _ := json.Marshal(cpuPoints)

	ramPoints := buildPOCChartPoints(history, pred, end, true)
	ramPayload, _ := json.Marshal(ramPoints)

	name := n.Name
	if strings.TrimSpace(name) == "" {
		name = n.ID
	}

	predCPUCores := n.CpuCores * pred.CPUUpperBound / 100.0
	headroomCPUCores := n.CpuCores - predCPUCores
	if headroomCPUCores < 0 {
		headroomCPUCores = 0
	}

	predRAMGB := n.RamGB * pred.RAMUpperBound / 100.0
	headroomRAMGB := n.RamGB - predRAMGB
	if headroomRAMGB < 0 {
		headroomRAMGB = 0
	}

	cpuCoverage := normalizeCoveragePercent(pred.CPUQuality.Coverage)
	ramCoverage := normalizeCoveragePercent(pred.RAMQuality.Coverage)
	coverageLow := cpuCoverage < 80.0 || ramCoverage < 80.0

	data := nodePageData{
		NodeID:       n.ID,
		NodeName:     name,
		CPUCores:     n.CpuCores,
		RAMGB:        n.RamGB,
		SafeBoundary: safeBoundaryOrDefault(n.SafeBoundary),

		CurrentCPU: pred.CPUCurrent,
		CurrentRAM: pred.RAMCurrent,

		CPULowerBound: pred.CPULowerBound,
		CPUUpperBound: pred.CPUUpperBound,

		RAMLowerBound: pred.RAMLowerBound,
		RAMUpperBound: pred.RAMUpperBound,

		Headroom:           pred.Headroom,
		PredictedCPUCores:  predCPUCores,
		PredictedRAMGB:     predRAMGB,
		HeadroomCPUCores:   headroomCPUCores,
		HeadroomRAMGB:      headroomRAMGB,
		Risk:               strings.ToUpper(pred.RiskLevel),
		RiskClass:          riskClass(pred.RiskLevel),
		Horizon:            pred.HorizonMin,
		CPUQualityMAE:      pred.CPUQuality.MAE,
		CPUQualityCoverage: cpuCoverage,
		RAMQualityMAE:      pred.RAMQuality.MAE,
		RAMQualityCoverage: ramCoverage,
		QualityLowCoverage: coverageLow,
		ChartJSON:          template.JS(cpuPayload),
		RAMChartJSON:       template.JS(ramPayload),
		NowISO:             end.Format(time.RFC3339),
		CoverageLow:        coverageLow,
		UpperBoundExceeded: pred.UpperBoundExceeded,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "node.html", data)
}

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
			CPU:       pred.CPUCurrent,
			RAM:       pred.RAMCurrent,
			CPULower:  pred.CPULowerBound,
			CPUUpper:  pred.CPUUpperBound,
			RAMLower:  pred.RAMLowerBound,
			RAMUpper:  pred.RAMUpperBound,
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
				ID:        n.ID,
				Risk:      pred.RiskLevel,
				RiskClass: riskClass(pred.RiskLevel),
				CPU:       pred.CPUCurrent,
				RAM:       pred.RAMCurrent,
				CPULower:  pred.CPULowerBound,
				CPUUpper:  pred.CPUUpperBound,
				RAMLower:  pred.RAMLowerBound,
				RAMUpper:  pred.RAMUpperBound,
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

func buildPOCChartPoints(history []domain.Metric, pred *domain.ForecastResponse, now time.Time, isRAM bool) []chartPoint {
	start := now.Add(-1 * time.Hour)
	filtered := make([]domain.Metric, 0, len(history))
	for _, m := range history {
		t := m.Timestamp.UTC()
		if !t.Before(start) && !t.After(now) {
			filtered = append(filtered, m)
		}
	}

	getCurrent := func(m domain.Metric) float64 {
		if isRAM {
			return m.RAMPercent
		}
		return m.CPUPercent
	}

	getPredicted := func() float64 {
		if isRAM {
			return pred.RAMCurrent
		}
		return pred.CPUCurrent
	}

	getLower := func() float64 {
		if isRAM {
			return pred.RAMLowerBound
		}
		return pred.CPULowerBound
	}

	getUpper := func() float64 {
		if isRAM {
			return pred.RAMUpperBound
		}
		return pred.CPUUpperBound
	}

	points := make([]chartPoint, 0, len(filtered)+5)
	for _, m := range filtered {
		cur := getCurrent(m)
		points = append(points, chartPoint{
			Timestamp: m.Timestamp.UTC().Format(time.RFC3339),
			CPU:       m.CPUPercent,
			RAM:       m.RAMPercent,
			Lower:     cur,
			Upper:     cur,
		})
	}

	cur := getPredicted()
	points = append(points, chartPoint{
		Timestamp: now.Format(time.RFC3339),
		CPU:       pred.CPUCurrent,
		RAM:       pred.RAMCurrent,
		Lower:     cur,
		Upper:     cur,
	})

	steps := []struct{ min int }{
		{pred.HorizonMin / 3},
		{2 * pred.HorizonMin / 3},
		{pred.HorizonMin},
	}
	lower := getLower()
	upper := getUpper()
	for _, s := range steps {
		frac := float64(s.min) / float64(pred.HorizonMin)
		t := now.Add(time.Duration(s.min) * time.Minute)
		points = append(points, chartPoint{
			Timestamp: t.Format(time.RFC3339),
			CPU:       pred.CPUCurrent + (pred.CPUUpperBound-pred.CPUCurrent)*frac,
			RAM:       pred.RAMCurrent + (pred.RAMUpperBound-pred.RAMCurrent)*frac,
			Lower:     cur + (lower-cur)*frac,
			Upper:     cur + (upper-cur)*frac,
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
			RAM:       m.RAMPercent,
			Lower:     m.CPUPercent,
			Upper:     m.CPUPercent,
		})
	}
	points = append(points, chartPoint{
		Timestamp: now.Format(time.RFC3339),
		CPU:       pred.CPUCurrent,
		RAM:       pred.RAMCurrent,
		Lower:     pred.CPUCurrent,
		Upper:     pred.CPUCurrent,
	})
	points = append(points, chartPoint{
		Timestamp: now.Add(time.Duration(pred.HorizonMin) * time.Minute).Format(time.RFC3339),
		CPU:       pred.CPUUpperBound,
		RAM:       pred.RAMUpperBound,
		Lower:     pred.CPULowerBound,
		Upper:     pred.CPUUpperBound,
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

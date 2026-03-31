package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/app"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
)

type Handler struct {
	forecaster *app.ForecasterService
	nodeRepo   domain.NodeRepository
}

func NewHandler(f *app.ForecasterService, nr domain.NodeRepository) *Handler {
	return &Handler{forecaster: f, nodeRepo: nr}
}

func (h *Handler) GetPrediction(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		http.Error(w, "missing node_id", http.StatusBadRequest)
		return
	}

	val, err := h.forecaster.PredictForNode(r.Context(), nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"node_id":    nodeID,
		"prediction": val,
		"unit":       "percent",
	})

	if err != nil {
		panic(err)
	}
}

type UpdateNodeRequest struct {
	Name         string `json:"name"`
	CpuCores     any    `json:"cpu_cores"`
	RamGB        any    `json:"ram_gb"`
	SafeBoundary any    `json:"safe_boundary"`
}

func (h *Handler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		http.Error(w, "missing node_id", http.StatusBadRequest)
		return
	}

	var req UpdateNodeRequest
	bodyBytes, _ := io.ReadAll(r.Body)
	fmt.Printf("DEBUG UpdateNode: body=%s\n", string(bodyBytes))
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	fmt.Printf("DEBUG UpdateNode: parsed req=%+v\n", req)

	// Parse CPU cores with validation
	cpuCores := 1.0
	switch v := req.CpuCores.(type) {
	case float64:
		if v <= 0 {
			http.Error(w, "Invalid input format for CPU Cores: must be > 0", http.StatusBadRequest)
			return
		}
		cpuCores = v
	case string:
		if _, err := fmt.Sscanf(v, "%f", &cpuCores); err != nil || cpuCores <= 0 {
			http.Error(w, "Invalid input format for CPU Cores: must be a positive number", http.StatusBadRequest)
			return
		}
	}

	// Parse RAM GB with validation
	ramGB := 1.0
	switch v := req.RamGB.(type) {
	case float64:
		if v <= 0 {
			http.Error(w, "Invalid input format for RAM GB: must be > 0", http.StatusBadRequest)
			return
		}
		ramGB = v
	case string:
		if _, err := fmt.Sscanf(v, "%f", &ramGB); err != nil || ramGB <= 0 {
			http.Error(w, "Invalid input format for RAM GB: must be a positive number", http.StatusBadRequest)
			return
		}
	}

	// Parse SafeBoundary with validation
	safeBoundary := 75.0
	switch v := req.SafeBoundary.(type) {
	case float64:
		if v < 1 || v > 100 {
			http.Error(w, "Invalid input format for Safe Boundary (%): must be between 1 and 100", http.StatusBadRequest)
			return
		}
		safeBoundary = v
	case string:
		var sb float64
		if _, err := fmt.Sscanf(v, "%f", &sb); err != nil || sb < 1 || sb > 100 {
			http.Error(w, "Invalid input format for Safe Boundary (%): must be a number between 1 and 100", http.StatusBadRequest)
			return
		}
		safeBoundary = sb
	}

	node := domain.Node{
		ID:           nodeID,
		Name:         req.Name,
		CpuCores:     cpuCores,
		RamGB:        ramGB,
		SafeBoundary: safeBoundary,
	}

	if err := h.nodeRepo.UpdateNode(r.Context(), node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

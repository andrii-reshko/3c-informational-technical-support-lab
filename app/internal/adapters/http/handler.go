package http

import (
	"encoding/json"
	"net/http"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/app"
)

type Handler struct {
	forecaster *app.ForecasterService
}

func NewHandler(f *app.ForecasterService) *Handler {
	return &Handler{forecaster: f}
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

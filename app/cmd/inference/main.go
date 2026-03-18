package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	http2 "github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/http"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/web"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/app"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/bootstrap"
	log "github.com/sirupsen/logrus"
)

func main() {
	time.Local = time.UTC

	c := bootstrap.NewContainer()
	log.Info("inference: boot")

	// 1. Завантажуємо налаштування нормалізації (той самий JSON)
	// modelCfgFile, _ := os.ReadFile("assets/model_config.json")

	// 2. Ініціалізуємо TFLite Predictor
	modelId := "global_trace_cpu_usage_15m"
	ml := app.NewLoader([]string{modelId})
	model, err := ml.Get(modelId)
	if err != nil {
		log.Fatalf("inference: failed to load TFLite model: %v", err)
	}

	// 3. Збираємо сервіси
	fe := app.NewFeatureEngineer(*model.Scaler)
	forecaster := app.NewForecasterService(c.NodeRepo, c.MetricsRepo, fe, model.Predictor, model.Meta)
	handler := http2.NewHandler(forecaster)
	webHandler, err := web.NewHandler(forecaster, c.NodeRepo, c.MetricsRepo)
	if err != nil {
		log.Fatalf("inference: failed to init web handler: %v", err)
	}

	// 4. Налаштовуємо роутинг
	mux := http.NewServeMux()
	mux.HandleFunc("/predict", handler.GetPrediction)
	mux.HandleFunc("/ui", webHandler.Dashboard)
	mux.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/ui/" || r.URL.Path == "/ui":
			webHandler.Dashboard(w, r)
		case r.URL.Path == "/ui/nodes" || r.URL.Path == "/ui/nodes/":
			webHandler.Dashboard(w, r)
		case len(r.URL.Path) >= len("/ui/nodes/") && r.URL.Path[:len("/ui/nodes/")] == "/ui/nodes/":
			webHandler.NodeDetails(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	mux.Handle("/ui/static/", http.StripPrefix("/ui/static/", http.FileServer(http.Dir("internal/adapters/web/static"))))
	mux.HandleFunc("/ws", webHandler.WSStream)

	server := &http.Server{
		Addr:    ":" + "9001",
		Handler: mux,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Infof("inference: listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("inference: listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info("inference: shutting down")

	off, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(off)
	c.Close(off)

	log.Info("inference: stopped")
}

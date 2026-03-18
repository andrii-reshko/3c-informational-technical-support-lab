package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/bootstrap"
	log "github.com/sirupsen/logrus"
)

func main() {
	time.Local = time.UTC

	c := bootstrap.NewContainer()
	log.Info("collector: boot")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("collector: sync nodes")
	c.Collector.SyncNodes(ctx)
	log.Info("collector: backfill data")
	c.Collector.Backfill(ctx)

	ticker := time.NewTicker(c.Config.CollectInterval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(24 * time.Hour) // once a day
	defer cleanupTicker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				c.Collector.Collect(ctx)
			case <-cleanupTicker.C:
				c.Collector.Cleanup(ctx, c.Config.RetentionDays)
			case <-ctx.Done():
				log.Debugf("collector: signal received %v", ctx.Err())
				return
			}
		}
	}()

	<-ctx.Done()
	log.Info("collector: shutting down")

	off, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.Close(off)

	log.Info("collector stopped")
}

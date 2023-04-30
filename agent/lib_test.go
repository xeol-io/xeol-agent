package agent

import (
	"context"
	"testing"
	"time"

	"github.com/noqcks/xeol-agent/agent/inventory"
	"github.com/noqcks/xeol-agent/internal/config"
)

func TestPeriodicallyGetInventoryReport(t *testing.T) {
	cfg := &config.Application{
		PollingIntervalSeconds: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Duration(cfg.PollingIntervalSeconds)*time.Second)
	defer cancel()

	reportTimes := []time.Time{}
	reportHandler := func(report inventory.Report, cfg *config.Application) error {
		reportTimes = append(reportTimes, time.Now())
		return nil
	}

	getInventoryReportFunc := func(cfg *config.Application) (inventory.Report, error) {
		return inventory.Report{}, nil
	}
	startTime := time.Now()
	PeriodicallyGetInventoryReport(ctx, cfg, getInventoryReportFunc, reportHandler)

	if len(reportTimes) < 2 {
		t.Fatalf("Expected at least two reports, but got %d", len(reportTimes))
	}

	firstInterval := reportTimes[0].Sub(startTime)
	if firstInterval >= time.Duration(cfg.PollingIntervalSeconds)*time.Second {
		t.Fatalf("Expected the first report to be immediate, but the interval was %v", firstInterval)
	}
}

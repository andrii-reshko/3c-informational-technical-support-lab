package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type DataPoint struct {
	Timestamp   int64
	Container   string
	CPUUsage    float64
	AssignedMem float64
}

var systemContainers = map[string]bool{
	"":             true,
	"prometheus":   true,
	"telegraf":     true,
	"docker-proxy": true,
}

func isSystemContainer(name string) bool {
	return systemContainers[name]
}

func getContainers(api v1.API, ctx context.Context) ([]string, error) {
	query := "docker_container_mem_limit"
	val, _, err := api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query containers: %w", err)
	}

	vector, ok := val.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", val)
	}

	seen := make(map[string]bool)
	var containers []string

	for _, sample := range vector {
		name := string(sample.Metric["container_name"])
		if isSystemContainer(name) || seen[name] {
			continue
		}
		seen[name] = true
		containers = append(containers, name)
	}

	return containers, nil
}

func main() {
	promURL := flag.String("prom", "http://localhost:9090", "Prometheus URL")
	cpuMetric := flag.String("cpu-metric", "docker_container_cpu_usage_percent", "CPU metric name")
	memMetric := flag.String("mem-metric", "docker_container_mem_limit", "Memory metric name")
	days := flag.Int("days", 30, "Number of days to fetch")
	stepStr := flag.String("step", "5m", "Prometheus step (e.g., 5m, 1h)")
	flag.Parse()

	stepDuration, err := time.ParseDuration(*stepStr)
	if err != nil {
		fmt.Printf("Error parsing step: %v\n", err)
		os.Exit(1)
	}

	client, err := api.NewClient(api.Config{
		Address: *promURL,
	})
	if err != nil {
		fmt.Printf("Error creating Prometheus client: %v\n", err)
		os.Exit(1)
	}

	apiClient := v1.NewAPI(client)
	ctx := context.Background()

	containers, err := getContainers(apiClient, ctx)
	if err != nil {
		fmt.Printf("Error getting containers: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d containers\n", len(containers))

	end := time.Now()
	start := end.AddDate(0, 0, -*days)

	queryRange := v1.Range{
		Start: start,
		End:   end,
		Step:  stepDuration,
	}

	var allPoints []DataPoint

	for _, container := range containers {
		fmt.Printf("Fetching: %s ...\n", container)

		cpuQuery := fmt.Sprintf(`%s{container_name="%s"}`, *cpuMetric, container)
		memQuery := fmt.Sprintf(`%s{container_name="%s"}`, *memMetric, container)

		cpuRes, _, err := apiClient.QueryRange(ctx, cpuQuery, queryRange)
		if err != nil {
			fmt.Printf("  Warning: CPU query failed: %v\n", err)
			continue
		}

		memRes, _, err := apiClient.QueryRange(ctx, memQuery, queryRange)
		if err != nil {
			fmt.Printf("  Warning: Memory query failed: %v\n", err)
			continue
		}

		dataMap := make(map[int64]*DataPoint)

		if matrix, ok := cpuRes.(model.Matrix); ok {
			for _, stream := range matrix {
				for _, sample := range stream.Values {
					ts := sample.Timestamp.Unix()
					dataMap[ts] = &DataPoint{
						Timestamp:   ts,
						Container:   container,
						CPUUsage:    float64(sample.Value),
						AssignedMem: 0,
					}
				}
			}
		}

		if matrix, ok := memRes.(model.Matrix); ok {
			for _, stream := range matrix {
				for _, sample := range stream.Values {
					ts := sample.Timestamp.Unix()
					if dp, exists := dataMap[ts]; exists {
						dp.AssignedMem = float64(sample.Value)
					}
				}
			}
		}

		for _, dp := range dataMap {
			allPoints = append(allPoints, *dp)
		}
	}

	if len(allPoints) == 0 {
		fmt.Println("No data found")
		os.Exit(1)
	}

	sort.Slice(allPoints, func(i, j int) bool {
		if allPoints[i].Timestamp != allPoints[j].Timestamp {
			return allPoints[i].Timestamp < allPoints[j].Timestamp
		}
		return allPoints[i].Container < allPoints[j].Container
	})

	firstTs := allPoints[0].Timestamp

	outputName := fmt.Sprintf("my_macbook_docker_cpu_mem_%s.csv", time.Now().Format("2006-01-02"))
	file, err := os.Create(outputName)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fmt.Fprintf(file, "timestamp,container,cpu_usage,assigned_mem\n")

	for _, dp := range allPoints {
		relativeTs := dp.Timestamp - firstTs
		safeName := strings.ReplaceAll(dp.Container, ",", "_")
		fmt.Fprintf(file, "%d,%s,%.8f,%.1f\n", relativeTs, safeName, dp.CPUUsage, dp.AssignedMem)
	}

	fmt.Printf("\nCSV written to %s (%d rows)\n", outputName, len(allPoints))
	fmt.Printf("Time range: %s to %s\n", start.Format(time.RFC3339), end.Format(time.RFC3339))
}

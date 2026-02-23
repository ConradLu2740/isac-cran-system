package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type BenchmarkConfig struct {
	URL         string
	Method      string
	Payload     interface{}
	Concurrency int
	Requests    int
	Timeout     time.Duration
}

type BenchmarkResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	QPS             float64
	LatencyP50      time.Duration
	LatencyP90      time.Duration
	LatencyP99      time.Duration
}

type LatencyRecorder struct {
	latencies []time.Duration
	mu        sync.Mutex
}

func NewLatencyRecorder(size int) *LatencyRecorder {
	return &LatencyRecorder{
		latencies: make([]time.Duration, 0, size),
	}
}

func (r *LatencyRecorder) Record(latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.latencies = append(r.latencies, latency)
}

func (r *LatencyRecorder) Percentile(p float64) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.latencies) == 0 {
		return 0
	}

	n := len(r.latencies)
	index := int(float64(n) * p / 100)
	if index >= n {
		index = n - 1
	}

	sorted := make([]time.Duration, n)
	copy(sorted, r.latencies)

	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[index]
}

func RunBenchmark(config BenchmarkConfig) (*BenchmarkResult, error) {
	var successCount, failedCount int64
	recorder := NewLatencyRecorder(config.Requests)
	startTime := time.Now()

	var wg sync.WaitGroup
	requestsPerWorker := config.Requests / config.Concurrency

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				latency, err := sendRequest(config)
				if err != nil {
					atomic.AddInt64(&failedCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
					recorder.Record(latency)
				}
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	result := &BenchmarkResult{
		TotalRequests:   int64(config.Requests),
		SuccessRequests: successCount,
		FailedRequests:  failedCount,
		TotalDuration:   totalDuration,
		QPS:             float64(successCount) / totalDuration.Seconds(),
		LatencyP50:      recorder.Percentile(50),
		LatencyP90:      recorder.Percentile(90),
		LatencyP99:      recorder.Percentile(99),
	}

	return result, nil
}

func sendRequest(config BenchmarkConfig) (time.Duration, error) {
	var body io.Reader
	if config.Payload != nil {
		jsonData, _ := json.Marshal(config.Payload)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(config.Method, config.URL, body)
	if err != nil {
		return 0, err
	}

	if config.Payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: config.Timeout}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	if resp.StatusCode >= 400 {
		return latency, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return latency, nil
}

func (r *BenchmarkResult) String() string {
	return fmt.Sprintf(`
=== Benchmark Result ===
Total Requests:   %d
Success Requests: %d
Failed Requests:  %d
Total Duration:   %v
QPS:              %.2f
Latency P50:      %v
Latency P90:      %v
Latency P99:      %v
========================`,
		r.TotalRequests,
		r.SuccessRequests,
		r.FailedRequests,
		r.TotalDuration,
		r.QPS,
		r.LatencyP50,
		r.LatencyP90,
		r.LatencyP99,
	)
}

func main() {
	baseURL := "http://localhost:8080"

	tests := []struct {
		name   string
		config BenchmarkConfig
	}{
		{
			name: "Health Check",
			config: BenchmarkConfig{
				URL:         baseURL + "/api/v1/health",
				Method:      "GET",
				Concurrency: 10,
				Requests:    1000,
				Timeout:     5 * time.Second,
			},
		},
		{
			name: "System Info",
			config: BenchmarkConfig{
				URL:         baseURL + "/api/v1/info",
				Method:      "GET",
				Concurrency: 10,
				Requests:    1000,
				Timeout:     5 * time.Second,
			},
		},
		{
			name: "IRS Status",
			config: BenchmarkConfig{
				URL:         baseURL + "/api/v1/irs/status",
				Method:      "GET",
				Concurrency: 10,
				Requests:    1000,
				Timeout:     5 * time.Second,
			},
		},
		{
			name: "Sensor List",
			config: BenchmarkConfig{
				URL:         baseURL + "/api/v1/sensor/list",
				Method:      "GET",
				Concurrency: 10,
				Requests:    1000,
				Timeout:     5 * time.Second,
			},
		},
		{
			name: "Runtime Metrics",
			config: BenchmarkConfig{
				URL:         baseURL + "/debug/metrics",
				Method:      "GET",
				Concurrency: 10,
				Requests:    1000,
				Timeout:     5 * time.Second,
			},
		},
	}

	fmt.Println("========================================")
	fmt.Println("  ISAC-CRAN System Performance Test")
	fmt.Println("========================================")
	fmt.Println()

	for _, test := range tests {
		fmt.Printf(">>> Running: %s\n", test.name)
		result, err := RunBenchmark(test.config)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Println(result.String())
	}

	fmt.Println("\n========================================")
	fmt.Println("  Performance Test Completed")
	fmt.Println("========================================")
}

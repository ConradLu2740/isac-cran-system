package e2e

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkHealthEndpoint(b *testing.B) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get("http://localhost:8080/api/v1/health")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			resp.Body.Close()
		}
	})
}

func BenchmarkInfoEndpoint(b *testing.B) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:8080/api/v1/info")
		if err != nil {
			b.Errorf("Request failed: %v", err)
			continue
		}
		resp.Body.Close()
	}
}

func TestConcurrentRequests(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	numRequests := 100
	numWorkers := 10
	requestsPerWorker := numRequests / numWorkers

	var successCount int64
	var failCount int64
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				resp, err := client.Get("http://localhost:8080/api/v1/health")
				if err != nil {
					atomic.AddInt64(&failCount, 1)
					continue
				}
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Concurrent test results:")
	t.Logf("  Total requests: %d", numRequests)
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Failed: %d", failCount)
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Requests/sec: %.2f", float64(numRequests)/elapsed.Seconds())

	if failCount > 0 {
		t.Errorf("Some requests failed: %d", failCount)
	}
}

func TestLatency(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	numRequests := 50
	latencies := make([]time.Duration, 0, numRequests)

	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := client.Get("http://localhost:8080/api/v1/health")
		latency := time.Since(start)

		if err != nil {
			t.Errorf("Request failed: %v", err)
			continue
		}
		resp.Body.Close()

		latencies = append(latencies, latency)
	}

	if len(latencies) == 0 {
		t.Fatal("No successful requests")
	}

	var total time.Duration
	var min, max time.Duration = latencies[0], latencies[0]
	for _, l := range latencies {
		total += l
		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
	}

	avg := total / time.Duration(len(latencies))

	t.Logf("Latency statistics:")
	t.Logf("  Min: %v", min)
	t.Logf("  Max: %v", max)
	t.Logf("  Avg: %v", avg)
	t.Logf("  Total requests: %d", len(latencies))

	if avg > 100*time.Millisecond {
		t.Errorf("Average latency too high: %v", avg)
	}
}

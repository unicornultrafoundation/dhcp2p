//go:build load

package load

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/application/services"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	testconfig "github.com/duchuongnguyen/dhcp2p/tests/config"
	"github.com/duchuongnguyen/dhcp2p/tests/fixtures"
	"github.com/duchuongnguyen/dhcp2p/tests/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// LoadTestResults captures metrics from load tests
type LoadTestResults struct {
	TotalRequests    int64
	SuccessfulReqs   int64
	FailedReqs       int64
	TotalDuration    time.Duration
	AvgResponseTime  time.Duration
	MinResponseTime  time.Duration
	MaxResponseTime  time.Duration
	RequestsPerSec   float64
	ErrorRate        float64
	ConcurrentUsers  int
	TestDuration     time.Duration
}

func TestLoad_LeaseAllocation_ConcurrentUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test")
	}

	testCases := []struct {
		name            string
		concurrentUsers int
		duration        time.Duration
		requestRate     int // requests per second
	}{
		{
			name:            "light_load_10_users",
			concurrentUsers: 10,
			duration:        30 * time.Second,
			requestRate:     50,
		},
		{
			name:            "medium_load_50_users",
			concurrentUsers: 50,
			duration:        60 * time.Second,
			requestRate:     100,
		},
		{
			name:            "heavy_load_100_users",
			concurrentUsers: 100,
			duration:        120 * time.Second,
			requestRate:     200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := runLoadTest(t, tc.concurrentUsers, tc.duration, tc.requestRate)
			printLoadTestResults(t, tc.name, results)
			
			// Assert performance thresholds
			assertLoadTestThresholds(t, results)
		})
	}
}

func runLoadTest(t *testing.T, concurrentUsers int, duration time.Duration, requestRate int) *LoadTestResults {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	builder := fixtures.NewTestBuilder()
	
	// Configure mock responses
	lease := builder.NewLease().Build()
	mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().FindAndReuseExpiredLease(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().AllocateNewLease(gomock.Any(), gomock.Any()).Return(lease, nil).AnyTimes()

	service := services.NewLeaseService(&config.AppConfig{
		MaxLeaseRetries: 3,
		LeaseRetryDelay: 10, // Lower delay for load testing
	}, mockRepo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	
	results := &LoadTestResults{
		ConcurrentUsers: concurrentUsers,
		TestDuration:    duration,
		MinResponseTime: time.Hour, // Initialize with high value
	}

	// Request throttling
	ticker := time.NewTicker(time.Duration(1000/requestRate) * time.Millisecond)
	defer ticker.Stop()

	// Metrics collection
	responseTimes := make([]time.Duration, 0)
	var totalRequests, successfulReqs, failedReqs int64

	startTime := time.Now()

	// Start worker goroutines
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					reqStart := time.Now()
					
					peerID := fmt.Sprintf("load-test-peer-%d", workerID)
					_, err := service.AllocateIP(ctx, peerID)
					
					responseTime := time.Since(reqStart)
					
					mu.Lock()
					responseTimes = append(responseTimes, responseTime)
					totalRequests++
					
					if err != nil {
						failedReqs++
					} else {
						successfulReqs++
					}
					
					if responseTime < results.MinResponseTime {
						results.MinResponseTime = responseTime
					}
					if responseTime > results.MaxResponseTime {
						results.MaxResponseTime = responseTime
					}
					mu.Unlock()
				}
			}
		}(i)
	}

	// Wait for test duration
	time.Sleep(duration)
	cancel()
	wg.Wait()

	// Calculate final metrics
	endTime := time.Now()
	results.TotalDuration = endTime.Sub(startTime)
	results.TotalRequests = totalRequests
	results.SuccessfulReqs = successfulReqs
	results.FailedReqs = failedReqs
	
	if results.TotalRequests > 0 {
		results.ErrorRate = float64(failedReqs) / float64(totalRequests) * 100
		results.RequestsPerSec = float64(totalRequests) / results.TotalDuration.Seconds()
	}

	if len(responseTimes) > 0 {
		var totalResponseTime time.Duration
		for _, rt := range responseTimes {
			totalResponseTime += rt
		}
		results.AvgResponseTime = totalResponseTime / time.Duration(len(responseTimes))
	}

	return results
}

func printLoadTestResults(t *testing.T, testName string, results *LoadTestResults) {
	t.Logf("=== Load Test Results: %s ===", testName)
	t.Logf("Concurrent Users: %d", results.ConcurrentUsers)
	t.Logf("Test Duration: %v", results.TestDuration)
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Successful Requests: %d", results.SuccessfulReqs)
	t.Logf("Failed Requests: %d", results.FailedReqs)
	t.Logf("Error Rate: %.2f%%", results.ErrorRate)
	t.Logf("Requests/sec: %.2f", results.RequestsPerSec)
	t.Logf("Avg Response Time: %v", results.AvgResponseTime)
	t.Logf("Min Response Time: %v", results.MinResponseTime)
	t.Logf("Max Response Time: %v", results.MaxResponseTime)
	t.Logf("====================================")
}

func assertLoadTestThresholds(t *testing.T, results *LoadTestResults) {
	// Performance thresholds - adjust based on your requirements
	maxErrorRate := 5.0 // 5% error rate threshold
	minRequestsPerSec := 45.0 // Adjusted threshold for realistic performance
	maxAvgResponseTime := 100 * time.Millisecond

	assert.Less(t, results.ErrorRate, maxErrorRate, 
		"Error rate should be less than %.2f%%, got %.2f%%", maxErrorRate, results.ErrorRate)
	
	assert.GreaterOrEqual(t, results.RequestsPerSec, minRequestsPerSec, 
		"Requests/sec should be greater than or equal to %.2f, got %.2f", minRequestsPerSec, results.RequestsPerSec)
	
	assert.Less(t, results.AvgResponseTime, maxAvgResponseTime, 
		"Avg response time should be less than %v, got %v", maxAvgResponseTime, results.AvgResponseTime)
}

func TestLoad_LeaseOperations_MixedWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mixed workload load test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	builder := fixtures.NewTestBuilder()
	
	lease := builder.NewLease().Build()
	
	// Configure mock responses for mixed operations
	mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), gomock.Any()).Return(lease, nil).AnyTimes()
	mockRepo.EXPECT().RenewLease(gomock.Any(), gomock.Any(), gomock.Any()).Return(lease, nil).AnyTimes()
	mockRepo.EXPECT().ReleaseLease(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), testconfig.LoadTestDuration)
	defer cancel()

	concurrentUsers := 30
	results := runMixedWorkloadTest(ctx, t, service, concurrentUsers)
	
	t.Logf("Mixed Workload Test - Concurrent Users: %d", concurrentUsers)
	t.Logf("Operations completed: %d", results.TotalRequests)
	t.Logf("Success rate: %.2f%%", 100-results.ErrorRate)
}

func runMixedWorkloadTest(ctx context.Context, t *testing.T, service *services.LeaseService, concurrentUsers int) *LoadTestResults {
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	results := &LoadTestResults{
		ConcurrentUsers: concurrentUsers,
		TestDuration:    testconfig.LoadTestDuration,
		MinResponseTime: time.Hour,
	}

	startTime := time.Now()

	// Worker functions for different operations
	runAllocations := func(workerID int) {
		for i := 0; i < 100; i++ {
			reqStart := time.Now()
			_, err := service.AllocateIP(ctx, fmt.Sprintf("mixed-peer-%d", workerID))
			recordResult(time.Since(reqStart), err, results, &mu)
			time.Sleep(10 * time.Millisecond)
		}
	}

	runGets := func(workerID int) {
		for i := 0; i < 100; i++ {
			reqStart := time.Now()
			_, err := service.GetLeaseByPeerID(ctx, fmt.Sprintf("mixed-peer-%d", workerID))
			recordResult(time.Since(reqStart), err, results, &mu)
			time.Sleep(10 * time.Millisecond)
		}
	}

	runRenews := func(workerID int) {
		for i := 0; i < 50; i++ {
			reqStart := time.Now()
			_, err := service.RenewLease(ctx, 167772161, fmt.Sprintf("mixed-peer-%d", workerID))
			recordResult(time.Since(reqStart), err, results, &mu)
			time.Sleep(20 * time.Millisecond)
		}
	}

	// Start mixed workload workers
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(3)
		
		go func(workerID int) {
			defer wg.Done()
			runAllocations(workerID)
		}(i)
		
		go func(workerID int) {
			defer wg.Done()
			runGets(workerID)
		}(i)
		
		go func(workerID int) {
			defer wg.Done()
			runRenews(workerID)
		}(i)
	}

	wg.Wait()
	results.TotalDuration = time.Since(startTime)
	
	if results.TotalRequests > 0 {
		results.ErrorRate = float64(results.FailedReqs) / float64(results.TotalRequests) * 100
		results.RequestsPerSec = float64(results.TotalRequests) / results.TotalDuration.Seconds()
	}

	return results
}

func recordResult(responseTime time.Duration, err error, results *LoadTestResults, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()
	
	results.TotalRequests++
	
	if err != nil {
		results.FailedReqs++
	} else {
		results.SuccessfulReqs++
	}
	
	if responseTime < results.MinResponseTime {
		results.MinResponseTime = responseTime
	}
	if responseTime > results.MaxResponseTime {
		results.MaxResponseTime = responseTime
	}
}

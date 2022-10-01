package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/square/go-jose.v2/json"
)

// SimulatedLatencies ..
type SimulatedLatencies map[string]int64

// AppendSimulateLatencyRoute Add a route to allow client to simulate simulatedLatencies
func AppendSimulateLatencyRoute(ctx context.Context, router *http.ServeMux) {
	router.Handle("/simulateDelay", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				body, _ := io.ReadAll(r.Body)
				var request SimulatedLatencies
				_ = json.Unmarshal(body, &request)
				fmt.Printf("Set latencies data %v\n", request)
				GetStore(ctx).SaveSimulatedLatencies(request)
			}
		},
	))
}

// SimulateLatencies ..
func SimulateLatencies(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delay := GetStore(ctx).GetSimulatedLatencies(r)
		if delay > 0 {
			fmt.Printf("Has latency %d ms appied\n", delay)
			time.Sleep(time.Duration(delay * int64(time.Millisecond)))
		}
	})
}

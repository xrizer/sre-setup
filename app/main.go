package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by path and status code.",
	}, []string{"path", "status"})

	duration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"path"})

	inFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Requests currently being served.",
	})

	// Chaos settings, changed at runtime via /admin/chaos.
	errorRateBits  atomic.Uint64 // float64 stored as bits
	extraLatencyMs atomic.Int64

	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// instrument records metrics and a structured access log for every request.
func instrument(path string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		inFlight.Inc()
		defer inFlight.Dec()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		h(rec, req)

		elapsed := time.Since(start)
		duration.WithLabelValues(path).Observe(elapsed.Seconds())
		requests.WithLabelValues(path, strconv.Itoa(rec.status)).Inc()

		level := slog.LevelInfo
		if rec.status >= 500 {
			level = slog.LevelError
		}
		logger.Log(req.Context(), level, "request completed",
			"path", path, "method", req.Method, "status", rec.status,
			"duration_ms", elapsed.Milliseconds())
	}
}

// book simulates booking a clear-aligner consultation.
func book(w http.ResponseWriter, _ *http.Request) {
	delay := 20 + rand.Intn(60) + int(extraLatencyMs.Load())
	time.Sleep(time.Duration(delay) * time.Millisecond)

	if rand.Float64() < math.Float64frombits(errorRateBits.Load()) {
		logger.Error("booking failed",
			"component", "clinic-scheduler",
			"error", "clinic-schedule-db: connection pool exhausted")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"booking_id": fmt.Sprintf("BK-%06d", rand.Intn(1_000_000)),
		"status":     "confirmed",
	})
}

// slow always responds after ~800 ms, for demoing latency without the chaos switch.
func slow(w http.ResponseWriter, _ *http.Request) {
	time.Sleep(time.Duration(700+rand.Intn(200)) * time.Millisecond)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// fail always returns 500, for demoing errors without the chaos switch.
func fail(w http.ResponseWriter, _ *http.Request) {
	logger.Error("request failed",
		"component", "payment-gateway",
		"error", "upstream timeout calling payment provider")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error":"internal error"}`))
}

// chaos lets the demo inject failures: /admin/chaos?error_rate=0.3&latency_ms=800
func chaos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v := q.Get("error_rate"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 || f > 1 {
			http.Error(w, "error_rate must be between 0 and 1", http.StatusBadRequest)
			return
		}
		errorRateBits.Store(math.Float64bits(f))
	}
	if v := q.Get("latency_ms"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			http.Error(w, "latency_ms must be >= 0", http.StatusBadRequest)
			return
		}
		extraLatencyMs.Store(int64(n))
	}

	rate := math.Float64frombits(errorRateBits.Load())
	lat := extraLatencyMs.Load()
	logger.Warn("chaos settings changed", "error_rate", rate, "latency_ms", lat)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"error_rate": rate, "latency_ms": lat})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/book", instrument("/book", book))
	mux.HandleFunc("/slow", instrument("/slow", slow))
	mux.HandleFunc("/error", instrument("/error", fail))
	mux.HandleFunc("/admin/chaos", chaos)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	mux.Handle("/metrics", promhttp.Handler())

	logger.Info("booking-api starting", "port", 8080)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Error("server stopped", "error", err.Error())
		os.Exit(1)
	}
}

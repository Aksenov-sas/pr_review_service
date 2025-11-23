package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
)

func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := &responseWriter{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(ww, r)

		// Получаем маршрут из chi (если доступен)
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		GlobalMetrics.HTTPRequestCount.With(
			prometheus.Labels{
				"method":   r.Method,
				"endpoint": routePattern,
				"status":   strconv.Itoa(ww.statusCode),
			},
		).Inc()

		duration := time.Since(start).Seconds()
		GlobalMetrics.HTTPResponseTime.With(
			prometheus.Labels{
				"method":   r.Method,
				"endpoint": routePattern,
			},
		).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

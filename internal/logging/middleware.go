package logging

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// HTTPMiddleware provides logging middleware for HTTP requests
type HTTPMiddleware struct {
	logger *Logger
}

// NewHTTPMiddleware creates a new HTTP logging middleware
func NewHTTPMiddleware(logger *Logger) *HTTPMiddleware {
	return &HTTPMiddleware{
		logger: logger,
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
// and supports all standard interfaces like Hijacker, Flusher, Pusher, etc.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Hijacker interface support for WebSocket connections
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("responseWriter does not implement http.Hijacker")
}

// Flusher interface support
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Pusher interface support (HTTP/2 Server Push)
func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return fmt.Errorf("responseWriter does not implement http.Pusher")
}

// LogRequest is middleware that logs HTTP requests
func (m *HTTPMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Create wrapped response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Create request-scoped logger
		reqLogger := m.logger.
			WithRequestID(requestID).
			WithOperation("http_request")

		// Log request start
		reqLogger.Info("Request started: %s %s", r.Method, r.URL.Path)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion with timing and status
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, wrapped.statusCode, duration, requestID)
	})
}

// LogRequestWithComponent is middleware that logs HTTP requests with a specific component name
func (m *HTTPMiddleware) LogRequestWithComponent(component string) func(http.Handler) http.Handler {
	componentLogger := m.logger.ForComponent(component)
	componentMiddleware := &HTTPMiddleware{logger: componentLogger}
	return componentMiddleware.LogRequest
}

// Simple convenience function to create logging middleware
func CreateHTTPLoggingMiddleware(component string) func(http.Handler) http.Handler {
	logger := GetLogger().ForComponent(component)
	middleware := NewHTTPMiddleware(logger)
	return middleware.LogRequest
}

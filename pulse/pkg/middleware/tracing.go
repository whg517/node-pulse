package middleware

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

// TraceIDHeader is the HTTP response header that carries the active trace ID.
// Front-end clients and log-correlation tools can use this to link a UI interaction
// to its backend trace in Jaeger / Grafana Tempo.
const TraceIDHeader = "X-Trace-Id"

// OtelGinMiddleware returns the otelgin middleware configured for the given service name.
// It creates an OpenTelemetry span for every incoming HTTP request and attaches it to
// the request context so that downstream code can use otel.Tracer(...).Start(ctx, ...).
func OtelGinMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// TraceIDMiddleware injects the active OpenTelemetry trace ID into:
//
//  1. The HTTP response as the "X-Trace-Id" header so callers can correlate
//     browser network requests with distributed traces.
//  2. The standard library logger prefix so that trace IDs appear in server-side
//     log lines, enabling log-to-trace correlation without a full structured logger.
//
// This middleware must be placed AFTER the otelgin tracing middleware so that the
// span is already attached to the request context when TraceIDMiddleware runs.
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		spanCtx := trace.SpanFromContext(c.Request.Context()).SpanContext()

		if spanCtx.IsValid() {
			traceID := spanCtx.TraceID().String()
			spanID := spanCtx.SpanID().String()

			// Inject into response header (for client-side correlation)
			c.Header(TraceIDHeader, traceID)

			// Annotate the Go standard logger for the duration of this request so that
			// any log.Printf calls include the trace context.  The prefix is cleared
			// automatically when the handler returns because we restore the previous one.
			prev := log.Prefix()
			log.SetPrefix(fmt.Sprintf("[trace=%s span=%s] ", traceID, spanID))
			defer log.SetPrefix(prev)
		}

		c.Next()
	}
}

package filters

import (
	"context"
	"net/http"

	"github.com/lucheng0127/bumblebee/pkg/utils/host"
	"github.com/lucheng0127/bumblebee/pkg/utils/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func WithTrace(next http.Handler, enable, master bool, tracer trace.TracerProvider, zone string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operationID := r.Header.Get(HeaderOperationID)
		if !enable {
			next.ServeHTTP(w, r)
			return
		}

		if master {
			// Start span
			ctx, span := tracer.Tracer(runtime.TraceServiceName).Start(r.Context(), operationID)
			defer span.End()

			span.SetAttributes(
				attribute.String("url", r.URL.Path),
				attribute.String("zone", zone),
				attribute.String("host", host.GetHostname()),
			)

			// Set span context to http.Request.Header so slave can get span context from it
			propagator := propagation.TraceContext{}
			propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))
		} else {
			// Extract span context from http header
			propagator := propagation.TraceContext{}
			ctx := context.Background()
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

			// Start span from context
			_, span := tracer.Tracer(runtime.TraceServiceName).Start(ctx, operationID)
			defer span.End()

			span.SetAttributes(
				attribute.String("url", r.URL.Path),
				attribute.String("zone", zone),
				attribute.String("host", host.GetHostname()),
			)
		}

		next.ServeHTTP(w, r)
	})
}

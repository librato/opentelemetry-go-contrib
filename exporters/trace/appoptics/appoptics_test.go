package appoptics_test

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/exporters/trace/appoptics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func AOSpanTest() {
	exporter, _ := appoptics.NewExporter()
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp))
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("ex.com/basic")
	ctx := baggage.ContextWithValues(context.Background(), attribute.Int("foo", 1))

	func(ctx context.Context) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "operation")
		defer span.End()

		span.AddEvent("Nice operation!", trace.WithAttributes(attribute.Int("bogons", 100)))

		func(ctx context.Context) {
			var span trace.Span
			ctx, span = tracer.Start(ctx, "Sub operation...")
			defer span.End()

			span.AddEvent("Sub span event")
		}(ctx)
	}(ctx)

	time.Sleep(10*time.Second)
}



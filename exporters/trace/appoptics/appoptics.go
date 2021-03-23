package appoptics

import (
	"context"

	export "go.opentelemetry.io/otel/sdk/export/trace"
)

type Exporter struct {

}

var (
	_ export.SpanExporter = &Exporter{}
)

func NewExporter() (*Exporter, error) {
	return &Exporter{}, nil
}

func (e *Exporter) 	ExportSpans(ctx context.Context, ss []*export.SpanSnapshot) error {
	return nil
}

func (e *Exporter) 	Shutdown(ctx context.Context) error {
	return nil
}

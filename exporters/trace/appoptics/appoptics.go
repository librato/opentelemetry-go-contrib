package appoptics

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/appoptics/appoptics-apm-go/v1/ao"
	"go.opentelemetry.io/otel/attribute"
	export "go.opentelemetry.io/otel/sdk/export/trace"
)

const (
	xtraceVersionHeader = "2B"
	sampledFlags        = "01"
)

var wsKeyMap = map[string]string{
	"http.method":      "HTTPMethod",
	"http.url":         "URL",
	"http.status_code": "Status",
}
var queryKeyMap = map[string]string{
	"db.connection_string": "RemoteHost",
	"db.name":              "Database",
	"db.statement":         "Query",
	"db.system":            "Flavor",
}

type Exporter struct {
	cfg Config
}

type Config struct {
	ServiceKey string
}

var (
	_ export.SpanExporter = &Exporter{}
)

func NewExporter(c Config) (*Exporter, error) {
	e := &Exporter{cfg: c}
	ao.SetServiceKey(e.cfg.ServiceKey)
	return e, nil
}

func (e *Exporter) 	ExportSpans(ctx context.Context, ss []*export.SpanSnapshot) error {
	for _, span := range ss {
		xTraceID := getXTraceID(span.SpanContext.TraceID().String(), span.SpanContext.SpanID().String())

		startOverrides := ao.Overrides{
			ExplicitTS:    span.StartTime,
			ExplicitMdStr: xTraceID,
		}
		endOverrides := ao.Overrides{
			ExplicitTS: span.EndTime,
		}
		var traceContext context.Context
		kvs := extractKvs(span)

		if !span.ParentSpanID.IsValid() {
			trace := ao.NewTraceWithOverrides(span.Name, startOverrides, nil)
			trace.SetTransactionName(span.Name)
			traceContext = ao.NewContext(context.Background(), trace)
			trace.SetStartTime(span.StartTime) // This is for histogram only
			trace.EndWithOverrides(endOverrides, kvs...)
		} else {
			parentXTraceID := getXTraceID(span.SpanContext.TraceID().String(), span.ParentSpanID.String())
			traceContext = ao.FromXTraceIDContext(context.Background(), parentXTraceID)
			aoSpan, _ := ao.BeginSpanWithOverrides(traceContext, span.Name, ao.SpanOptions{}, startOverrides)
			aoSpan.EndWithOverrides(endOverrides, kvs...)
		}
	}
	return nil
}

func (e *Exporter) 	Shutdown(ctx context.Context) error {
	return nil
}

func getXTraceID(traceID string, spanID string) string {
	taskId := strings.ToUpper(strings.ReplaceAll(fmt.Sprintf("%0-40v", traceID), " ", "0"))
	opId := strings.ToUpper(strings.ReplaceAll(fmt.Sprintf("%0-16v", spanID), " ", "0"))
	return xtraceVersionHeader + taskId + opId + sampledFlags
}

func extractKvs(span *export.SpanSnapshot) []interface{} {
	var kvs []interface{}
	for _, kv := range span.Attributes {
		k := kv.Key
		v := kv.Value
		kvs = append(kvs, k, fromAttributeValue(v))
	}
	if len(span.ParentSpanID.String()) == 0 {
		kvs = append(kvs, extractWebserverKvs(span)...)
	}
	kvs = append(kvs, extractQueryKvs(span)...)

	return kvs
}

func fromAttributeValue(v attribute.Value) interface{} {
	switch v.Type() {
	case attribute.INT64:
		return v.AsInt64()
	case attribute.FLOAT64:
		return v.AsFloat64()
	case attribute.BOOL:
		return v.AsBool()
	case attribute.STRING:
		return v.AsString()
	case attribute.ARRAY:
		panic("array is not yet supported.")
	default:
		return nil
	}
}

func extractWebserverKvs(span *export.SpanSnapshot) []interface{} {
	return extractSpecKvs(span, wsKeyMap, "ws")
}

func extractQueryKvs(span *export.SpanSnapshot) []interface{} {
	return extractSpecKvs(span, queryKeyMap, "query")
}

func extractSpecKvs(span *export.SpanSnapshot, lookup map[string]string, specValue string) []interface{} {
	attrMap := make(map[string]attribute.Value)
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value
	}
	var result []interface{}
	for otKey, aoKey := range lookup {
		if val, ok := attrMap[otKey]; ok {
			result = append(result, aoKey)
			result = append(result, fromAttributeValue(val))
		}
	}
	if len(result) > 0 {
		result = append(result, "Spec")
		result = append(result, specValue)
	}
	return result
}
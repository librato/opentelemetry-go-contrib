module github.com/librato/opentelemetry-go-contrib/exporters/trace/appoptics

go 1.15

require (
	github.com/appoptics/appoptics-apm-go v1.14.0
	go.opentelemetry.io/otel/sdk v0.19.0
	github.com/appoptics/appoptics-apm-go v1.2.3
)

replace github.com/appoptics/appoptics-apm-go => ../../../../../appoptics/appoptics-apm-go

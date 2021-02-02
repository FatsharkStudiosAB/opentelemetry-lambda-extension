module github.com/open-telemetry/opentelemetry-lambda-extension

go 1.14

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticexporter v0.18.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	go.opentelemetry.io/collector v0.18.0
	go.uber.org/zap v1.16.0
)

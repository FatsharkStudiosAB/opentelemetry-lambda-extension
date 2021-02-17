package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/open-telemetry/opentelemetry-lambda-extension/extension"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticexporter"
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	printPrefix     = fmt.Sprintf("[%s]", extensionName)
)

func components() (component.Factories, error) {
	var errs []error

	extensions, err := component.MakeExtensionFactoryMap()
	if err != nil {
		errs = append(errs, err)
	}

	receivers, err := component.MakeReceiverFactoryMap(
		otlpreceiver.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	exporters, err := component.MakeExporterFactoryMap(
		loggingexporter.NewFactory(),
		elasticexporter.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	processors, err := component.MakeProcessorFactoryMap(
		batchprocessor.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	factories := component.Factories{
		Extensions: extensions,
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
	}

	return factories, componenterror.CombineErrors(errs)
}

func main() {
	factories, _ := components()
	collector := NewInProcessCollector(factories)
	collector.prepareConfig()
	collector.start()

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(printPrefix, "Received", s)
		println(printPrefix, "Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	println(printPrefix, "Register response:", prettyPrint(res))

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx, collector)
}

func processEvents(ctx context.Context, collector *InProcessCollector) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			println(printPrefix, "Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				println(printPrefix, "Error:", err)
				println(printPrefix, "Exiting")
				return
			}
			println(printPrefix, "Received event:", prettyPrint(res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				collector.stop() // TODO: handle return values
				println(printPrefix, "Received SHUTDOWN event")
				println(printPrefix, "Exiting")
				return
			}
		}
	}
}

func prettyPrint(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(data)
}

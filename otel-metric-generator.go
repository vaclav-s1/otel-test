package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	go_otlp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func main() {
	ctx := context.Background()

	// Set up OTLP exporter to local collector
	exp, err := go_otlp.New(ctx,
		go_otlp.WithInsecure(),
		go_otlp.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		fmt.Println("failed to create exporter:", err)
		os.Exit(1)
	}
	defer exp.Shutdown(ctx)
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attribute.String("service.name", "otel-metric-generator"),
		),
	)

	reader := sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(2*time.Second))
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	meter := provider.Meter("test-meter")

	// Emit a dummy metric where the value to be summed is in an attribute
	valueMetric, err := meter.Int64Counter(
		"filesScanCount",
		metric.WithDescription("filesScanCount metric for sumconnector attribute aggregation"),
	)
	if err != nil {
		fmt.Println("failed to create counter:", err)
		os.Exit(1)
	}

	bucketIds := []string{"bucketA", "bucketB"}
	accountIds := []string{"account1", "account2"}

	fmt.Println("Metric generation started. Press Ctrl+C to stop.")

	for {
		// Simulate scanning files
		for idx, bucketId := range bucketIds {
			count := rand.Int63n(100)
			valueMetric.Add(ctx, 1, metric.WithAttributes(
				attribute.String("bucketId", bucketId),
				attribute.String("accountId", accountIds[idx]),
				attribute.Int64("filesScanCount", count), // value as attribute
			))
			fmt.Printf("Emitted filesScanCount with filesScanCount=%d for bucketId=%s, accountId=%s\n", count, bucketId, accountIds[idx])
		}

		// Sleep for a while before the next iteration
		time.Sleep(5 * time.Second)
	}

	select {}

}

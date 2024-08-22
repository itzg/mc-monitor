package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.GetMeterProvider().Meter("minecraft")
)

func NewInt64ObservableGauge(
	name string,
	description string,
	callback func() int64,
	attributes []attribute.KeyValue,
) {
	_, err := meter.Int64ObservableGauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit("1"),
		metric.WithInt64Callback(
			func(ctx context.Context, observer metric.Int64Observer) error {
				observer.Observe(callback(), metric.WithAttributes(attributes...))
				return nil
			},
		),
	)
	handleError(fmt.Sprintf("Error creating %s metric", name), err)
	return
}

func NewFloat64ObservableGauge(
	name string,
	description string,
	callback func() float64,
	attributes []attribute.KeyValue,
) {
	_, err := meter.Float64ObservableGauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit("ms"),
		metric.WithFloat64Callback(
			func(ctx context.Context, observer metric.Float64Observer) error {
				observer.Observe(callback(), metric.WithAttributes(attributes...))
				return nil
			},
		),
	)
	handleError(fmt.Sprintf("Error creating %s metric", name), err)
}

func handleError(msg string, err error) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}

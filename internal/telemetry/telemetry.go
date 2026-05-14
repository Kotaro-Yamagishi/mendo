package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Init は OpenTelemetry のトレーサーを初期化する。
// 戻り値の関数を defer で呼ぶことで、shutdown 時にバッファをフラッシュする。
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	// OTLP HTTP exporter（Jaeger の OTLP 受信ポートに送信）
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(), // ローカルなので TLS 不要
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// リソース（サービス名等のメタデータ）
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// グローバルに設定
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

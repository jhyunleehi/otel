package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// Jaeger Exporter 생성: Tempo와 연결
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://192.168.137.30:14268/api/traces")))
	if err != nil {
		return nil, err
	}

	// Tracer Provider 생성
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp), // 데이터를 배치 단위로 전송
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("example-service"),
		)),
	)

	// 전역적으로 Tracer Provider 설정
	otel.SetTracerProvider(tp)

	return tp, nil
}

func main() {
	// Tracer 초기화
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		// 종료 시 트레이스 데이터 플러시
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

	tracer := otel.Tracer("example-service")

	// 부모 Span 시작 (최상위 트레이스)
	ctx, span := tracer.Start(context.Background(), "root-operation")
	defer span.End()

	// 하위 호출
	callAnotherService(ctx, tracer)
}

func callAnotherService(ctx context.Context, tracer trace.Tracer) {
	// 하위 Span 생성
	ctx, span := tracer.Start(ctx, "call-another-service")
	defer span.End()

	// 다른 서비스 호출을 시뮬레이션
	simulateServiceCall(ctx, tracer)
}

func simulateServiceCall(ctx context.Context, tracer trace.Tracer) {
	// 하위 Span 생성
	ctx, span := tracer.Start(ctx, "simulateServiceCall")
	defer span.End()

	// 다른 서비스 호출을 시뮬레이션
	rcall(ctx, tracer, 9)
}

func rcall(ctx context.Context, tracer trace.Tracer, i int) {
	ctx, span := tracer.Start(ctx, "rcall")
	defer span.End()
	if i == 0 {
		return
	}
	rcall(ctx, tracer, i)
}

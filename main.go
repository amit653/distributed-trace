package main

/*
go get go.opentelemetry.io/otel #Provides the trace.Tracer, context propagation, etc
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc #Sends spans to Jaeger’s OTLP endpoint (localhost:4317).
go get go.opentelemetry.io/otel/sdk/trace       #Implements the TracerProvider and span processors.
go get github.com/jackc/pgx/v5/pgxpool
*/
import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func main() {
	ctx := context.Background()
	//set up exporter and trace provider
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-client"),
		)),
	)
	otel.SetTracerProvider(tp)
	defer func() {
		if err := tp.Shutdown(ctx); err != nil { // spans are batched and not exported before program exits.shut down the tracer provider at the end
			log.Fatal(err)
		}
	}()
	trc := otel.Tracer("go-client")

	conn, err := pgxpool.New(ctx, "postgres://postgres:postgres@192.xx:5432/std")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Root span
	ctx, span := trc.Start(ctx, "go-pg-request")
	defer span.End()
	// Child span for query
	time.Sleep(10 * time.Second)
	var total int
	//use the updated ctx from parent go-pg-request
	ctx, dbSpan := trc.Start(ctx, "postgres_query")
	row := conn.QueryRow(ctx, "select  count(*) from pg_class p1,pg_class p2,pg_class p3")
	if err = row.Scan(&total); err != nil {
		//log.Fatal(err)
		dbSpan.RecordError(err)
	}
	dbSpan.SetAttributes(attribute.Int("db.result.count", total))
	dbSpan.End()
}

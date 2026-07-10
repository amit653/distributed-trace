Workflow:
1) Instrument Python/GO code with OpenTelemtry SDK
2) Otlpspanexporter sends them over gRPC to collector endpoint (localhost:4317 (OTLP gRPC) → Jaeger)
3) Jaeger Collector receives spans from exporter forward them to storage , in this case its in memory or use Elasticsearch 
  using COLLECTOR_OTLP_ENABLED=true, Jaeger listens on 4317 (gRPC) and 4318 (HTTP)
4) Open Jaeger UI http://localhost:16686;  query service fetch traces 

docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \   # Jaeger UI
  -p 4317:4317 \     # OTLP gRPC
  -p 4318:4318 \     # OTLP HTTP
  jaegertracing/all-in-one:1.57
From local machine test python/go postgres tracing

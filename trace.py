'''
pip install opentelemetry-sdk opentelemetry-exporter-otlp opentelemetry-instrumentation-psycopg2
'''
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.psycopg2 import Psycopg2Instrumentor
import time
import psycopg2

# Setup tracer
trace.set_tracer_provider( TracerProvider(resource=Resource.create({"service.name": "python-client"})))
# 14250 is Jaeger’s native gRPC collector, not OTLP. #Port 4317 → OTLP gRPC collector
#  Add OTLP exporter (to Jaeger on port 4317)
exporter = OTLPSpanExporter(endpoint="http://localhost:4317", insecure=True)  #Jaeger native collector gRPC.
trace.get_tracer_provider().add_span_processor(BatchSpanProcessor(exporter))

#  Instrument psycopg2 BEFORE making any DB connections
Psycopg2Instrumentor().instrument()

tracer=trace.get_tracer(__name__) # Get tracer
# Connect to Postgres (via PgBouncer if you want pooling)
conn = psycopg2.connect("dbname=std user=postgres password=postgres host=192.xx")
cur = conn.cursor()

# Trace an API payload → DB query
with tracer.start_as_current_span("api_request") as span:
    time.sleep(10) # simulate network hop delay
    #span.set_attribute("api.payload.ride_id", 1)
    #cur.execute("SELECT * FROM rides WHERE ride_id=%s", (1,))
    #result = cur.fetchone()
    with tracer.start_as_current_span("postgres_request") as db_span:

        db_span.set_attribute("db.operation", "SELECT_ALL")
        db_span.set_attribute("db.table", "pg_class")
        cur.execute(" select  count(*) from pg_class p1,pg_class p2,pg_class p3 ")
        result = cur.fetchall()
        db_span.set_attribute("db.result.rowcount", cur.rowcount)
# otelwrapper
Simple wrapper for OTel tracer provider in golang 

## Installation

```bash
go get -u github.com/x-qdo/otelwrapper
```
also, most likely you need this as well
```bash
go get -u go.opentelemetry.io/otel/trace
```

## Initialization of TracerProvider
```go
// let's create a dummy context
ctx, cancelF := context.WithCancel(context.Background())
wg := new(sync.WaitGroup)

...

tp, err := tracer.InitTracerProvider("my_sexy_service", "default")
if err != nil {
    return nil, err
}

go tracer.ShutdownWaiting(tp, ctx, wg)

...

ctx, span := otel.Tracer("my_first_tracer").Start(ctx, "my_first_span")
defer span.End()

...

span.AddEvent("my fist event")
span.SetAttributes(attribute.Key("foo").String("boo"))
```
So, as you can see we initiate new global `tracesdk.TracerProvider` here and set up a listener to gracefully shut it down as soon as `cancelF()` will be called. `sync.WaitGroup` we need here just to wait until an application finalizes everything.

Under the hood `InitTracerProvider()` creates an instance of OTLP Exporter. Please pay your attention that you must set `OTEL_EXPORTER_OTLP_ENDPOINT` to as its environment variable.
For example,
```bash
OTEL_EXPORTER_OTLP_ENDPOINT="0.0.0.0:4317"
```

If you don't want to use OTLP as your span delivery, then you have to create an exporter by yourself. For example, you need Jaeger:
```go
//Create the Jaeger exporter
exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT"))))
if err != nil {
    return nil, err
}

tp, err := tracer.InitTracerProvider("my_sexy_service", "default", exp)
if err != nil {
    return nil, err
}
```
Btw, you can have multiple exporters:
```go
tracer.InitTracerProvider("my_sexy_service", "default", exp1, exp2, ...)
```
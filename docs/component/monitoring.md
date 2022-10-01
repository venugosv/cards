[Home](../../README.md)

# Monitoring

Collecting, processing, aggregating, and displaying real-time quantitative data about a system. Common sources of monitoring data are metrics, logs and distributed traces.

## Dashboards & reports

**Stats**

Currently we have one monitoring dashboard for each service in stackdriver:

- [cardcontrols dashboard](https://console.cloud.google.com/monitoring/dashboards/custom/13062528041815601817?project=anz-x-fabric-np-641432&timeDomain=6h)
- [cards dashboard](https://console.cloud.google.com/monitoring/dashboards/custom/2352205604382254424?project=anz-x-fabric-np-641432&timeDomain=6h)

We can create dashboard per environment later on when there is more data.

**Trace**

- [cardcontrols st](https://console.cloud.google.com/traces/list?project=anz-x-fabric-np-641432&tid=9bb30ee86319ecf9ca6762d4b7729123&pageState=(%22traceIntervalPicker%22:(%22groupValue%22:%22PT6H%22),%22traceFilter%22:(%22chips%22:%22%255B%257B_22k_22_3A_22LABEL_3Aapp_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22cardcontrols_5C_22_22%257D_2C%257B_22k_22_3A_22LABEL_3Aenv_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22staging_5C_22_22%257D%255D%22)))
- [cards st](https://console.cloud.google.com/traces/list?project=anz-x-fabric-np-641432&tid=9bb30ee86319ecf9ca6762d4b7729123&pageState=(%22traceIntervalPicker%22:(%22groupValue%22:%22PT6H%22),%22traceFilter%22:(%22chips%22:%22%255B%257B_22k_22_3A_22LABEL_3Aenv_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22staging_5C_22_22%257D_2C%257B_22k_22_3A_22LABEL_3Aapp_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22cards_5C_22_22%257D%255D%22)))

*Note: You can easily change the env filter to view traces in other ENVs(currently not available because we have not deployed to other environments yet)*

## Metrics & Distributed trace

### Definitions

- Metrics

> Metrics are any quantifiable piece of data that you would like to track, such as latency in a service or database, request content length, or number of open file descriptors.

- Distributed trace

> Distributed tracing, also called distributed request tracing, is a method used to profile and monitor applications, especially those built using a microservices architecture. Distributed tracing helps pinpoint where failures occur and what causes poor performance.

### Tools

[OpenCensus](https://opencensus.io/) is a set of libraries for various languages that allow us to collect application metrics and distributed traces. It is what we use in fabric. The data collected by OpenCensus can be transferred to a backend of our choice in realtime.

Here are the backend systems we use in different environments:
  - Local
    - [Prometheus](https://prometheus.io/) for metrics
    - [Jaeger](https://www.jaegertracing.io/) for distributed trace

  - Cloud
    - [Stackdriver Metrics](https://cloud.google.com/monitoring/docs) for metrics
    - [Stackdriver Trace](https://cloud.google.com/trace/docs) for distributed trace

### How to view metrics & distributed trace locally?

Simply run `make run-pencensus`, it will start docker containers for both prometheus and jaeger locally.
You can access those two services at:
- Prometheus: http://localhost:9090
- Jaeger:  http://localhost:16686

## Instrumentation

Note: this section is mainly "borrowed" from fabric-payments. There are 3 areas that tracing/metrics are implemented in fabric-cards:

### gRPC

**Server**

```go
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.ChainUnaryInterceptor(
			grpcValidator.UnaryServerInterceptor(),
			logware.UnaryServerLoggerInjectionInterceptor(logger),
			logware.UnaryServerResponseTimeLoggingInterceptor(metadata.ApplicationName, logware.UnaryBypassPathPrefixes(healthPrefix, reflectionPrefix)),
			middleware.UnaryServerGRPCOpenCensusInterceptor(logger, metadata.ApplicationName),
			logware.UnaryServerRequestLoggerInterceptor(logware.UnaryBypassPathPrefixes(healthPrefix, reflectionPrefix)),
		),
	)
```

Two lines of code did the magic:

`grpc.StatsHandler(&ocgrpc.ServerHandler{})`

`ocgrpc.ServerHandler` Handler recording OpenCensus stats and traces. Use with gRPC servers.

 - `middleware.UnaryServerGRPCOpenCensusInterceptor(logger, metadata.ApplicationName)`

The `UnaryServerGRPCOpenCensusInterceptor` starts the span for our server, adds the trace-id to the logger and then ends the span after the request has been handled.

**Client**

Simple implementation for the grpc client.

```go
conn, err := grpc.Dial(texURL.Host, grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{})
```

This enables tracing and metrics to be recorded by the grpc client.

### HTTP

**Client**

https://opencensus.io/guides/http/go/net_http/client/

[requestutils/request.go](../pkg/requestutils/request.go)
```go
// NewDefaultHTTPTransport returns a preconfigured default http transport
func NewDefaultHTTPTransport(transport http.RoundTripper) http.RoundTripper {
	return &preOCTransport{
		base: &ochttp.Transport{
			Base: &postOCTransport{
				base: transport,
			},
			Propagation: &b3.HTTPFormat{},
		},
	}
}
```

This HTTPTransport can be used with any client to enable OC Tracing and Metrics collection.

`&ochttp.Transport` Enables the base OC transport which handles most of the functionality. One drawback of the http client compared to the grpc client is it only collates metrics and tracing by method (POST, GET, ...). This isn't that useful as we may use one client for multiple different calls which would share the same method. Therefore we need to add a tag to the metrics and tracing as to the exact operation or call that the client is making. We do this thorugh `preOCTransport` and `postOCTransport`. These retrieve data from the request context and adds it to the trace and metrics attributes.

Then when we make the call to the http endpoint we simply can add a tag that specifies exactly what operation we are enacting:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
if err != nil {
    return nil, err
}
req = rest.TagClientOperation(req, "EXAMPLE_OPERATION")
resp, err := client.Do(req)
```

## Logging

**StackDriver logs**

Login to gcp and click the following links to access application logs:
- [Card Controls SIT](https://console.cloud.google.com/logs/viewer?project=anz-x-apps-np-e1bb39&customFacets=undefined&limitCustomFacetWidth=true&minLogLevel=0&expandAll=false&timestamp=2020-05-19T06:32:28.459000000Z&advancedFilter=resource.type%3D%22k8s_container%22%0Aresource.labels.cluster_name%3D%22anz-x-apps-np-gke%22%0Aresource.labels.namespace_name%3D%22fabric-services-sit%22%0Aresource.labels.container_name%3D%22cardcontrols%22%0A&dateRangeStart=2020-05-19T00:45:07.110Z&interval=PT6H&scrollTimestamp=2020-05-19T06:31:24.745057904Z&dateRangeEnd=2020-05-19T06:45:07.110Z)
- [Card SIT](https://console.cloud.google.com/logs/viewer?project=anz-x-apps-np-e1bb39&customFacets=undefined&limitCustomFacetWidth=true&minLogLevel=0&expandAll=false&timestamp=2020-05-19T06:32:28.459000000Z&advancedFilter=resource.type%3D%22k8s_container%22%0Aresource.labels.cluster_name%3D%22anz-x-apps-np-gke%22%0Aresource.labels.namespace_name%3D%22fabric-services-sit%22%0Aresource.labels.container_name%3D%22cards%22%0A&dateRangeStart=2020-05-19T00:44:00.693Z&interval=PT6H&scrollTimestamp=2020-05-19T06:34:17.476399384Z&dateRangeEnd=2020-05-19T06:44:00.693Z)

Once you opened the above links, you can easily change the env filter to see logs for other environments.

## Resources

- [OpenCensus](https://opencensus.io/)
- [Google SRE books](https://landing.google.com/sre/books/) - free books
- [Prometheus](https://prometheus.io/)
- [Jaeger](https://www.jaegertracing.io/)
- [Stackdriver Metrics](https://cloud.google.com/monitoring/docs)
- [Stackdriver Trace](https://cloud.google.com/trace/docs)

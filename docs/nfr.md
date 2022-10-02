# Non-Functional Requirement

> Capture the following components for Non-functional requirements of the service.

## Service Level Objective(SLO)

> Add SLO information

## Service Level Agreement(SLA)

> Add SLA information

## Availability

| Environment | Region | Status     | URL                                |
| ----------- | ------ | ---------- | ---------------------------------- |
| ST          |        | Delivered  | cards-st.fabric.gcpnp.anz:443      |
| SIT         |        | Delivered  | cards-sit.fabric.gcpnp.anz:443     |
| SIT-N         |        | Delivered  | cards-sit-n.fabric.gcpnp.anz:443     |
| INTPNV         |        | Delivered  | cards-intpnv.fabric.gcpnp.anz:443     |
| PNV         |        | Delivered  | cards-pnv.fabric.gcpnp.anz:443     |
| PreProd     |        | Inprogress | cards-preprod.fabric.gcpnp.anz:443 |
| Production  |        |            |                                 |

## Metrics

Metrics are any quantifiable piece of data that you would like to track, such as latency in a service or database,
request content length, or number of open file descriptors.

| Key Metric | Description |
| ---------- | ----------- |
| ..         | ..          | .. |
| ..         | ..          |

## Dashboards

Collecting, processing, aggregating, and displaying real-time quantitative data about a system. Common sources of
monitoring data are metrics, logs and distributed traces.

| Service              |       Type        | Cards                |
|----------------------|---------------------|----------------------|
| Grafana              | Dashboard           | [Link](https://grafana.fabric.gcpnp.anz/d/kpTGqFGGk/cards)                  |             |
| StackDriver          | Stats               | [Link](https://console.cloud.google.com/monitoring/dashboards/custom/2352205604382254424?project=anz-x-fabric-np-641432&amp;timeDomain=6h)                  |
| StackDriver          | Trace               | [Link](https://console.cloud.google.com/traces/list?project=anz-x-fabric-np-641432&amp;pageState=(%22traceIntervalPicker%22:(%22groupValue%22:%22P30D%22),%22traceFilter%22:(%22chips%22:%22%255B%257B_22k_22_3A_22LABEL_3Aapp_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22cards_5C_22_22%257D_2C%257B_22k_22_3A_22LABEL_3Aapp_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22cardcontrols_5C_22_22%257D%255D%22)))                  |
| StackDriver          | Logs                | [Link](https://console.cloud.google.com/logs/query;query=resource.type%3D%22k8s_container%22%0Aresource.labels.cluster_name%3D%22anz-x-apps-np-gke%22%0Aresource.labels.container_name%3D%22cards%22%20OR%20resource.labels.container_name%3D%22cardcontrols%22;timeRange=PT6H;summaryFields=undefined,jsonPayload%252Flevel,jsonPayload%252Ffields%252F%2522x-b3-traceid%2522:true:32:beginning?project=anz-x-apps-np-e1bb39)                  |

## Trace

Distributed tracing, also called distributed request tracing, is a method used to profile and monitor applications,
especially those built using a microservices architecture. Distributed tracing helps pinpoint where failures occur and
what causes poor performance. List of frequently used trace queries.

Trace can be found
at [Link](https://console.cloud.google.com/logs/query;query=resource.type%3D%22k8s_container%22%0Aresource.labels.cluster_name%3D%22anz-x-apps-np-gke%22%0Aresource.labels.container_name%3D%22cards%22%20OR%20resource.labels.container_name%3D%22cardcontrols%22;timeRange=PT6H;summaryFields=undefined,jsonPayload%252Flevel,jsonPayload%252Ffields%252F%2522x-b3-traceid%2522:true:32:beginning?project=anz-x-apps-np-e1bb39)

## Synthetic Checks

> List of endpoints which are synthetically check by `fabric-synthetics`.

## Alerts

> List of alerts configured for the service.





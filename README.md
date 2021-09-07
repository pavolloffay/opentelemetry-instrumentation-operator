# OpenTelemetry instrumentation operator

This project is POC/work in progress!

OpenTelemetry instrumentation operator injects OpenTelemetry auto instrumentation into deployment.
For the configuration see [instrumentation-cr](./examples/04-instrumentation.yaml).

## Enable instrumentation

Enable instrumentation for all applications in a namespace:

```bash
kubectl label namespace/default opentelemetry-java-enabled=true
```

### Java

```bash
kubectl label deployment.apps/java-app opentelemetry-java-enabled=true
```

### Python

```bash
kubectl label deployment.apps/python-app opentelemetry-python-enabled=true
```

## List instrumented apps

```bash
kubectl get deployments -l opentelemetry-java-enabled=true
```

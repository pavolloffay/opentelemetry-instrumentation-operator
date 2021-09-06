# OpenTelemetry instrumentation operator

OpenTelemetry instrumentation operator injects OpenTelemetry auto instrumentation into deployment.
For the configuration see [instrumentation-cr](./examples/04-instrumentation.yaml).

## Enable instrumentation

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

# OpenTelemetry instrumentation operator

OpenTelemetry instrumentation operator

## Enable instrumentation

### Java

```bash
kubectl label deployment.apps/spring-petclinic opentelemetry-java-enabled=true
```

### Python

```bash
kubectl label deployment.apps/python-app opentelemetry-python-enabled=true
```

## List instrumented apps

```bash
kubectl get deployments -l opentelemetry-java-enabled=true
```

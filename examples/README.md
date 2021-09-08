# Spring Pet Clinic example

## Install

### 1. Install tracing infrastructure and Spring Pet Clinic

```bash
kubectl apply -f examples/01*
kubectl apply -f examples/02*
kubectl apply -f examples/03*
```

### 2. Expose service to host

```bash

kubectl port-forward service/jaeger-query 16686:16686 -n jaeger
kubectl port-forward deployment.apps/spring-petclinic  8090:8080
```

### 3. Create instrumentation configuration - CR

```bash
kubectl apply -f examples/04*
```

### 4. Instrument the application

```bash
kubectl label deployment.apps/spring-petclinic opentelemetry-inst-java=enabled
```

or instrument all deployments in the namespace

```bash
kubectl label namespace/default opentelemetry-inst-java=enabled
```

## Change instrumentation config

```bash
kubectl edit opentelemetryinstrumentations opentelemetry-instrumentation
```

Add arbitrary attribute.

The workload pod should restart with the new configuration.


Or add annotation to workload:

```bash
kubectl annotate deployment.apps/spring-petclinic otel.tracesSamplerArg=0.5
```

# OpenTelemetry instrumentation operator

This project is POC/work in progress!

OpenTelemetry instrumentation operator injects OpenTelemetry auto instrumentation into deployment.
For the configuration see [instrumentation-cr](./examples/04-instrumentation.yaml).

## Enable instrumentation

Right now the instrumentation CR has to be created in every namespace where instrumentation is enabled.
Then label `opentelemetry-inst-java=enabled` has to be added on a workload or namespace.

Create the following CR in a namespace:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: opentelemetry.io/v1alpha1
kind: OpenTelemetryInstrumentation
metadata:
  name: opentelemetry-instrumentation
spec:
  OTLPEndpoint: http://otel-collector.otel:4317
  javaagentImage: ghcr.io/pavolloffay/otel-javaagent:1.5.3
  tracesSampler: parentbased_traceidratio
  tracesSamplerArg: "1"
  resourceAttributes:
    environment: prod
EOF
```

Enable instrumentation for all applications in a namespace:

```bash
kubectl label namespace/default opentelemetry-inst-java=enabled
```

### Enable instrumentation per workload

```bash
kubectl label deployment.apps/java-app opentelemetry-inst-java=enabled
```

## List instrumented apps

```bash
kubectl get deployments -l opentelemetry-java-enabled=true
```

The status object will list instrumentations in the future.

## List Instrumentation CRs

```bash
kubectl get opentelemetryinstrumentations 
```

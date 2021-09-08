package inject

import (
	"fmt"
	"strings"

	cachev1alpha1 "github.com/pavolloffay/opentelemetry-instrumentation-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	initContainerName = "opentelemetry-auto-instrumentation"
	volumeName        = "opentelemetry-auto-instrumentation"

	envJavaToolsOptions         = "JAVA_TOOL_OPTIONS"
	envOTELServiceName          = "OTEL_SERVICE_NAME"
	envOTELTracesSampler        = "OTEL_TRACES_SAMPLER"
	envOTELTracesSamplerArg     = "OTEL_TRACES_SAMPLER_ARG"
	envOTELResourceAttrs        = "OTEL_RESOURCE_ATTRIBUTES"
	envOTELExporterOTLPEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"

	javaJVMArgument = " -javaagent:/otel-auto-instrumentation/javaagent.jar"
)

func IsInstrumentationEnabled(label string, meta ...metav1.ObjectMeta) bool {
	for _, ometa := range meta {
		val, ok := ometa.Labels[label]
		if !ok {
			continue
		}
		return val == "true"
	}
	return false
}

func InjectPod(workloadMeta metav1.ObjectMeta, pod *corev1.PodSpec, instrumentation cachev1alpha1.OpenTelemetryInstrumentationSpec) {
	idx := getIndexOfContainer(pod.InitContainers, "opentelemetry-auto-instrumentation")
	if idx == -1 {
		pod.InitContainers = append(pod.InitContainers, corev1.Container{
			Name:            initContainerName,
			Image:           instrumentation.JavaagentImage,
			ImagePullPolicy: corev1.PullAlways,
			Command:         []string{"cp", "/javaagent.jar", "/otel-auto-instrumentation/javaagent.jar"},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      volumeName,
				MountPath: "/otel-auto-instrumentation",
			}},
		})
	}

	idx = getIndexOfVolume(pod.Volumes, "opentelemetry-auto-instrumentation")
	if idx == -1 {
		pod.Volumes = append(pod.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}})
	}

	injectContainer(workloadMeta, &pod.Containers[0], instrumentation)
}

func injectContainer(parentMeta metav1.ObjectMeta, container *corev1.Container, inst cachev1alpha1.OpenTelemetryInstrumentationSpec) {
	idx := getIndexOfEnv(container.Env, envJavaToolsOptions)
	if idx > -1 && strings.Contains(container.Env[idx].Value, javaJVMArgument) {
		// nothing
	} else if idx == -1 {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  envJavaToolsOptions,
			Value: javaJVMArgument,
		})
	}

	idx = getIndexOfEnv(container.Env, envOTELExporterOTLPEndpoint)
	if idx > -1 {
		container.Env[idx].Value = inst.OTLPEndpoint
	} else {
		container.Env = append(container.Env, corev1.EnvVar{Name: envOTELExporterOTLPEndpoint, Value: inst.OTLPEndpoint})
	}

	idx = getIndexOfVolumeMount(container.VolumeMounts, volumeName)
	if idx == -1 {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: "/otel-auto-instrumentation",
		})
	}

	idx = getIndexOfEnv(container.Env, envOTELServiceName)
	if idx > -1 {
		container.Env[idx].Value = parentMeta.Name
	} else {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  envOTELServiceName,
			Value: parentMeta.Name,
		})
	}

	if len(inst.ResourceAttributes) > 0 {
		resourceAttributes := ""
		for k, v := range inst.ResourceAttributes {
			if resourceAttributes != "" {
				resourceAttributes += ","
			}
			resourceAttributes += fmt.Sprintf("%s=%s", k, v)
		}
		resourceAttributes += ",k8s.namespace=" + parentMeta.Namespace
		resourceAttributes += ",k8s.deployment=" + parentMeta.Name
		//resourceAttributes += ",k8s.pod=" + metadata.podName
		resourceAttributes += ",k8s.container=" + container.Name

		idx = getIndexOfEnv(container.Env, envOTELResourceAttrs)
		if idx > -1 {
			container.Env[idx].Value = resourceAttributes
		} else {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  envOTELResourceAttrs,
				Value: resourceAttributes,
			})
		}
	}

	if inst.TracesSampler != "" {
		sampler := inst.TracesSampler
		if samplerAnnotation := parentMeta.GetAnnotations()["otel.tracesSampler"]; samplerAnnotation != "" {
			sampler = samplerAnnotation
		}

		idx = getIndexOfEnv(container.Env, envOTELTracesSampler)
		if idx > -1 {
			container.Env[idx].Value = sampler
		} else {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  envOTELTracesSampler,
				Value: sampler,
			})
		}
	}

	if inst.TracesSamplerArg != "" {
		samplerArg := inst.TracesSamplerArg
		if samplerAnnotationArg := parentMeta.GetAnnotations()["otel.tracesSamplerArg"]; samplerAnnotationArg != "" {
			samplerArg = samplerAnnotationArg
		}

		idx = getIndexOfEnv(container.Env, envOTELTracesSamplerArg)
		if idx > -1 {
			container.Env[idx].Value = samplerArg
		} else {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  envOTELTracesSamplerArg,
				Value: samplerArg,
			})
		}
	}
}

func getIndexOfEnv(envs []corev1.EnvVar, name string) int {
	for i := range envs {
		if envs[i].Name == name {
			return i
		}
	}
	return -1
}

func getIndexOfVolumeMount(mounts []corev1.VolumeMount, name string) int {
	for i := range mounts {
		if mounts[i].Name == name {
			return i
		}
	}
	return -1
}

func getIndexOfVolume(volumes []corev1.Volume, name string) int {
	for i := range volumes {
		if volumes[i].Name == name {
			return i
		}
	}
	return -1
}

func getIndexOfContainer(containers []corev1.Container, name string) int {
	for i := range containers {
		if containers[i].Name == name {
			return i
		}
	}
	return -1
}

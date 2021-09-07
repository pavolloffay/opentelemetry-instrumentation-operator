package controllers

import (
	"fmt"
	"strings"

	cachev1alpha1 "github.com/pavolloffay/opentelemetry-instrumentation-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isInstrumentationEnabled(label string, meta ...metav1.ObjectMeta) bool {
	for _, ometa := range meta {
		val, ok := ometa.Labels[label]
		if !ok {
			continue
		}
		return val == "true"
	}
	return false
}

func injectPod(metadata metadata, workloadMeta metav1.ObjectMeta, pod *corev1.PodSpec, instrumentation cachev1alpha1.OpenTelemetryInstrumentationSpec) {
	idx := getIndexOfContainer(pod.InitContainers, "opentelemetry-auto-instrumentation")
	if idx == -1 {
		pod.InitContainers = append(pod.InitContainers, corev1.Container{
			Name:            "opentelemetry-auto-instrumentation",
			Image:           instrumentation.JavaagentImage,
			ImagePullPolicy: corev1.PullAlways,
			Command:         []string{"cp", "/javaagent.jar", "/otel-auto-instrumentation/javaagent.jar"},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "opentelemetry-auto-instrumentation",
				MountPath: "/otel-auto-instrumentation",
			}},
		})
	}

	idx = getIndexOfVolume(pod.Volumes, "opentelemetry-auto-instrumentation")
	if idx == -1 {
		pod.Volumes = append(pod.Volumes, corev1.Volume{
			Name: "opentelemetry-auto-instrumentation",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}})
	}

	injectContainer(metadata, workloadMeta, &pod.Containers[0], instrumentation)
}

func injectContainer(metadata metadata, parentMeta metav1.ObjectMeta, container *corev1.Container, inst cachev1alpha1.OpenTelemetryInstrumentationSpec) {
	javaagent := " -javaagent:/otel-auto-instrumentation/javaagent.jar"
	idx := getIndexOfEnv(container.Env, "JAVA_TOOL_OPTIONS")
	if idx > -1 && strings.Contains(container.Env[idx].Value, javaagent) {
		// nothing
	} else if idx == -1 {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "JAVA_TOOL_OPTIONS",
			Value: javaagent,
		})
	}

	idx = getIndexOfEnv(container.Env, "OTEL_EXPORTER_OTLP_ENDPOINT")
	if idx > -1 {
		container.Env[idx].Value = inst.OTLPEndpoint
	} else {
		container.Env = append(container.Env, corev1.EnvVar{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: inst.OTLPEndpoint})
	}

	idx = getIndexOfVolumeMount(container.VolumeMounts, "opentelemetry-auto-instrumentation")
	if idx == -1 {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "opentelemetry-auto-instrumentation",
			MountPath: "/otel-auto-instrumentation",
		})
	}

	idx = getIndexOfEnv(container.Env, "OTEL_SERVICE_NAME")
	if idx > -1 {
		container.Env[idx].Value = metadata.deploymentName
	} else {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "OTEL_SERVICE_NAME",
			Value: metadata.deploymentName,
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
		resourceAttributes += ",k8s.namespace=" + metadata.namespace
		resourceAttributes += ",k8s.deployment=" + metadata.deploymentName
		//resourceAttributes += ",k8s.pod=" + metadata.podName
		resourceAttributes += ",k8s.container=" + metadata.containerName

		idx = getIndexOfEnv(container.Env, "OTEL_RESOURCE_ATTRIBUTES")
		if idx > -1 {
			container.Env[idx].Value = resourceAttributes
		} else {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  "OTEL_RESOURCE_ATTRIBUTES",
				Value: resourceAttributes,
			})
		}
	}

	if inst.TracesSampler != "" {
		sampler := inst.TracesSampler
		if samplerAnnotation := parentMeta.GetAnnotations()["otel.tracesSampler"]; samplerAnnotation != "" {
			sampler = samplerAnnotation
		}

		idx = getIndexOfEnv(container.Env, "OTEL_TRACES_SAMPLER")
		if idx > -1 {
			container.Env[idx].Value = sampler
		} else {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  "OTEL_TRACES_SAMPLER",
				Value: sampler,
			})
		}
	}

	if inst.TracesSamplerArg != "" {
		samplerArg := inst.TracesSamplerArg
		if samplerAnnotationArg := parentMeta.GetAnnotations()["otel.tracesSamplerArg"]; samplerAnnotationArg != "" {
			samplerArg = samplerAnnotationArg
		}

		idx = getIndexOfEnv(container.Env, "OTEL_TRACES_SAMPLER_ARG")
		if idx > -1 {
			container.Env[idx].Value = samplerArg
		} else {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  "OTEL_TRACES_SAMPLER_ARG",
				Value: samplerArg,
			})
		}
	}
}

type metadata struct {
	namespace      string
	deploymentName string
	// pod name is not known at this time
	//podName        string
	containerName string
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

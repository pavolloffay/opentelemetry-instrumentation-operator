package inject

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func Clean(pod *corev1.PodSpec) bool {
	initContainerIdx := -1
	for i, c := range pod.InitContainers {
		if c.Name == initContainerName {
			fmt.Println(c.Name)
			initContainerIdx = i
			break
		}
	}
	if initContainerIdx == -1 {
		return false
	}

	pod.InitContainers = append(pod.InitContainers[:initContainerIdx], pod.InitContainers[initContainerIdx+1:]...)
	for i, v := range pod.Volumes {
		if v.Name == volumeName {
			pod.Volumes = append(pod.Volumes[:i], pod.Volumes[i+1:]...)
			break
		}
	}

	if len(pod.Containers) < 1 {
		// nothing to remove from containers - should not happen
		return true
	}
	appContainer := &pod.Containers[0]

	for i, vm := range appContainer.VolumeMounts {
		if vm.Name == volumeName {
			appContainer.VolumeMounts = append(appContainer.VolumeMounts[:i], appContainer.VolumeMounts[i+1:]...)
			break
		}
	}
	removeEnvVar(envOTELExporterOTLPEndpoint, appContainer)
	removeEnvVar(envOTELServiceName, appContainer)
	removeEnvVar(envOTELResourceAttrs, appContainer)
	removeEnvVar(envOTELTracesSampler, appContainer)
	removeEnvVar(envOTELTracesSamplerArg, appContainer)

	for i, e := range appContainer.Env {
		if appContainer.Env[i].Name == envJavaToolsOptions {
			appContainer.Env[i].Value = strings.Replace(e.Value, javaJVMArgument, "", 1)
			if appContainer.Env[i].Value == "" {
				appContainer.Env = append(appContainer.Env[:i], appContainer.Env[i+1:]...)
			}
			break
		}
	}
	return true
}

func removeEnvVar(name string, container *corev1.Container) {
	for i, e := range container.Env {
		if e.Name == name {
			container.Env = append(container.Env[:i], container.Env[i+1:]...)
			break
		}
	}
}

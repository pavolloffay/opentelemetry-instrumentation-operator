/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	cachev1alpha1 "github.com/pavolloffay/opentelemetry-instrumentation-operator/api/v1alpha1"
)

type PodControllerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=opentelemetry.io,resources=opentelemetryinstrumentations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opentelemetry.io,resources=opentelemetryinstrumentations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opentelemetry.io,resources=opentelemetryinstrumentations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenTelemetryInstrumentation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *PodControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Println("pod controller")
	fmt.Println(req.Name)
	fmt.Println(req.Namespace)
	fmt.Println(req.NamespacedName)

	dep := &v1.Deployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, dep); err != nil {
		fmt.Println("getting deployment failed")
		return ctrl.Result{}, err
	}

	if dep.Labels["opentelemetry-java-enabled"] == "true" {
		fmt.Println("injection enabled")
		instr := &cachev1alpha1.OpenTelemetryInstrumentation{}
		if err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: req.Namespace,
			Name:      "opentelemetry-instrumentation",
		}, instr); err != nil {
			return ctrl.Result{}, err
		}

		fmt.Println("injection is enabled")
		injectPod(&(dep.Spec.Template.Spec), instr.Spec)

		if err := r.Client.Update(ctx, dep); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		fmt.Println("injection disabled")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Deployment{}).
		Complete(r)
}

func injectPod(pod *corev1.PodSpec, instrumentation cachev1alpha1.OpenTelemetryInstrumentationSpec) {
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

	injectContainer(&pod.Containers[0], instrumentation)
}

func injectContainer(container *corev1.Container, inst cachev1alpha1.OpenTelemetryInstrumentationSpec) {
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

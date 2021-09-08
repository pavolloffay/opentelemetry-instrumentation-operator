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
	v1alpha1 "github.com/pavolloffay/opentelemetry-instrumentation-operator/api/v1alpha1"
	"github.com/pavolloffay/opentelemetry-instrumentation-operator/inject"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type NamespaceControllerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=opentelemetry.io,resources=opentelemetryinstrumentations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opentelemetry.io,resources=opentelemetryinstrumentations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opentelemetry.io,resources=opentelemetryinstrumentations/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *NamespaceControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	fmt.Println("Namespace reconcile: " + req.Namespace + "/" + req.Name)

	ns := &corev1.Namespace{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name: req.Name,
	}, ns); err != nil {
		return ctrl.Result{}, err
	}

	deps := &v1.DeploymentList{}
	if err := r.Client.List(ctx, deps, client.InNamespace(req.Name)); err != nil {
		return ctrl.Result{}, err
	}

	instrumentation := &v1alpha1.OpenTelemetryInstrumentation{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: req.Name,
		Name:      "opentelemetry-instrumentation",
	}, instrumentation)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("opentelemetry-instrumentation CR does not exists in: " + req.Namespace)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	for _, dep := range deps.Items {
		if inject.IsInstrumentationEnabled(javaInstrumentationLablel, dep.ObjectMeta, ns.ObjectMeta) {
			m := inject.Metadata{
				Namespace:      req.Namespace,
				DeploymentName: dep.Name,
				ContainerName:  dep.Spec.Template.Spec.Containers[0].Name,
			}
			inject.InjectPod(m, dep.ObjectMeta, &dep.Spec.Template.Spec, instrumentation.Spec)
			if err := r.Client.Update(ctx, &dep); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			objectUpdated := inject.Clean(&dep.Spec.Template.Spec)
			if objectUpdated {
				fmt.Println("-----> Updating removed instance " + req.String())
				if err := r.Client.Update(ctx, &dep); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

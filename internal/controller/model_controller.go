/*
Copyright 2025.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	ollamav1alpha1 "github.com/sivchari/ollama-operator/api/v1alpha1"
)

// ModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	OllamaContainerImage string
}

// +kubebuilder:rbac:groups=ollama.sivchari.io,resources=models,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ollama.sivchari.io,resources=models/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ollama.sivchari.io,resources=models/finalizers,verbs=update
func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	model := &ollamav1alpha1.Model{}
	if err := r.Get(ctx, req.NamespacedName, model); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	newModel := model.DeepCopy()
	patch := client.MergeFrom(model)
	defer func() {
		if err := r.Status().Patch(ctx, newModel, patch); err != nil {
			ctrl.LoggerFrom(ctx).V(1).Error(err, "unable to update Model status")
		}

		if err := r.Patch(ctx, newModel, patch); err != nil {
			ctrl.LoggerFrom(ctx).V(1).Error(err, "unable to update Model status")
		}
	}()
	if !model.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(ctx, model)
	}
	return ctrl.Result{}, r.reconcileNormal(ctx, newModel)
}

func (r *ModelReconciler) reconcileDelete(ctx context.Context, model *ollamav1alpha1.Model) error {
	if !controllerutil.ContainsFinalizer(model, ollamav1alpha1.ModelFinalizer) {
		return nil
	}
	pod := &corev1.Pod{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: model.Namespace, Name: model.Name}, pod); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := r.Delete(ctx, pod); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}
	controllerutil.RemoveFinalizer(model, ollamav1alpha1.ModelFinalizer)
	return nil
}

func (r *ModelReconciler) reconcileNormal(ctx context.Context, model *ollamav1alpha1.Model) error {
	if !controllerutil.ContainsFinalizer(model, ollamav1alpha1.ModelFinalizer) {
		controllerutil.AddFinalizer(model, ollamav1alpha1.ModelFinalizer)
		if err := r.Update(ctx, model); err != nil {
			return err
		}
	}
	if err := r.reconcilePod(ctx, model); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ollamav1alpha1.Model{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(podToModel),
		).
		Named("model").
		Complete(r)
}

func podToModel(ctx context.Context, obj client.Object) []ctrl.Request {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil
	}
	ownerRefs := pod.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		return nil
	}
	requests := make([]ctrl.Request, 0)
	for _, ownerRef := range ownerRefs {
		if ownerRef.Kind != "Model" && ownerRef.APIVersion != ollamav1alpha1.GroupVersion.String() {
			continue
		}
		requests = append(requests, ctrl.Request{
			NamespacedName: client.ObjectKey{
				Namespace: pod.Namespace,
				Name:      ownerRef.Name,
			},
		})
	}
	return requests
}

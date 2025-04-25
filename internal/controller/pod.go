package controller

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	ollamav1alpha1 "github.com/sivchari/ollama-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ModelReconciler) reconcilePod(ctx context.Context, model *ollamav1alpha1.Model) error {
	pod := &corev1.Pod{}
	var name string
	if model.Status.PodRef != nil {
		name = model.Status.PodRef.Name
	}
	err := r.Get(ctx, client.ObjectKey{Namespace: model.Namespace, Name: name}, pod)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return r.createPod(ctx, model)
	}
	return r.updatePod(ctx, model, pod)
}

func (r *ModelReconciler) createPod(ctx context.Context, model *ollamav1alpha1.Model) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    model.Namespace,
			GenerateName: fmt.Sprintf("%s-", model.Name),
		},
	}
	if err := controllerutil.SetOwnerReference(model, pod, r.Scheme); err != nil {
		return err
	}
	r.modelToPod(model, pod)
	if err := r.Create(ctx, pod); err != nil {
		return err
	}
	model.Status.PodRef = &corev1.ObjectReference{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
	return nil
}

func (r *ModelReconciler) updatePod(ctx context.Context, model *ollamav1alpha1.Model, pod *corev1.Pod) error {
	before := pod.DeepCopy()
	r.modelToPod(model, pod)
	if !equality.Semantic.DeepEqual(before.Spec, pod.Spec) {
		if err := r.Delete(ctx, before); err != nil {
			return err
		}
		if err := r.createPod(ctx, model); err != nil {
			return err
		}
	}
	return nil
}

func (r *ModelReconciler) modelToPod(model *ollamav1alpha1.Model, pod *corev1.Pod) {
	image := "ollama/ollama:latest"
	if r.OllamaContainerImage != "" {
		image = r.OllamaContainerImage
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("postStart").Parse(postStartScript))
	_ = tmpl.Execute(&buf, PostStartInput{
		Images: model.Spec.Images,
	})

	if model.Spec.Template != nil && model.Spec.Template.Metadata != nil {
		pod.Labels = model.Spec.Template.Metadata.Labels
		pod.Annotations = model.Spec.Template.Metadata.Annotations
	}

	var volumeMounts []corev1.VolumeMount
	if model.Spec.Template != nil && model.Spec.Template.Spec != nil {
		pod.Spec.Volumes = model.Spec.Template.Spec.Volumes
		pod.Spec.NodeSelector = model.Spec.Template.Spec.NodeSelector
		pod.Spec.Affinity = model.Spec.Template.Spec.Affinity
		pod.Spec.Tolerations = model.Spec.Template.Spec.Tolerations
		pod.Spec.TopologySpreadConstraints = model.Spec.Template.Spec.TopologySpreadConstraints
		volumeMounts = append(volumeMounts, model.Spec.Template.Spec.VolumeMounts...)
	}

	pod.Spec.Containers = []corev1.Container{
		{
			Name:  "ollama-server",
			Image: image,
			Ports: []corev1.ContainerPort{
				{
					Name:          "ollama-server",
					ContainerPort: 11434,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			VolumeMounts: volumeMounts,
			// TODO: Deal with other env vars
			Env: []corev1.EnvVar{
				{
					Name:  "OLLAMA_HOST",
					Value: "0.0.0.0:11434",
				},
			},
			Args: []string{
				"serve",
			},
			Lifecycle: &corev1.Lifecycle{
				PostStart: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/bin/sh",
							"-c",
							buf.String(),
						},
					},
				},
			},
		},
	}
}

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
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ollamav1alpha1 "github.com/sivchari/ollama-operator/api/v1alpha1"
)

const (
	modelReconcilerName      = "model-reconciler-test"
	modelReconcilerNamespace = "model-reconciler-test"
)

func TestModelReconciler(t *testing.T) {
	ns, err := env.CreateNamespace(ctx, modelReconcilerNamespace)
	if err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}
	t.Cleanup(func() {
		if err := env.Delete(ctx, ns); err != nil {
			t.Fatalf("failed to delete namespace: %v", err)
		}
	})

	t.Run("Should create Model and Pod", func(t *testing.T) {
		g := NewWithT(t)
		model := &ollamav1alpha1.Model{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: modelReconcilerName,
				Namespace:    ns.Name,
			},
			Spec: ollamav1alpha1.ModelSpec{
				Images: []string{"llama3"},
			},
		}
		g.Expect(env.Create(ctx, model)).To(Succeed())
		t.Cleanup(func() {
			g.Expect(env.Delete(ctx, model)).To(Succeed())
		})

		key := client.ObjectKey{
			Name:      model.Name,
			Namespace: ns.Name,
		}

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			g.Expect(model.GetFinalizers()).To(ContainElement(ollamav1alpha1.ModelFinalizer))
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			g.Expect(model.Status.PodRef).NotTo(BeNil())
			g.Expect(model.Status.PodRef.Name).NotTo(BeEmpty())
			g.Expect(model.Status.PodRef.Namespace).To(Equal(ns.Name))
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())

			pod := &corev1.Pod{}
			g.Expect(env.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: model.Status.PodRef.Name}, pod)).To(Succeed())
			g.Expect(pod.Spec.Containers).To(HaveLen(1))
			g.Expect(pod.Spec.Containers[0].Image).To(Equal("ollama/ollama:latest"))
			g.Expect(pod.Spec.Containers[0].Lifecycle.PostStart.Exec.Command).To(Equal([]string{"/bin/sh", "-c", "ollama pull llama3;"}))
		}).Should(Succeed())
	})

	t.Run("Should recreate Pod when the Model is updated", func(t *testing.T) {
		g := NewWithT(t)
		model := &ollamav1alpha1.Model{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: modelReconcilerName,
				Namespace:    ns.Name,
			},
			Spec: ollamav1alpha1.ModelSpec{
				Images: []string{"llama3"},
			},
		}
		g.Expect(env.Create(ctx, model)).To(Succeed())
		t.Cleanup(func() {
			g.Expect(env.Delete(ctx, model)).To(Succeed())
		})

		key := client.ObjectKey{
			Name:      model.Name,
			Namespace: ns.Name,
		}

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			g.Expect(model.GetFinalizers()).To(ContainElement(ollamav1alpha1.ModelFinalizer))
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			g.Expect(model.Status.PodRef).NotTo(BeNil())
			g.Expect(model.Status.PodRef.Name).NotTo(BeEmpty())
			g.Expect(model.Status.PodRef.Namespace).To(Equal(ns.Name))
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			pod := &corev1.Pod{}
			g.Expect(env.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: model.Status.PodRef.Name}, pod)).To(Succeed())
			g.Expect(pod.Spec.Containers).To(HaveLen(1))
			g.Expect(pod.Spec.Containers[0].Image).To(Equal("ollama/ollama:latest"))
			g.Expect(pod.Spec.Containers[0].Lifecycle.PostStart.Exec.Command).To(Equal([]string{"/bin/sh", "-c", "ollama pull llama3;"}))
		}).Should(Succeed())

		model.Spec.Images = append(model.Spec.Images, "hf.co/mlabonne/Meta-Llama-3.1-8B-Instruct-abliterated-GGUF")
		g.Expect(updateModel(model)).To(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			pod := &corev1.Pod{}
			g.Expect(env.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: model.Status.PodRef.Name}, pod)).To(Succeed())
			g.Expect(pod.Spec.Containers).To(HaveLen(1))
			g.Expect(pod.Spec.Containers[0].Image).To(Equal("ollama/ollama:latest"))
			g.Expect(pod.Spec.Containers[0].Lifecycle.PostStart.Exec.Command).To(Equal([]string{"/bin/sh", "-c", "ollama pull llama3;ollama pull hf.co/mlabonne/Meta-Llama-3.1-8B-Instruct-abliterated-GGUF;"}))
		}).Should(Succeed())
	})

	t.Run("Should recreate Pod when the current Pod is deleted", func(t *testing.T) {
		g := NewWithT(t)
		model := &ollamav1alpha1.Model{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: modelReconcilerName,
				Namespace:    ns.Name,
			},
			Spec: ollamav1alpha1.ModelSpec{
				Images: []string{"llama3"},
			},
		}
		g.Expect(env.Create(ctx, model)).To(Succeed())
		t.Cleanup(func() {
			g.Expect(env.Delete(ctx, model)).To(Succeed())
		})

		key := client.ObjectKey{
			Name:      model.Name,
			Namespace: ns.Name,
		}

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			g.Expect(model.GetFinalizers()).To(ContainElement(ollamav1alpha1.ModelFinalizer))
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			g.Expect(model.Status.PodRef).NotTo(BeNil())
			g.Expect(model.Status.PodRef.Name).NotTo(BeEmpty())
			g.Expect(model.Status.PodRef.Namespace).To(Equal(ns.Name))
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			pod := &corev1.Pod{}
			g.Expect(env.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: model.Status.PodRef.Name}, pod)).To(Succeed())
			g.Expect(pod.Spec.Containers).To(HaveLen(1))
			g.Expect(pod.Spec.Containers[0].Image).To(Equal("ollama/ollama:latest"))
			g.Expect(pod.Spec.Containers[0].Lifecycle.PostStart.Exec.Command).To(Equal([]string{"/bin/sh", "-c", "ollama pull llama3;"}))
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			pod := &corev1.Pod{}
			g.Expect(env.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: model.Status.PodRef.Name}, pod)).To(Succeed())
			g.Expect(env.Delete(ctx, pod)).To(Succeed())
		}).Should(Succeed())

		g.Eventually(func(g Gomega) {
			model := &ollamav1alpha1.Model{}
			g.Expect(env.Get(ctx, key, model)).To(Succeed())
			pod := &corev1.Pod{}
			g.Expect(env.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: model.Status.PodRef.Name}, pod)).To(Succeed())
			g.Expect(pod.Spec.Containers).To(HaveLen(1))
			g.Expect(pod.Spec.Containers[0].Image).To(Equal("ollama/ollama:latest"))
			g.Expect(pod.Spec.Containers[0].Lifecycle.PostStart.Exec.Command).To(Equal([]string{"/bin/sh", "-c", "ollama pull llama3;"}))
		}).Should(Succeed())
	})
}

func updateModel(obj *ollamav1alpha1.Model) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		model := &ollamav1alpha1.Model{}
		if err := env.Get(ctx, client.ObjectKey{Namespace: obj.Namespace, Name: obj.Name}, model); err != nil {
			return err
		}
		model.Spec = obj.Spec
		if err := env.Update(ctx, model); err != nil {
			return err
		}
		return nil
	})
}

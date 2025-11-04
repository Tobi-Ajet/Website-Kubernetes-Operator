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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/Tobi-Ajet/Website-Kubernetes-Operator/api/v1alpha1"
)

var _ = Describe("Website Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		website := &appsv1alpha1.Website{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Website")

			err := k8sClient.Get(ctx, typeNamespacedName, website)
			if err != nil && errors.IsNotFound(err) {
				replicas := int32(1)
				resource := &appsv1alpha1.Website{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: appsv1alpha1.WebsiteSpec{
						Replicas:    &replicas,
						IndexHTML:   "<html><body><h1>Test Website</h1></body></html>",
						ServiceType: "ClusterIP",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("cleaning up the specific Website resource instance")
			resource := &appsv1alpha1.Website{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			controllerReconciler := &WebsiteReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify that the resource exists and has expected fields
			fetched := &appsv1alpha1.Website{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, fetched)).To(Succeed())
			Expect(fetched.Spec.Replicas).NotTo(BeNil())
			Expect(*fetched.Spec.Replicas).To(Equal(int32(1)))
			Expect(fetched.Spec.IndexHTML).To(ContainSubstring("Test Website"))
		})
	})
})

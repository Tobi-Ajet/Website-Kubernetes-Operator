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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/youruser/website-operator/api/v1alpha1"
)

const (
	webPort         = 8080
	indexKey        = "index.html"
	cmNameSuffix    = "-content"
	deployNameSfx   = "-deploy"
	svcNameSfx      = "-svc"
	defaultImage    = "nginx:1.27-alpine"
	labelKey        = "app.kubernetes.io/name"
)

// WebsiteReconciler reconciles a Website object
type WebsiteReconciler struct {
	ctrl.Client
	Scheme *runtime.Scheme
}

// RBAC markers:
//+kubebuilder:rbac:groups=apps.prophix.cloud,resources=websites,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.prophix.cloud,resources=websites/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.prophix.cloud,resources=websites/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps;services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *WebsiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var site appsv1alpha1.Website
	if err := r.Get(ctx, req.NamespacedName, &site); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Desired names
	cmName := site.Name + cmNameSuffix
	deployName := site.Name + deployNameSfx
	svcName := site.Name + svcNameSfx

	labels := map[string]string{
		labelKey: site.Name,
	}

	// 1) ConfigMap with index.html
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: site.Namespace,
			Labels:    labels,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[indexKey] = site.Spec.IndexHTML
		return controllerutil.SetControllerReference(&site, cm, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// 2) Deployment serving the HTML via nginx
	replicas := ptr.Deref(site.Spec.Replicas, 1)
	image := defaultImage
	if site.Spec.Image != "" {
		image = site.Spec.Image
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: site.Namespace,
			Labels:    labels,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		deploy.Spec.Replicas = &replicas
		deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		deploy.Spec.Template.ObjectMeta.Labels = labels
		deploy.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:  "nginx",
				Image: image,
				Ports: []corev1.ContainerPort{{ContainerPort: webPort}},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "html",
					MountPath: "/usr/share/nginx/html",
				}},
			},
		}
		deploy.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "html",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: cmName},
						Items: []corev1.KeyToPath{{
							Key:  indexKey,
							Path: "index.html",
						}},
					},
				},
			},
		}
		return controllerutil.SetControllerReference(&site, deploy, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// 3) Service
	svcType := corev1.ServiceTypeClusterIP
	if site.Spec.ServiceType != "" {
		svcType = corev1.ServiceType(site.Spec.ServiceType)
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: site.Namespace,
			Labels:    labels,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec.Type = svcType
		svc.Spec.Selector = labels
		svc.Spec.Ports = []corev1.ServicePort{{
			Name:       "http",
			Port:       80,
			TargetPort: intstr.FromInt(webPort),
		}}
		return controllerutil.SetControllerReference(&site, svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// 4) Update status
	var freshDeploy appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: deployName, Namespace: site.Namespace}, &freshDeploy); err == nil {
		site.Status.AvailableReplicas = freshDeploy.Status.AvailableReplicas
	}
	site.Status.ServiceName = svcName
	baseURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", svcName, site.Namespace)
	site.Status.URL = baseURL

	if err := r.Status().Update(ctx, &site); err != nil {
		// If conflict, requeue
		if apierrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("reconciled Website", "name", site.Name, "available", site.Status.AvailableReplicas)
	return ctrl.Result{}, nil
}

func (r *WebsiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Website{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

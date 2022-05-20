/*
Copyright 2022.

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
	//"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	monitoringv1alpha1 "github.com/marieroque/best-prometheus-operator-in-the-world/api/v1alpha1"
)

// PrometheusReconciler reconciles a Prometheus object
type PrometheusReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const container_image = "quay.io/prometheus/prometheus"

//+kubebuilder:rbac:groups=monitoring.mroque,resources=prometheuses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.mroque,resources=prometheuses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitoring.mroque,resources=prometheuses/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Prometheus object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *PrometheusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Prometheus instance
	prometheus := &monitoringv1alpha1.Prometheus{}
	err := r.Get(ctx, req.NamespacedName, prometheus)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Prometheus resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Prometheus")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: prometheus.Name, Namespace: prometheus.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForPrometheus(prometheus)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// This point, we have the deployment object created
	// Ensure the image version is same as the spec
	img := container_image + ":v" + *prometheus.Spec.Version
	if found.Spec.Template.Spec.Containers[0].Image != img {
		found.Spec.Template.Spec.Containers[0].Image = img
		log.Info("Updating Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated return and requeue
		// Requeue for any reason other than an error
		return ctrl.Result{Requeue: true}, nil
	}

	// Check if the configmap already exists, if not create a new one
	foundConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: prometheus.Name + "-configmap", Namespace: prometheus.Namespace}, foundConfigMap)
	if err != nil && errors.IsNotFound(err) {
		// Define a new configmap
		cfm := r.configmapForPrometheus(prometheus)
		log.Info("Creating a new Configmap", "Configmap.Namespace", cfm.Namespace, "Configmap.Name", cfm.Name)
		err = r.Create(ctx, cfm)
		if err != nil {
			log.Error(err, "Failed to create new Configmap", "Configmap.Namespace", cfm.Namespace, "Configmap.Name", cfm.Name)
			return ctrl.Result{}, err
		}
		// Configmap created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Configmap")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// deploymentForPrometheus returns a prometheus Deployment object
func (r *PrometheusReconciler) deploymentForPrometheus(cr *monitoringv1alpha1.Prometheus) *appsv1.Deployment {
	ls := labelsForPrometheus(cr.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "prometheus",
						Image: container_image + ":v" + *cr.Spec.Version,
						Args:  []string{"--config.file=/etc/prometheus/prometheus.yml"},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/etc/prometheus/",
							Name:      "prometheus-config-volume",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "prometheus-config-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-configmap",
								},
							},
						},
					}},
					ServiceAccountName: cr.Namespace + "-controller-manager",
				},
			},
		},
	}
	// Set Prometheus instance as the owner and controller
	ctrl.SetControllerReference(cr, dep, r.Scheme)
	return dep
}

// labelsForPrometheus returns the labels for selecting the resources
// belonging to the given prometheus CR name.
func labelsForPrometheus(name string) map[string]string {
	return map[string]string{"app": "prometheus", "prometheus_cr": name}
}

// configmapForPrometheus returns a prometheus ConfigMap object
func (r *PrometheusReconciler) configmapForPrometheus(cr *monitoringv1alpha1.Prometheus) *corev1.ConfigMap {
	labels := map[string]string{
		"app": cr.Name,
	}
	cf := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-configmap",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"prometheus.yml": `scrape_configs:
      - job_name: 'best-prometheus-operator'
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - action: labelmap
          regex: __meta_kubernetes_pod_label_(.+)
        - source_labels: [__meta_kubernetes_namespace]
          action: replace
          target_label: kubernetes_namespace
        - source_labels: [__meta_kubernetes_pod_name]
          action: replace
          target_label: kubernetes_pod_name`, //yaml.Marshal(&cr.Spec.ScrapeConfigs),
		},
	}
	// Set Prometheus instance as the owner and controller
	ctrl.SetControllerReference(cr, cf, r.Scheme)
	return cf
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrometheusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Prometheus{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

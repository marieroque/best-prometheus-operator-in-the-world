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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Prometheus defines a Prometheus deployment.
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Prometheus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Prometheus cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec PrometheusSpec `json:"spec"`
	// Most recent observed status of the Prometheus cluster. Read-only.
	// More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status PrometheusStatus `json:"status,omitempty"`
}

// PrometheusList is a list of Prometheuses.
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
type PrometheusList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Prometheuses
	Items []*Prometheus `json:"items"`
}

// PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type PrometheusSpec struct {
	// Prometheus image version deployed
	// +kubebuilder:validation:Pattern=^[0-9]+\.[0-9]+\.[0-9]+$
	Version       *string         `json:"version"`
	ScrapeConfigs []*ScrapeConfig `json:"scrape_configs"`
}

// ScrapeConfig define a scrape configuration for the prometheus server
type ScrapeConfig struct {
	JobName        *string          `json:"job_name"`
	K8SSDConfigs   []*K8SSDConfig   `json:"kubernetes_sd_configs"`
	RelabelConfigs []*RelabelConfig `json:"relabel_configs,omitempty"`
}

// K8SSDConfig define a kubernetes service discovery config
type K8SSDConfig struct {
	// +kubebuilder:validation:Enum=node;pod;service;ingress
	Role *string `json:"role"`
}

type RelabelConfig struct {
	// +optional
	SourceLabels []*string `json:"source_labels,omitempty"`
	// +optional
	Action *string `json:"action,omitempty"`
	// +optional
	Regex *string `json:"regex,omitempty"`
	// +optional
	TargetLabel *string `json:"target_label,omitempty"`
}

// PrometheusStatus is the most recent observed status of the Prometheus cluster.
// More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type PrometheusStatus struct {
	//TODO
}

func init() {
	SchemeBuilder.Register(&Prometheus{}, &PrometheusList{})
}

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebsiteSpec defines the desired state of Website
type WebsiteSpec struct {
	// Container image to use for the website (defaults to nginx)
	// +optional
	Image string `json:"image,omitempty"`

	// Number of replicas for the website deployment
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas"`

	// The HTML content that will be served at "/"
	// +kubebuilder:validation:MinLength=1
	IndexHTML string `json:"indexHTML"`

	// Service type (ClusterIP, NodePort, or LoadBalancer)
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// +optional
	ServiceType string `json:"serviceType,omitempty"`
}

// WebsiteStatus defines the observed state of Website.
type WebsiteStatus struct {
	// Number of available replicas from the Deployment
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// The name of the Service created for this website
	ServiceName string `json:"serviceName,omitempty"`

	// The internal URL for accessing the website
	URL string `json:"url,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Website is the Schema for the websites API
type Website struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebsiteSpec   `json:"spec,omitempty"`
	Status WebsiteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebsiteList contains a list of Website
type WebsiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Website `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Website{}, &WebsiteList{})
}

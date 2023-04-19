/*
SPDX-License-Identifier: Apache-2.0

Copyright Contributors to the Submariner project.

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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make generate" to regenerate code after modifying this file

// ClusterSetNetworkPolicySpec defines the desired state of ClusterSetNetworkPolicy.
type ClusterSetNetworkPolicySpec struct {
	PolicyTypes []string                                `json:"policyTypes"`
	PodSelector metav1.LabelSelector                    `json:"podSelector"`
	Ingress     []networkingv1.NetworkPolicyIngressRule `json:"ingress"`
	Egress      []networkingv1.NetworkPolicyEgressRule  `json:"egress"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterSetNetworkPolicy is the Schema for the ClusterSetnetworkpolicies API.
type ClusterSetNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSetNetworkPolicySpec   `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterSetNetworkPolicyList contains a list of ClusterSetNetworkPolicy.
type ClusterSetNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSetNetworkPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterSetNetworkPolicy{}, &ClusterSetNetworkPolicyList{})
}

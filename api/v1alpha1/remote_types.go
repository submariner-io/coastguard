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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make generate" to regenerate code after modifying this file

// RemoteNamespaceSpec defines the desired state of RemoteNamespace.
type RemoteNamespaceSpec struct {
	ClusterID string `json:"clusterID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RemoteNamespace represents each tracked remote namespace
type RemoteNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RemoteNamespaceSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// RemoteNamespaceList contains a list of RemoteNamespace
type RemoteNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteNamespace `json:"items"`
}

// RemotePodSpec defines the desired state of RemotePod
type RemotePodSpec struct {
	ClusterID string   `json:"clusterID"`
	PodIPs    []string `json:"podIPs"`
	Namespace string   `json:"namespace"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RemotePod represents each tracked remote pod
type RemotePod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RemotePodSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// RemotePodList contains a list of RemotePod
type RemotePodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemotePod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteNamespace{}, &RemoteNamespaceList{}, &RemotePod{}, &RemotePodList{})
}

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

package networkpolicy

import (
	"fmt"
	"reflect"

	"github.com/submariner-io/coastguard/pkg/remotecluster"
	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
)

type RemoteNetworkPolicy struct {
	// Cluster is the origin of the network policy
	Cluster *remotecluster.RemoteCluster

	// Np is the NetworkPolicy origin of this tracking
	Np *v1net.NetworkPolicy

	// remotePods are the pods affected by this specific policy,
	// when changes are detected on such Pod
	remotePods map[string]*RemotePod

	// GeneratedPolicy is the generated network policy for the remote NetworkPolicy
	// if any policy is generated
	GeneratedPolicy *v1net.NetworkPolicy

	ObjID string
}

type RemotePod struct {
	cluster *remotecluster.RemoteCluster
	Pod     *v1.Pod
	ObjID   string
}

func NewRemotePod(pod *v1.Pod, remoteCluster *remotecluster.RemoteCluster, objID string) *RemotePod {
	return &RemotePod{
		cluster: remoteCluster,
		Pod:     pod,
		ObjID:   objID,
	}
}

func NewRemoteNetworkPolicy(np *v1net.NetworkPolicy, remoteCluster *remotecluster.RemoteCluster,
	objID string, existingPods map[string]*RemotePod,
) *RemoteNetworkPolicy {
	rnp := &RemoteNetworkPolicy{
		Cluster:    remoteCluster,
		Np:         np,
		remotePods: make(map[string]*RemotePod),
		ObjID:      objID,
	}

	for _, remotePod := range existingPods {
		rnp.processAddedPod(remotePod)
	}

	return rnp
}

func (rnp *RemoteNetworkPolicy) AddedPod(event *remotecluster.Event) {
	pod := event.Objs[0].(*v1.Pod)
	remotePod := NewRemotePod(pod, event.Cluster, event.ObjID)
	rnp.processAddedPod(remotePod)
}

func (rnp *RemoteNetworkPolicy) processAddedPod(remotePod *RemotePod) {
	if oldPod, exists := rnp.remotePods[remotePod.ObjID]; !exists {
		if rnp.ingressSelectsPod(remotePod.Pod, remotePod.cluster) {
			rnp.addRemotePod(remotePod)
		}
	} else {
		klog.Warningf("Somehow added Pod %s already was tracked by policy %s, updating", remotePod.ObjID, rnp.ObjID)
		event := remotePod.cluster.NewUpdateEvent(oldPod.Pod, remotePod.Pod)
		rnp.UpdatedPod(event)
	}
}

func (rnp *RemoteNetworkPolicy) UpdatedPod(event *remotecluster.Event) {
	if remotePod, exists := rnp.remotePods[event.ObjID]; exists {
		newPod := event.Objs[1].(*v1.Pod)
		if !reflect.DeepEqual(remotePod.Pod.ObjectMeta.Labels, newPod.ObjectMeta.Labels) {
			if !rnp.ingressSelectsPod(newPod, event.Cluster) {
				rnp.removeRemotePod(remotePod)
				return
			}
		}

		updatedRemotePod := NewRemotePod(newPod, event.Cluster, event.ObjID)
		rnp.updatedRemotePod(updatedRemotePod)
	} else {
		klog.Warningf("Received a Pod update event for a Pod we didn't know about %s", event.ObjID)
		rnp.AddedPod(event.ToAdded())
	}
}

func (rnp *RemoteNetworkPolicy) DeletedPod(event *remotecluster.Event) {
	if remotePod, exists := rnp.remotePods[event.ObjID]; exists {
		rnp.removeRemotePod(remotePod)
	} else {
		klog.Warningf("Received a Pod delete event for a Pod we didn't know about %s", event.ObjID)
	}
}

// selectsPod returs true or false, based on the network policy ingress selectors.
func (rnp *RemoteNetworkPolicy) ingressSelectsPod(pod *v1.Pod, remoteCluster *remotecluster.RemoteCluster) bool {
	// never select pods from it's own cluster, it's not our business
	if rnp.Cluster.ClusterID == remoteCluster.ClusterID {
		return false
	}

	for i := range rnp.Np.Spec.Ingress {
		if rnp.ingressRuleSelectsPod(&rnp.Np.Spec.Ingress[i], pod) {
			return true
		}
	}

	return false
}

func (rnp *RemoteNetworkPolicy) ingressRuleSelectsPod(rule *v1net.NetworkPolicyIngressRule, pod *v1.Pod) bool {
	for _, peer := range rule.From {
		if peer.PodSelector != nil && peer.NamespaceSelector == nil {
			return rnp.matchesPodSelector(peer.PodSelector, pod)
		} else if peer.NamespaceSelector != nil && peer.PodSelector == nil {
			if len(peer.NamespaceSelector.MatchLabels) == 0 && len(peer.NamespaceSelector.MatchExpressions) == 0 {
				return true
			}
			// TODO: Implement namespace selector
			klog.Error("Namespace selector still not fully handled")
		} else if peer.NamespaceSelector != nil && peer.PodSelector != nil {
			// TODO: Implement namespace and Pod selector combination
			klog.Error("Namespace selector + podSelector still not handled")
		}
	}

	return false
}

func (rnp *RemoteNetworkPolicy) matchesPodSelector(podSelector *metav1.LabelSelector, pod *v1.Pod) bool {
	if len(podSelector.MatchLabels) == 0 && len(podSelector.MatchExpressions) == 0 {
		// The PodSelector is empty, meaning it selects all pods in this namespace
		return pod.Namespace == rnp.Np.Namespace
	}
	// Verify if the Pod is in the same namespace as the policy, and then the podselector
	if pod.Namespace != rnp.Np.Namespace {
		return false
	}

	if sel, err := metav1.LabelSelectorAsSelector(podSelector); err == nil {
		return sel.Matches(labels.Set(pod.Labels))
	}

	klog.Errorf("error validating Np %s PodSelector %v", rnp.ObjID, podSelector)

	return false
}

func (rnp *RemoteNetworkPolicy) removeRemotePod(remotePod *RemotePod) {
	delete(rnp.remotePods, remotePod.ObjID)
	rnp.updateGeneratedPolicy()
}

func (rnp *RemoteNetworkPolicy) addRemotePod(remotePod *RemotePod) {
	rnp.remotePods[remotePod.ObjID] = remotePod
	rnp.updateGeneratedPolicy()
}

func (rnp *RemoteNetworkPolicy) updatedRemotePod(remotePod *RemotePod) {
	rnp.remotePods[remotePod.ObjID] = remotePod
	rnp.updateGeneratedPolicy()
}

// The originating network policy ID.
const coastGuardUIDLabel = "submariner-io/coastguard-Np-uid"

// The name of the originating NetworkPolicy.
const coastGuardNameLabel = "submariner-io/coastguard-Np"

// The name internal coastguard ID for the originating policy ID.
const coastGuardObjID = "submariner-io/coastguard-objid"

func IsGenerated(np *v1net.NetworkPolicy) bool {
	_, annotationExists := np.Annotations[coastGuardObjID]
	return annotationExists
}

func OriginatingObjID(np *v1net.NetworkPolicy) string {
	return np.Annotations[coastGuardObjID]
}

func (rnp *RemoteNetworkPolicy) updateGeneratedPolicy() {
	if len(rnp.remotePods) == 0 {
		rnp.GeneratedPolicy = nil
	} else {
		// make a copy so we maintain the same podSelector, etc...
		newPol := &v1net.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: rnp.Np.Namespace,
				Name:      generatePolicyName(rnp.Np),
				Annotations: map[string]string{
					coastGuardObjID: rnp.ObjID,
				},
				Labels: map[string]string{
					coastGuardNameLabel: rnp.Np.Name,
					coastGuardUIDLabel:  string(rnp.Np.UID),
				},
			},
			Spec: v1net.NetworkPolicySpec{
				PodSelector: rnp.Np.Spec.PodSelector,
				Ingress:     rnp.generateCIDRIngressRules(rnp.Np.Spec.Ingress),
			},
		}

		if len(newPol.Spec.Ingress) > 0 {
			if rnp.GeneratedPolicy != nil && ArePolicyRulesDifferent(rnp.GeneratedPolicy, newPol) ||
				rnp.GeneratedPolicy == nil {
				rnp.GeneratedPolicy = newPol
				klog.Infof("a new policy has been generated for %s", rnp.ObjID)
			}
		} else {
			if rnp.GeneratedPolicy != nil {
				klog.Infof("no matching pods on ingress rules for %s, no policy generated anymore", rnp.ObjID)
			}
			rnp.GeneratedPolicy = nil
		}
	}
}

func (rnp *RemoteNetworkPolicy) generateCIDRIngressRules(ingressRules []v1net.NetworkPolicyIngressRule) []v1net.NetworkPolicyIngressRule {
	newIngressRules := []v1net.NetworkPolicyIngressRule{}

	for i := range ingressRules {
		newRule := ingressRules[i].DeepCopy()
		newRule.From = rnp.buildPodPeersForIngressRule(&ingressRules[i])

		if len(newRule.From) > 0 {
			newIngressRules = append(newIngressRules, *newRule)
		}
	}

	return newIngressRules
}

func (rnp *RemoteNetworkPolicy) buildPodPeersForIngressRule(rule *v1net.NetworkPolicyIngressRule) []v1net.NetworkPolicyPeer {
	peers := []v1net.NetworkPolicyPeer{}

	for _, rp := range rnp.remotePods {
		if rnp.ingressRuleSelectsPod(rule, rp.Pod) && rp.Pod.Status.PodIP != "" {
			// NOTE: this can be optimized in a future by aggregatting multiple pods over CIDRs
			peers = append(peers, v1net.NetworkPolicyPeer{IPBlock: &v1net.IPBlock{CIDR: rp.Pod.Status.PodIP + "/32"}})
		}
	}

	return peers
}

func generatePolicyName(np *v1net.NetworkPolicy) string {
	return fmt.Sprintf("coastguard-%s", np.UID)
}

func (rnp *RemoteNetworkPolicy) GeneratedPolicyName() string {
	return generatePolicyName(rnp.Np)
}

func ArePolicyRulesDifferent(actualNp, expectedNp *v1net.NetworkPolicy) bool {
	return !reflect.DeepEqual(actualNp.Spec.PodSelector, expectedNp.Spec.PodSelector) ||
		!reflect.DeepEqual(actualNp.Spec.Ingress, expectedNp.Spec.Ingress)
}

package networkpolicy

import (
	"encoding/json"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"

	"github.com/submariner-io/coastguard/pkg/remotecluster"
)

type RemoteNetworkPolicy struct {
	//cluster is the origin of the network policy
	cluster *remotecluster.RemoteCluster

	//Np is the NetworkPolicy origin of this tracking
	Np *v1net.NetworkPolicy

	//remotePods are the pods affected by this specific policy,
	//when changes are detected on such Pod
	remotePods map[string]*RemotePod

	//genPol it's the generated network policy for the remote NetworkPolicy
	//if any policy is generated
	genPol *v1net.NetworkPolicy

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
	objID string, existingPods map[string]*RemotePod) *RemoteNetworkPolicy {

	rnp := &RemoteNetworkPolicy{
		cluster:    remoteCluster,
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

// selectsPod returs true or false, based on the network policy ingress selectors
func (rnp *RemoteNetworkPolicy) ingressSelectsPod(pod *v1.Pod, remoteCluster *remotecluster.RemoteCluster) bool {

	// never select pods from it's own cluster, it's not our business
	if rnp.cluster.ClusterID == remoteCluster.ClusterID {
		return false
	}

	for _, rule := range rnp.Np.Spec.Ingress {
		if rnp.ingressRuleSelectsPod(&rule, pod) {
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
			} else {
				//TODO: Implement namespace selector
				klog.Error("Namespace selector still not fully handled")
			}
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
	} else {
		// Verify if the Pod is in the same namespace as the policy, and then the podselector
		if pod.Namespace != rnp.Np.Namespace {
			return false
		}
		if sel, err := metav1.LabelSelectorAsSelector(podSelector); err == nil {
			return sel.Matches(labels.Set(pod.Labels))
		} else {
			klog.Errorf("error validating Np %s PodSelector %v", rnp.ObjID, podSelector)
			return false
		}
	}
}

// removeRemotePod
func (rnp *RemoteNetworkPolicy) removeRemotePod(remotePod *RemotePod) {
	delete(rnp.remotePods, remotePod.ObjID)
	rnp.updateGeneratedPolicy()
}

// addRemotePod
func (rnp *RemoteNetworkPolicy) addRemotePod(remotePod *RemotePod) {
	rnp.remotePods[remotePod.ObjID] = remotePod
	rnp.updateGeneratedPolicy()
}

// updateRemotePod
func (rnp *RemoteNetworkPolicy) updatedRemotePod(remotePod *RemotePod) {
	rnp.remotePods[remotePod.ObjID] = remotePod
	rnp.updateGeneratedPolicy()
}

const coastGuardUIDLabel = "submariner-io/coastguard-Np-uid"
const coastGuardNameLabel = "submariner-io/coastguard-Np"

func (rnp *RemoteNetworkPolicy) updateGeneratedPolicy() {

	if len(rnp.remotePods) == 0 {
		rnp.genPol = nil
	} else {
		// make a copy so we maintain the same podSelector, etc...
		newPol := &v1net.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: rnp.Np.Namespace,
				Name:      generatePolicyName(rnp.Np),
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
			rnp.genPol = newPol
			newPolStr, _ := json.MarshalIndent(newPol, "", "\t")
			klog.Infof("a new policy has been generated for %s:\n%s", rnp.ObjID, newPolStr)
		} else {
			rnp.genPol = nil
			klog.Infof("no matching pods on ingress rules for %s, no policy generated anymore", rnp.ObjID)
		}
	}
}

func (rnp *RemoteNetworkPolicy) generateCIDRIngressRules(ingressRules []v1net.NetworkPolicyIngressRule) []v1net.NetworkPolicyIngressRule {
	newIngressRules := []v1net.NetworkPolicyIngressRule{}
	for _, rule := range ingressRules {
		newRule := rule.DeepCopy()
		newRule.From = rnp.buildPodPeersForIngressRule(&rule)
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
			//NOTE: this can be optimized in a future by aggregatting multiple pods over CIDRs
			podCIDR := rp.Pod.Status.PodIP + "/32"
			peers = append(peers, v1net.NetworkPolicyPeer{IPBlock: &v1net.IPBlock{CIDR: podCIDR}})
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

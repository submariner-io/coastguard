package scenarios

import (
	. "github.com/onsi/ginkgo"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	framework "github.com/submariner-io/submariner/test/e2e/framework"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var f = framework.NewDefaultFramework("podselectornp")

var _ = PDescribe("[Dataplane] PodSelector based Network Policies for ingress between clusters", func() {
	PContext("Allow communication between selected pods on different clusters in same project", func() {
		PIt("Should allow communication between selected pods", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			// create pod1 with label "role = backend" in "boring namespace"
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 1")
			// create pod2 with label "role = frontend" in "boring namespace"
			By("creating network policy in cluster 1 that allows communication to pod 1 from pod 2 in cluster 2")
			_,err := f.ClusterClients[framework.ClusterA].NetworkingV1().NetworkPolicies(f.Namespace).Create(podSelectorNp)
			framework.ExpectNoError(err)
			By("testing connectivity between the two pods")
			// connectivity between pod 1 and pod2
			// Negative tests
			By("creating connecting pod 3 with label 3 in cluster 2 in namespace 1")
			// create pod 3 with label "role = another" in "boring namespace"
			By("testing non connectivity between the two pods")
			// Tests no connectivity between pod 3 and pod 1
		})
	})

	PContext("[Dataplane] Allow communication between all pods on first cluster to selected pods in second cluster in same project", func() {
		PIt("Should allow communication between selected pods on 1 cluster and all pods on other cluster", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			// create pod1 with label "role = backend" in "boring namespace"
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 1")
			// create pod2 with label "role = frontend" in "boring namespace"
			By("creating network policy in cluster 1 that allows communication to pod 1 from all pods in cluster 2")
			_, err := f.ClusterClients[framework.ClusterA].NetworkingV1().NetworkPolicies(f.Namespace).Create(allPodSelectorNp)
			framework.ExpectNoError(err)
			By("testing connectivity between the two pods")
			// connectivity between pod 1 and pod2
			By("creating listener pod 3 with label 2 in cluster 1 in namespace 1")
			// create pod 3 with label "role = another" in "boring namespace"
			By("creating connecting pod 4 with label 2 in cluster 2 in namespace 1")
			// create pod 4 with label "role = frontend" in "boring namespace"
			By("testing connectivity between the two pods")
			// connectivity between pod 3 and pod 4
		})
	})
})

var _ = PDescribe("[Ctlplane] PodSelector based Network Policies for ingress between clusters", func() {
	PContext("Allow communication between selected pods on different clusters in same project", func() {
		PIt("Should allow communication between selected pods", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 2 with label 2 in cluster 2 in namespace 1")
			By("creating network policy in cluster 1 that allows communication to pod 1 from pod 2 in cluster 2")
			By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 2 IP in ipBlocks")
		})
	})

	PContext("[Ctlplane] Allow communication between selected pods on first cluster to all pods in second cluster in same project", func() {
		PIt("Should allow communication between selected pods on 1 cluster and all pods on other cluster", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 2 with label 2 in cluster 2 in namespace 1")
			By("creating network policy in cluster 1 that allows communication to pod 1 from pod 2 in cluster 2")
			By("Waiting for a NetworkPolicy to appear in cluster 1 containing all pods with label 2 IP in cluster 2 in ipBlocks")
		})
	})
})

var podSelectorNp = &networkingv1.NetworkPolicy {
	ObjectMeta: metav1.ObjectMeta {
		Name: "pod1-pod2-same-ns-different-labels",
		Namespace: "boring",
	},
	Spec: networkingv1.NetworkPolicySpec {
		PodSelector: metav1.LabelSelector {
			MatchLabels: map[string]string {
				"role": "backend",
			},
		},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		Ingress: []networkingv1.NetworkPolicyIngressRule {{
			From: []networkingv1.NetworkPolicyPeer {{
				PodSelector: &metav1.LabelSelector {
					MatchLabels: map[string]string {
						"role": "frontend",
					},
				},
			}},
			Ports: []networkingv1.NetworkPolicyPort {{
				//Protocol: TCP is the default protocol. Need not mention it
				Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 1234},
				// '1234' value comes from the https://github.com/submariner-io/submariner/blob/master/test/e2e/framework/network_pods.go#L13
			}},
		}},
	},
}

var allPodSelectorNp = &networkingv1.NetworkPolicy {
	ObjectMeta: metav1.ObjectMeta {
		Name: "all-pods-same-ns-different-labels",
		Namespace: "boring",
		ClusterName: string(framework.ClusterA),
	},
	Spec: networkingv1.NetworkPolicySpec {
		PodSelector: metav1.LabelSelector {
			// all pods on cluster A are selected
		},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		Ingress: []networkingv1.NetworkPolicyIngressRule {{
			From: []networkingv1.NetworkPolicyPeer {{
				PodSelector: &metav1.LabelSelector {
					MatchLabels: map[string]string {
						"role": "frontend",
					},
				},
			}},
			Ports: []networkingv1.NetworkPolicyPort {{
				Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 1234},
			}},
		}},
	},
}

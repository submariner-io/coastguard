package scenarios

import (
	. "github.com/onsi/ginkgo"
	"github.com/submariner-io/submariner/test/e2e/framework"
	"k8s.io/apimachinery/pkg/util/intstr"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = PDescribe("[Dataplane] Pod and Namespace selector based Network Policies for ingress between clusters", func() {
	f := framework.NewDefaultFramework("np-pod-namespace")
	PIt("Should allow communication between selected pods in selected namespace", func() {
		By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
		By("creating connecting pod 2 with label 2 in cluster 2 in namespace 2")
		By("creating listener pod 3 with label 1 in cluster 1 in namespace 1")
		By("creating connecting pod 4 with label 1 in cluster 2 in namespace 1")
		By("creating network policy in cluster 1 that allows communication to pod 1 from any pod with label 2 and in namespace 2 in cluster 2")
		By("testing connectivity between pods 1 and 2")
		By("testing non connectivity between pods 3 and 4")
	})
})

var _ = PDescribe("[Ctlplane] Pod and Namespace selector based Network Policies for ingress between clusters", func() {
	PIt("Should allow communication between selected pods in selected namespace", func() {
		By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
		By("creating listener pod 2 with label 2 in cluster 2 in namespace 2")
		By("creating listener pod 3 with label 2 in cluster 2 in namespace 2")
		By("creating network policy in cluster 1 that allows communication to pod 1 from any pod with label 2 and in namespace 2 in cluster 2")
		By("Waiting for a NetworkPolicy to appear in cluster 1 containing IPs of pods with label 2 and in namespace 2 in cluster 2 in ipBlocks")
	})
})

var podNamespaceSelectorNp = &networkingv1.NetworkPolicy {
	ObjectMeta: metav1.ObjectMeta {
		Name: "pod-namespace-selector-policy",
		Namespace: "fancy",
		ClusterName: string(framework.ClusterA),
	},
	Spec: networkingv1.NetworkPolicySpec {
		PodSelector: metav1.LabelSelector {
			MatchLabels: map[string]string {
				"role": "frontend",
			},
		},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		Ingress: []networkingv1.NetworkPolicyIngressRule {{
			From: []networkingv1.NetworkPolicyPeer {{
				NamespaceSelector: &metav1.LabelSelector {
					MatchLabels: map[string]string {
						"project": "fancy",
					},
				},
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string {
						"role": "backend",
					},
				},
			}},
			Ports: []networkingv1.NetworkPolicyPort {{
				//Protocol: &protocolTCP,
				Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 1234},
			}},
		}},
	},
}
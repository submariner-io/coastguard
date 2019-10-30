package scenarios

import (
	. "github.com/onsi/ginkgo"
	"github.com/submariner-io/submariner/test/e2e/framework"
	"k8s.io/apimachinery/pkg/util/intstr"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = PDescribe("[Dataplane] Namespace Selector based Network Policies for ingress between clusters", func() {

	f := framework.NewDefaultFramework("namespaceselectornp")

	//f.GenerateNamespace("fancy", map["project"]["fancy"])

	PIt("Should allow communication between pods in selected namespace", func() {
		By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
		//pod1 in namespace "project:fancy" with label "role:frontend"
		By("creating connecting pod 2 with label 2 in cluster 2 in namespace 2")
		//pod2 in namespace "project:fancy" with label "role:backend"
		By("creating network policy in cluster 1 that allows communication to pod 1 from any pod in namespace 2 in cluster 2")
		_,err := f.ClusterClients[framework.ClusterA].NetworkingV1().NetworkPolicies("fancy").Create(namespaceSelectorNp)
		framework.ExpectNoError(err)
		By("testing connectivity between the two pods")
	})
})

var _ = PDescribe("[Ctlplane] Namespace Selector based Network Policies for ingress between clusters", func() {
	PIt("Should allow communication between pods in selected namespace", func() {
		By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
		By("creating listener pod 2 with label 2 in cluster 2 in namespace 2")
		By("creating network policy in cluster 1 that allows communication to pod 1 from any pod in namespace 2 in cluster 2")
		_,err := f.ClusterClients[framework.ClusterA].NetworkingV1().NetworkPolicies("fancy").Create(namespaceSelectorNp)
		framework.ExpectNoError(err)
		By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 2 IP in ipBlocks")
	})
})

var namespaceSelectorNp = &networkingv1.NetworkPolicy {
	ObjectMeta: metav1.ObjectMeta {
		Name: "pod1-pod2-different-ns",
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
			}},
			Ports: []networkingv1.NetworkPolicyPort {{
				//Protocol: &protocolTCP,
				Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 1234},
			}},
		}},
	},
}
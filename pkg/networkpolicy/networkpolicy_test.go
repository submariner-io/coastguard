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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/submariner-io/coastguard/pkg/remotecluster"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
)

var _ = Describe("RemoteNetworkPolicies", func() {
	klog.InitFlags(nil)

	Describe("Event handing", describeEventHandling)
	Describe("NetworkPolicy processing", describeRuleMatching)
})

func testPodName(clusterIndex, podIdx int, label string, namespaceIndex int) string {
	return fmt.Sprintf("c%dpod%d-%s-namespace%d", clusterIndex, podIdx, label, namespaceIndex)
}

func testPodIP(clusterIndex, labelIndex, namespaceIndex int) string {
	return fmt.Sprintf("%d.%d.1.%d", clusterIndex, labelIndex, namespaceIndex)
}

func describeRuleMatching() {
	var (
		rnp         *RemoteNetworkPolicy
		clusters    []*remotecluster.RemoteCluster
		clusterPods [][]*v1.Pod
	)

	BeforeEach(func() {
		var cluster1 *remotecluster.RemoteCluster

		rnp, cluster1 = newDefaultRemotePolicyAndCluster()
		clusters = []*remotecluster.RemoteCluster{
			cluster1,
			remotecluster.New(clusterID2, fake.NewSimpleClientset()),
			remotecluster.New(clusterID3, fake.NewSimpleClientset()),
		}

		// create pods: clusters * namespaces * labels pods
		//
		//                cluster1                            cluster2                             cluster3
		//                ~~~~~~~~                            ~~~~~~~~                             ~~~~~~~~
		// namespace1     c1pod1-selected-pods-namespace1     c2pod1-selected-pods-namespace1      c3pod1-selected-pods-namespace1
		//                   1.1.1.1                              2.1.1.1                              3.1.1.1
		//                c1pod2-non-selected-pods-namespace1 c2pod2-non-selected-pods-namespace1  c3pod2-non-selected-pods-namespace1
		//                   1.2.1.1                              2.2.1.1                              3.2.1.1
		//                c1pod3-other-pods-namespace1        c2pod3-other-pods-namespace1         c3pod3-other-pods-namespace1
		//                   1.3.1.1                              1.3.1.1                              3.3.1.1
		//
		// namespace2     c1pod4-selected-pods-namespace2     c2pod4-selected-pods-namespace2      c3pod4-selected-pods-namespace2
		//                   1.1.1.2                              2.1.1.2                              3.1.1.2
		//                c1pod5-non-selected-pods-namespace2 c2pod5-non-selected-pods-namespace2  c3pod5-non-selected-pods-namespace2
		//                   1.2.1.2                              2.2.1.2                              3.2.1.2
		//                c1pod6-other-pods-namespace2        c2pod6-other-pods-namespace2         c3pod6-other-pods-namespace2
		//                   1.3.1.2                              2.3.1.2                              3.3.1.2
		//
		// namespace3     c1pod7-selected-pods-namespace3     c2pod7-selected-pods-namespace3      c3pod7-selected-pods-namespace3
		//                   1.1.1.3                              2.1.1.3                              3.1.1.3
		//                c1pod8-non-selected-pods-namespace3 c2pod8-non-selected-pods-namespace3  c3pod8-non-selected-pods-namespace3
		//                   1.2.1.3                              2.2.1.3                              3.2.1.3
		//                c1pod9-other-pods-namespace3        c2pod9-other-pods-namespace3         c3pod9-other-pods-namespace3
		//                   1.3.1.3                              2.3.1.3                              3.3.1.3

		clusterPods = createTestPodMatrix(clusters)
	})

	When("Translating policies", func() {
		It("Should copy the Port definitions", func() {
			addAllPods(rnp, clusters, clusterPods)
			Expect(rnp.GeneratedPolicy.Spec.Ingress).To(HaveLen(1))
			ingressPolicy := rnp.GeneratedPolicy.Spec.Ingress[0]
			Expect(ingressPolicy.Ports).To(HaveLen(1))
			Expect(ingressPolicy.Ports[0].Port.IntVal).To(BeIdenticalTo(int32(testPort)))
		})
	})

	When("Policies have one rule", func() {
		It("Should convert ingress selectors to pod IPs", func() {
			addAllPods(rnp, clusters, clusterPods)
			Expect(rnp.GeneratedPolicy.Spec.Ingress).To(HaveLen(1))
			ingressPolicy := rnp.GeneratedPolicy.Spec.Ingress[0]

			// This policy lives on cluster1, and should select pods with
			// selected-pods label, from cluster2 and cluster3, namespace1
			cidrs := getCIDRsFromPeers(ingressPolicy.From)
			Expect(cidrs).To(HaveLen(2))

			for _, ip := range []string{"2.1.1.1", "3.1.1.1"} {
				ipCidr := fmt.Sprintf("%s/32", ip)
				Expect(cidrs).To(ContainElement(ipCidr))
			}
		})
	})

	When("Policies have several rules with different matchings", func() {
		It("Should convert ingress selectors to pod IPs", func() {
			rnp.Np.Spec.Ingress = append(rnp.Np.Spec.Ingress,
				networkingv1.NetworkPolicyIngressRule{
					From:  []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"pods": testOtherPods}}}},
					Ports: []networkingv1.NetworkPolicyPort{{Port: &intstr.IntOrString{IntVal: testPort443}}},
				})
			addAllPods(rnp, clusters, clusterPods)

			Expect(rnp.GeneratedPolicy.Spec.Ingress).To(HaveLen(2))

			By("Verifying the 2.1.1.1 and 3.1.1.1 IPs are on ingress rule one")
			CheckCIDRs(rnp.GeneratedPolicy.Spec.Ingress[0].From, []string{"2.1.1.1", "3.1.1.1"})

			By("Verifying the 2.3.1.1 and 3.3.1.1 IPs are on ingress rule two")
			CheckCIDRs(rnp.GeneratedPolicy.Spec.Ingress[1].From, []string{"2.3.1.1", "3.3.1.1"})
		})
	})

	When("Policies have pod matchings and ipBlocks", func() {
		It("Should convert ingress selectors to pod IPs and ignore the initial ip-block", func() {
			rnp.Np.Spec.Ingress = append(rnp.Np.Spec.Ingress,
				networkingv1.NetworkPolicyIngressRule{
					From:  []networkingv1.NetworkPolicyPeer{{IPBlock: &networkingv1.IPBlock{CIDR: "8.8.8.8/32"}}},
					Ports: []networkingv1.NetworkPolicyPort{{Port: &intstr.IntOrString{IntVal: testPort443}}},
				})
			addAllPods(rnp, clusters, clusterPods)
			By("checking that the generated policy only has one ingress rule, removing the original np IPBlock")
			Expect(rnp.GeneratedPolicy.Spec.Ingress).To(HaveLen(1))

			CheckCIDRs(rnp.GeneratedPolicy.Spec.Ingress[0].From, []string{"2.1.1.1", "3.1.1.1"})
		})
	})

	When("Ingress rules have no matching pods", func() {
		It("Should not generate policies", func() {
			rnp.Np.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{{
				From:  []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"pods": "I-DONT-MATCH"}}}},
				Ports: []networkingv1.NetworkPolicyPort{{Port: &intstr.IntOrString{IntVal: testPort443}}},
			}}
			addAllPods(rnp, clusters, clusterPods)
			Expect(rnp.GeneratedPolicy).To(BeNil())
		})
	})

	When("Ingress rules have ipBlocks and no matching pods", func() {
		It("Should not generate policies", func() {
			rnp.Np.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{{
				From:  []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"pods": "I-DONT-MATCH"}}}},
				Ports: []networkingv1.NetworkPolicyPort{{Port: &intstr.IntOrString{IntVal: testPort443}}},
			}, {
				From:  []networkingv1.NetworkPolicyPeer{{IPBlock: &networkingv1.IPBlock{CIDR: "8.8.8.8/32"}}},
				Ports: []networkingv1.NetworkPolicyPort{{Port: &intstr.IntOrString{IntVal: testPort443}}},
			}}
			addAllPods(rnp, clusters, clusterPods)
			Expect(rnp.GeneratedPolicy).To(BeNil())
		})
	})

	When("Ingress rules have namespaceSelectors to select all pods", func() {
		It("Should generate a policy that includes ipBlocks for all pods in all the other clusters", func() {
			By("Creating an ingress rule that selects all pods on all namespaces")
			rnp.Np.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{
				{
					From:  []networkingv1.NetworkPolicyPeer{{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{}}}},
					Ports: []networkingv1.NetworkPolicyPort{{Port: &intstr.IntOrString{IntVal: testPort}}},
				},
			}
			addAllPods(rnp, clusters, clusterPods)

			By("By checking that the CIDRs in ipblocks match just and only the cluster2-3 pods")

			ips := collectPodIPs(clusterPods[1:])

			CheckCIDRs(rnp.GeneratedPolicy.Spec.Ingress[0].From, ips)
		})
	})
}

func collectPodIPs(clusterPods [][]*v1.Pod) []string {
	ips := []string{}

	for _, pods := range clusterPods {
		for _, pod := range pods {
			ips = append(ips, pod.Status.PodIP)
		}
	}

	return ips
}

func describeEventHandling() {
	var (
		rnp                *RemoteNetworkPolicy
		cluster1, cluster2 *remotecluster.RemoteCluster
	)

	BeforeEach(func() {
		rnp, cluster1 = newDefaultRemotePolicyAndCluster()
		cluster2 = remotecluster.New(clusterID2, fake.NewSimpleClientset())
	})

	When("Adding pods which are not selected by policy", func() {
		It("Should not keep a reference", func() {
			pod := newPod(testPod1, testNamespace, testNonSelectedPods, testPodIP1)
			addEvent := cluster2.NewAddEvent(pod)
			rnp.AddedPod(addEvent)
			Expect(rnp.remotePods).ShouldNot(HaveKey(addEvent.ObjID))
		})
	})

	When("Updating pods so they are not selected by policy", func() {
		It("Should not keep a reference", func() {
			pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			addEvent := cluster2.NewAddEvent(pod)
			rnp.AddedPod(addEvent)
			updatedPod := newPod(testPod1, testNamespace, testNonSelectedPods, testPodIP1)
			rnp.UpdatedPod(cluster2.NewUpdateEvent(pod, updatedPod))
			Expect(rnp.remotePods).ShouldNot(HaveKey(addEvent.ObjID))
		})
	})
	When("Deleting pods previously selected by a policy", func() {
		It("Should not keep a reference", func() {
			pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			addEvent := cluster2.NewAddEvent(pod)
			rnp.AddedPod(addEvent)
			rnp.DeletedPod(cluster2.NewDeleteEvent(pod))
			Expect(rnp.remotePods).ShouldNot(HaveKey(addEvent.ObjID))
		})
	})

	When("Adding pods", func() {
		It("Should keep a reference to pods which are selected by policy", func() {
			pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			addEvent := cluster2.NewAddEvent(pod)
			rnp.AddedPod(addEvent)
			Expect(rnp.remotePods).Should(HaveKey(addEvent.ObjID))
		})
	})
	When("Updating pods so that they contain the selected labels", func() {
		It("Should keep a reference to pods which are selected by policy", func() {
			pod := newPod(testPod1, testNamespace, testNonSelectedPods, testPodIP1)
			addEvent := cluster2.NewAddEvent(pod)
			rnp.AddedPod(addEvent)
			updatedPod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			rnp.UpdatedPod(cluster2.NewUpdateEvent(pod, updatedPod))
			Expect(rnp.remotePods).Should(HaveKey(addEvent.ObjID))
		})
	})

	When("Adding a pod twice", func() {
		It("Should not crash", func() {
			pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			addEvent := cluster2.NewAddEvent(pod)
			rnp.AddedPod(addEvent)
			rnp.AddedPod(addEvent)
			Expect(rnp.remotePods).Should(HaveKey(addEvent.ObjID))
		})
	})

	When("Updating a pod which didn't exist", func() {
		It("Should not crash", func() {
			pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			updateEvent := cluster2.NewUpdateEvent(pod, pod)
			rnp.UpdatedPod(updateEvent)
			Expect(rnp.remotePods).Should(HaveKey(updateEvent.ObjID))
		})
	})

	When("Deleting a pod twice", func() {
		It("Should not crash", func() {
			pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			addEvent := cluster2.NewAddEvent(pod)
			rnp.AddedPod(addEvent)
			rnp.DeletedPod(cluster2.NewDeleteEvent(pod))
			rnp.DeletedPod(cluster2.NewDeleteEvent(pod))
			Expect(rnp.remotePods).ShouldNot(HaveKey(addEvent.ObjID))
		})
	})

	When("Deleting a pod which we didn't track yet", func() {
		It("Should not crash", func() {
			pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
			deleteEvent := cluster2.NewDeleteEvent(pod)
			rnp.DeletedPod(deleteEvent)
			Expect(rnp.remotePods).ShouldNot(HaveKey(deleteEvent.ObjID))
		})
	})

	It("Should not select pods on the same cluster the NetworkPolicy lives", func() {
		pod := newPod(testPod1, testNamespace, testSelectedPods, testPodIP1)
		addEvent := cluster1.NewAddEvent(pod)
		rnp.AddedPod(addEvent)
		Expect(rnp.remotePods).ShouldNot(HaveKey(addEvent.ObjID))
	})
}

func newDefaultRemotePolicyAndCluster() (*RemoteNetworkPolicy, *remotecluster.RemoteCluster) {
	return createRemotePolicyAndCluster(testAppliedPods, testSelectedPods, testNamespace, clusterID1)
}

func createRemotePolicyAndCluster(selectedPods, ingressPods, namespace,
	clusterID string,
) (*RemoteNetworkPolicy, *remotecluster.RemoteCluster) {
	np := createPodSelectorNetworkPolicy(selectedPods, ingressPods, namespace)
	rc1 := remotecluster.New(clusterID, fake.NewSimpleClientset())
	rp := NewRemoteNetworkPolicy(np, rc1, remotecluster.ObjID(np.Namespace, np.Name, rc1.ClusterID, np.UID), nil)

	return rp, rc1
}

const (
	testPort    = 80
	testPort443 = 443
	testPod1    = "test-pod1"
	testPodIP1  = "1.1.1.1"

	testAppliedPods     = "applied-pods"
	testSelectedPods    = "selected-pods"
	testOtherPods       = "other-pods"
	testNonSelectedPods = "noningress-pods"
	testNamespace       = "namespace1"
	clusterID1          = "cluster-1"
	clusterID3          = "cluster-3"
	clusterID2          = "cluster-2"
	testPodNamespaces   = 3
)

func createPodSelectorNetworkPolicy(appliedPods, selectedPods, namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"pods": appliedPods,
				},
			},
			// Allow traffic only from client-a
			Ingress: []networkingv1.NetworkPolicyIngressRule{{
				From: []networkingv1.NetworkPolicyPeer{{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"pods": selectedPods,
						},
					},
				}},
				Ports: []networkingv1.NetworkPolicyPort{{
					Port: &intstr.IntOrString{IntVal: testPort},
				}},
			}},
		},
	}
}

/*
func createEmptyNamespaceSelector(appliedPods string, namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"pods": appliedPods,
				},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{{
				From: []networkingv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				}},
				Ports: []networkingv1.NetworkPolicyPort{{
					Port: &intstr.IntOrString{IntVal: testPort},
				}},
			}},
		},
	}
}
*/

func CheckCIDRs(peers []networkingv1.NetworkPolicyPeer, ips []string) {
	cidrs := getCIDRsFromPeers(peers)
	Expect(cidrs).To(HaveLen(len(ips)))

	for _, ip := range ips {
		ipCidr := fmt.Sprintf("%s/32", ip)
		Expect(cidrs).To(ContainElement(ipCidr))
	}
}

func createTestPodMatrix(clusters []*remotecluster.RemoteCluster) [][]*v1.Pod {
	clusterPods := make([][]*v1.Pod, len(clusters))

	labels := []string{testSelectedPods, testNonSelectedPods, testOtherPods}

	for clusterIdx := range clusters {
		podIdx := 1

		for namespaceIdx := 1; namespaceIdx <= testPodNamespaces; namespaceIdx++ {
			for labelIdx, label := range labels {
				namespace := fmt.Sprintf("namespace%d", namespaceIdx)

				pod := newPod(testPodName(clusterIdx+1, podIdx, label, namespaceIdx),
					namespace, label, testPodIP(clusterIdx+1, labelIdx+1, namespaceIdx))

				clusterPods[clusterIdx] = append(clusterPods[clusterIdx], pod)

				podIdx++
			}
		}
	}

	return clusterPods
}

func getCIDRsFromPeers(peers []networkingv1.NetworkPolicyPeer) []string {
	cidrs := []string{}

	for _, peer := range peers {
		cidrs = append(cidrs, peer.IPBlock.CIDR)
	}

	return cidrs
}

func addAllPods(policy *RemoteNetworkPolicy, clusters []*remotecluster.RemoteCluster, clusterPods [][]*v1.Pod) {
	for i, cluster := range clusters {
		for _, pod := range clusterPods[i] {
			policy.AddedPod(cluster.NewAddEvent(pod))
		}
	}
}

func newPod(name, namespace, podLabel, ip string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    map[string]string{"pods": podLabel},
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
			PodIP: ip,
		},
	}
}

func TestNetworkPolicies(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coastguard: RemoteNetworkPolicy suite")
}

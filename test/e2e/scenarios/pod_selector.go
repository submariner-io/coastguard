package scenarios

import (
	. "github.com/onsi/ginkgo"
)

var _ = PDescribe("[Dataplane] PodSelector based Network Policies for ingress between clusters", func() {
	PContext("Allow communication between selected pods on different clusters in same project", func() {
		PIt("Should allow communication between selected pods", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 1")
			By("creating network policy in cluster 1 that allows communication to pod 1 from pod 2 in cluster 2")
			By("testing connectivity between the two pods")
		})
	})

	PContext("[Dataplane] Allow communication between selected pods on first cluster to all pods in second cluster in same project", func() {
		PIt("Should allow communication between selected pods on 1 cluster and all pods on other cluster", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 1")
			By("creating network policy in cluster 1 that allows communication to pod 1 from all pods in cluster 2")
			By("testing connectivity between the two pods")
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

package scenarios

import (
	. "github.com/onsi/ginkgo"
)

var _ = PDescribe("[Dataplane] Pod and Namespace selector based Network Policies for ingress between clusters", func() {
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

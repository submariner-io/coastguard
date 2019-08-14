package scenarios

import (
	. "github.com/onsi/ginkgo"
)

var _ = PDescribe("[Dataplane] Namespace Selector based Network Policies for ingress between clusters", func() {
	PIt("Should allow communication between pods in selected namespace", func() {
		By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
		By("creating connecting pod 2 with label 2 in cluster 2 in namespace 2")
		By("creating network policy in cluster 1 that allows communication to pod 1 from any pod in namespace 2 in cluster 2")
		By("testing connectivity between the two pods")
	})
})

var _ = PDescribe("[Ctlplane] Namespace Selector based Network Policies for ingress between clusters", func() {
	PIt("Should allow communication between pods in selected namespace", func() {
		By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
		By("creating listener pod 2 with label 2 in cluster 2 in namespace 2")
		By("creating network policy in cluster 1 that allows communication to pod 1 from any pod in namespace 2 in cluster 2")
		By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 2 IP in ipBlocks")
	})
})

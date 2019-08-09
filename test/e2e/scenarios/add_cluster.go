package scenarios

import (
	. "github.com/onsi/ginkgo"
)

var _ = PDescribe("[Dataplane] Adding existing network policy to newly added cluster", func() {
	PContext("Registering cluster 3 with kubefed and then creating a pod", func() {
		PIt("Should implement existing namespace selector based network policy in newly added cluster", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 4 with label 1 in cluster 1 in namespace 1")
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 2")
			By("creating namespace selector based network policy on cluster 1")
			By("testing connectivity between pods")
			By("Registering cluster 3 to kubefed")
			By("creating connecting pod 3 with label 3 in cluster 3 in namespace 2")
			By("testing connectivity between pod 1 and 2")
			By("testing connectivity between pod 3 and 4")
			By("Unregistering cluster 3")
		})
	})

	PContext("[Dataplane] Creating a pod in cluster 3 and then registering it with kubefed", func() {
		PIt("Should implement existing namespace selector based network policy in newly added cluster", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 4 with label 1 in cluster 1 in namespace 1")
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 2")
			By("creating namespace selector based network policy on cluster 1")
			By("testing connectivity between pod 1 and 2")
			By("creating connecting pod 3 with label 3 in cluster 3 in namespace 2")
			By("testing non connectivity between pod 3 and 4")
			By("Registering cluster 3 to kubefed")
			By("testing connectivity between pod 3 and 4")
			By("Unregistering cluster 3")
		})
	})
})

var _ = PDescribe("[Ctlplane] Adding existing network policy to newly added cluster", func() {
	PContext("Registering cluster 3 with kubefed and then creating a pod", func() {
		PIt("Should implement existing namespace selector based network policy in newly added cluster", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 2 with label 2 in cluster 2 in namespace 2")
			By("creating network policy with namespace selector on cluster 1")
			By("Waiting for a NetworkPolicy to appear in cluster1 containing pod 2 IP in ipBlocks")
			By("Registering cluster3 to kubefed")
			By("creating listener pod 3 with label 3 in cluster 3 in namespace 2")
			By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 3 IP in ipBlocks")
			By("Unregistering cluster3")
		})
	})

	PContext("[Ctlplane] Creating a pod in cluster 3 and then registering it with kubefed", func() {
		PIt("Should implement existing namespace selector based network policy in newly added cluster", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 2 with label 2 in cluster 2 in namespace 2")
			By("creating network policy with namespace selector on cluster 1")
			By("Waiting for a NetworkPolicy to appear in cluster1 containing pod 2 IP in ipBlocks")
			By("creating listener pod 3 with label 3 in cluster 3 in namespace 2")
			By("Registering cluster3 to kubefed")
			By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 3 IP in ipBlocks")
			By("Unregistering cluster3")
		})
	})
})

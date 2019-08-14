package scenarios

import (
	. "github.com/onsi/ginkgo"
)

var _ = PDescribe("[Ctlplane] Removing policies related to removed cluster", func() {
	PIt("Should remove policies relevant to removed cluster", func() {

		By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
		By("creating listener pod 2 with label 2 in cluster 2 in namespace 1")
		By("Registering cluster 3 to kubefed")
		By("creating listener pod 3 with label 3 in cluster 3 in namespace 3")
		By("Adding pod selector based network policy to cluster1 to allow connection from pods in cluster 2 and 3")

		By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 3 IP in ipBlocks")

		By("Unregistering cluster 3")
		By("check NetworkPolicy related to cluster 3 pod IPs is evetually gone")
		By("test connectivity between pods in cluster 1 and 2")
	})
})

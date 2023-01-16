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

package scenarios

import (
	. "github.com/onsi/ginkgo/v2"
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

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

var _ = PDescribe("[Dataplane] Namespace selector based Network Policies for ingress between clusters, selecting multiple remote clusters",
	func() {
		PIt("Should allow communication between pods in selected namespace across multiple clusters", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 4 with label 1 in cluster 1 in namespace 1")
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 2")
			By("creating connecting pod 3 with label 2 in cluster 3 in namespace 2")
			By("creating network policy in cluster 1 allowing pods in namespace 2 on more than 2 clusters" +
				" to communicate with pods in namespace 1")
			By("testing connectivity between pod 1 and 2")
			By("testing connectivity between pod 3 and 4")
		})
	})

var _ = PDescribe("[Ctlplane] Namespace selector based Network Policies for ingress between clusters, selecting multiple remote clusters",
	func() {
		PIt("Should allow communication between pods in selected namespace across multiple clusters", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating listener pod 2 with label 2 in cluster 2 in namespace 2")
			By("creating listener pod 3 with label 3 in cluster 3 in namespace 2")
			By("creating network policy in cluster 1 allowing pods in namespace 2 on more than 2 clusters" +
				" to communicate with pods in namespace 1")
			By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 2 IP in ipBlocks")
			By("Waiting for a NetworkPolicy to appear in cluster 1 containing pod 3 IP in ipBlocks")
		})
	})

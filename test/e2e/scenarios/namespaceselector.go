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

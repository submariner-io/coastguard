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
// Test the scenario where there are no network policies
// implemented in which case all pods should be able to
// communicate with each other

package scenarios

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = PDescribe("[Dataplane] No network policy is implemented", func() {
	PContext("When no network policy is defined", func() {
		PIt("Should allow communication between all pods", func() {
			By("creating listener pod 1 with label 1 in cluster 1 in namespace 1")
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 1")
			By("Testing connectivity between pod 1 and pod 2")
		})
	})

	PContext("[Dataplane] When network plugin does not support network policy", func() {
		PIt("Should fail the tests and log the reason of failure", func() {
			By("creating an empty network policy")
			By("creating listener pod 1 with label 1 in cluster1 in namespace 1")
			By("creating connecting pod 2 with label 2 in cluster 2 in namespace 1")
			By("testing connectivity between pods")
			By("failing the tests if the connection between the pods is not established" +
				" and logging that the tests failed becasue the network plugin used does not seem to implement metwork policies")
		})
	})
})

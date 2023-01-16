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

package controller

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/submariner-io/coastguard/pkg/remotecluster"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
)

var _ = Describe("Coastguard Controller", func() {
	klog.InitFlags(nil)

	var (
		cgController *CoastguardController
		stopChan     chan struct{}
	)

	const clusterID1 = "test-cluster-1"
	const clusterID2 = "test-cluster-2"

	BeforeEach(func() {
		cgController = New()
		stopChan = make(chan struct{})
	})

	Context("Discovery of clusters", func() {
		It("Should add the cluster to the remoteClusters list", func() {
			clientSet := fake.NewSimpleClientset()
			cgController.addCluster(clusterID1, clientSet)
			Expect(cgController.remoteClusters).Should(HaveKey(clusterID1))
		})

		It("Should add a cluster to the synced cluster list once the informer is fully synchronized", func() {
			remoteCluster := remotecluster.New(clusterID1, fake.NewSimpleClientset())
			cgController.onClusterFinishedSyncing(remoteCluster)
			Expect(cgController.syncedClusters).Should(HaveKey(clusterID1))
		})
	})

	Context("Controller and remoteCluster interactions", func() {
		BeforeEach(func() {
			clientSet := fake.NewSimpleClientset()
			cgController.addCluster(clusterID1, clientSet)
			cgController.addCluster(clusterID2, clientSet)
		})

		It("Should connect remoteCluster channel to controller once it is fully synchronized", func() {
			remoteCluster := cgController.remoteClusters[clusterID1]

			By("Simulating a ClusterFinishedSyncing event")
			cgController.onClusterFinishedSyncing(remoteCluster)
			pod := &v1.Pod{}
			remoteCluster.OnAdd(pod)
			var event *remotecluster.Event
			Expect(cgController.clusterEvents).Should(Receive(&event))
			Expect(event.Type).Should(Equal(remotecluster.AddEvent))
			Expect(event.ObjType).Should(Equal(remotecluster.Pod))
		})

		It("Should not stop remoteClusters while running", func() {
			go func() {
				defer close(stopChan)

				// make sure the remote clusters declare themselves as started
				// before we close the channel
				for _, rc := range cgController.remoteClusters {
					Eventually(rc.Stopped(), 5).Should(BeFalse())
				}
			}()

			cgController.Run(stopChan)

			for _, rc := range cgController.remoteClusters {
				Eventually(rc.Stopped(), 5).Should(BeTrue())
			}
		})
	})
})

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coastguard: Controller suite")
}

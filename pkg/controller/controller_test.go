package controller

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"

	"github.com/submariner-io/coastguard/pkg/remotecluster"
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

		It("Should start watching a cluster which was added after discovery", func() {
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
			Expect(cgController.clusterEvents).Should(Receive())

		})

		It("Should not stop remoteClusters while running", func() {
			go func() {
				defer close(stopChan)
				// allow some time for the controller to really Run and wait on the stopChan
				time.Sleep(time.Second)
			}()

			cgController.Run(stopChan)

			for _, rc := range cgController.remoteClusters {
				Eventually(rc.Stopped(), 5).Should(BeTrue())
			}
		})

		It("Should stop remoteCluster objects once stopped", func() {
			go func() {
				defer close(stopChan)
				// allow some time for the controller to really Run and wait on the stopChan
				time.Sleep(time.Second)

				for _, rc := range cgController.remoteClusters {
					Expect(rc.Stopped()).NotTo(BeTrue())
				}
			}()
			cgController.Run(stopChan)
		})
	})
})

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coastguard: Controller suite")
}

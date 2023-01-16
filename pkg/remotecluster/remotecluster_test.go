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

package remotecluster

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	clusterID1 = "test-cluster-1"
	pod0       = "pod0"
	np0        = "np0"
)

var _ = Describe("Coastguard remotecluster", func() {
	klog.InitFlags(nil)

	Describe("extractEventDetails", describeExtractEventDetails)
	Describe("Event conversion", describeEventConversion)
	Describe("RemoteCluster class", describeRemoteCluster)
})

func describeExtractEventDetails() {
	var remoteCluster RemoteCluster

	BeforeEach(func() {
		remoteCluster = RemoteCluster{ClusterID: clusterID1}
	})
	When("a NetworkPolicy event is processed", func() {
		It("Should extract details properly", func() {
			pod := NewPod(testPodName)
			event := remoteCluster.extractEventDetails(pod, &Event{})
			Expect(event.ObjType).To(Equal(Pod))
			Expect(event.ObjID).To(Equal(clusterID1 + ":" + testNamespace + "/" + testPodName + "/" + testUID))
		})
	})

	When("a NetworkPolicy event is processed", func() {
		It("Should extract details properly", func() {
			np := NewNetworkPolicy(testNetworkPolicyName)
			event := remoteCluster.extractEventDetails(np, &Event{})
			Expect(event.ObjType).To(Equal(NetworkPolicy))
			Expect(event.ObjID).To(Equal(clusterID1 + ":" + testNamespace + "/" + testNetworkPolicyName + "/" + testUID))
		})
	})

	When("a cache.DeletedFinalStateUnknown containing another obj is processed", func() {
		It("Should extract a tombstone Pod properly", func() {
			dfsu := cache.DeletedFinalStateUnknown{Obj: NewPod(testPodName)}
			event := remoteCluster.extractEventDetails(dfsu, &Event{})
			Expect(event.ObjType).To(Equal(Pod))
		})
		It("Should extract a tombstone NetworkPolicy properly", func() {
			dfsu := cache.DeletedFinalStateUnknown{Obj: NewNetworkPolicy(testNetworkPolicyName)}
			event := remoteCluster.extractEventDetails(dfsu, &Event{})
			Expect(event.ObjType).To(Equal(NetworkPolicy))
		})
	})

	When("an unexpected object is received", func() {
		It("Should just ignore it and return nil", func() {
			unexpectedObj := "unexpected"
			event := remoteCluster.extractEventDetails(unexpectedObj, &Event{})
			Expect(event).To(BeNil())
		})
	})
}

func describeEventConversion() {
	var remoteCluster RemoteCluster

	BeforeEach(func() {
		remoteCluster = RemoteCluster{ClusterID: clusterID1}
	})

	When("an event is converted from add to update", func() {
		It("Should be correct", func() {
			event := remoteCluster.NewAddEvent(NewPod(testPodName))
			updateEvent := event.ToUpdatedFrom(NewPod(testPodNameOld))
			Expect(updateEvent.Type).To(Equal(UpdateEvent))
			Expect(updateEvent.Objs).To(HaveLen(2))

			oldPod, newPod := updateEvent.Objs[0].(*v1.Pod), updateEvent.Objs[1].(*v1.Pod)
			Expect(oldPod.Name).To(Equal(testPodNameOld))
			Expect(newPod.Name).To(Equal(testPodName))
		})
	})

	When("an event is converted from update to add", func() {
		It("Should be correct", func() {
			event := remoteCluster.NewUpdateEvent(NewPod(testPodNameOld), NewPod(testPodName))
			addedEvent := event.ToAdded()
			Expect(addedEvent.Type).To(Equal(AddEvent))
			Expect(addedEvent.Objs).To(HaveLen(1))

			newPod := addedEvent.Objs[0].(*v1.Pod)
			Expect(newPod.Name).To(Equal(testPodName))
		})
	})
}

func describeRemoteCluster() {
	var eventChannel chan *Event
	BeforeEach(func() {
		eventChannel = make(chan *Event, 10)
	})
	Context("Synchronization", func() {
		It("Should notify when cluster synchronization has finished", func() {
			remoteCluster := New(clusterID1, fake.NewSimpleClientset())
			defer remoteCluster.Stop()

			done := make(chan bool)

			remoteCluster.Run(func(*RemoteCluster) {
				done <- true
			})
			// Wait for sync first
			Eventually(done).Should(Receive(BeTrue()))
		})

		It("Should eventually return HasSynced true", func() {
			remoteCluster := New(clusterID1, fake.NewSimpleClientset())
			defer remoteCluster.Stop()

			remoteCluster.Run(nil)

			Eventually(remoteCluster.HasSynced).Should(BeTrue())
		})
	})
	Context("Event handling", func() {
		It("Should send events on discovered pods", func() {
			remoteCluster, _ := createRemoteClusterWithPod(eventChannel)
			defer remoteCluster.Stop()

			By("Waiting for the Pod AddEvent to be received")
			var event *Event
			Eventually(eventChannel).Should(Receive(&event))
			Expect(event.Type).Should(Equal(AddEvent))
			Expect(event.ObjType).Should(Equal(Pod))

			Consistently(eventChannel).ShouldNot(Receive())
		})

		It("Should send events on discovered NetworkPolicies", func() {
			remoteCluster, _ := createRemoteClusterWithNetworkPolicy(eventChannel)
			defer remoteCluster.Stop()

			By("Waiting for the NetworkPolicy AddEvent to be received")
			var event *Event
			Eventually(eventChannel).Should(Receive(&event))
			Expect(event.Type).Should(Equal(AddEvent))
			Expect(event.ObjType).Should(Equal(NetworkPolicy))

			Consistently(eventChannel).ShouldNot(Receive())
		})

		It("Should discover newly created Pods", func() {
			remoteCluster := createRemoteClusterWithObjects(eventChannel)

			_, err := remoteCluster.ClientSet.CoreV1().Pods("default").Create(NewPod("pod1"))
			Expect(err).ShouldNot(HaveOccurred())

			By("Waiting for the Pod AddEvent to be received")
			var event *Event
			Eventually(eventChannel).Should(Receive(&event))
			Expect(event.Type).Should(Equal(AddEvent))
			Expect(event.ObjType).Should(Equal(Pod))
			Expect(event.Objs).Should(HaveLen(1))
		})

		It("Should discover newly created NetworkPolicies", func() {
			remoteCluster := createRemoteClusterWithObjects(eventChannel)

			_, err := remoteCluster.ClientSet.NetworkingV1().NetworkPolicies("default").Create(NewNetworkPolicy(np0))
			Expect(err).ShouldNot(HaveOccurred())

			By("Waiting for the NetworkPolicy AddEvent to be received")
			var event *Event
			Eventually(eventChannel).Should(Receive(&event))
			Expect(event.Type).Should(Equal(AddEvent))
			Expect(event.ObjType).Should(Equal(NetworkPolicy))
			Expect(event.Objs).Should(HaveLen(1))
		})

		It("Should discover Pods being deleted", func() {
			remoteCluster, _ := createRemoteClusterWithPod(eventChannel)
			defer remoteCluster.Stop()
			cs := remoteCluster.ClientSet

			By("Deleting the test pod")
			err := cs.CoreV1().Pods("default").Delete(pod0, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())

			By("Waiting for the Pod AddEvent, then DeleteEvent to be received")

			for _, rcvType := range []EventType{AddEvent, DeleteEvent} {
				var event *Event
				Eventually(eventChannel).Should(Receive(&event))
				Expect(event.Type).Should(Equal(rcvType))
				Expect(event.Objs).Should(HaveLen(1))
			}
		})

		It("Should discover NetworkPolicies being deleted", func() {
			remoteCluster, _ := createRemoteClusterWithNetworkPolicy(eventChannel)
			defer remoteCluster.Stop()
			cs := remoteCluster.ClientSet

			By("Deleting the test pod")
			err := cs.NetworkingV1().NetworkPolicies("default").Delete(np0, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())

			By("Waiting for the NetworkPolicy AddEvent, then DeleteEvent to be received")

			for _, rcvType := range []EventType{AddEvent, DeleteEvent} {
				var event *Event
				Eventually(eventChannel).Should(Receive(&event))
				Expect(event.Type).Should(Equal(rcvType))
				Expect(event.Objs).Should(HaveLen(1))
			}
		})

		It("It should discover pods being updated", func() {
			remoteCluster, pod := createRemoteClusterWithPod(eventChannel)
			defer remoteCluster.Stop()
			cs := remoteCluster.ClientSet

			By("Updating the test pod label")
			pod.SetLabels(map[string]string{"label-one": "1"})
			_, err := cs.CoreV1().Pods("default").Update(pod)
			Expect(err).ShouldNot(HaveOccurred())

			By("Waiting for the Pod AddEvent, then UpdateEvent to be received")

			Eventually(eventChannel).Should(Receive())
			var event *Event
			Eventually(eventChannel).Should(Receive(&event))
			Expect(event.Type).Should(Equal(UpdateEvent))
			Expect(event.Objs).Should(HaveLen(2))
		})

		It("It should discover NetworkPolicies being updated", func() {
			remoteCluster, np := createRemoteClusterWithNetworkPolicy(eventChannel)
			defer remoteCluster.Stop()
			cs := remoteCluster.ClientSet

			By("Updating the test NetworkPolicy label")
			np.SetLabels(map[string]string{"label-one": "1"})
			_, err := cs.NetworkingV1().NetworkPolicies("default").Update(np)
			Expect(err).ShouldNot(HaveOccurred())

			By("Waiting for the NetworkPolicy AddEvent, then UpdateEvent to be received")

			Eventually(eventChannel).Should(Receive())
			var event *Event
			Eventually(eventChannel).Should(Receive(&event))
			Expect(event.Type).Should(Equal(UpdateEvent))
			Expect(event.Objs).Should(HaveLen(2))
		})
	})
	Context("Finalization of the RemoteCluster watcher", func() {
		It("Should return stopped once we stop it", func() {
			remoteCluster, _ := createRemoteClusterWithPod(eventChannel)
			Expect(remoteCluster.Stopped()).To(BeFalse())
			remoteCluster.Stop()
			Expect(remoteCluster.Stopped()).To(BeTrue())
		})
	})
	Context("Access to the informers cache", func() {
		It("Should be able to list existing pods in informer cache", func() {
			remoteCluster, _ := createRemoteClusterWithPod(eventChannel)
			Expect(remoteCluster.GetPods()).Should(HaveLen(1))
		})
	})
}

func createRemoteClusterWithObjects(eventChannel chan *Event, objects ...runtime.Object) *RemoteCluster {
	clientSet := fake.NewSimpleClientset(objects...)

	By("Creating a new remoteCluster with a clientset of one pod")

	remoteCluster := New(clusterID1, clientSet)
	remoteCluster.SetEventChannel(eventChannel)

	done := make(chan bool)

	remoteCluster.Run(func(*RemoteCluster) {
		done <- true
	})
	// Wait for sync first
	Eventually(done).Should(Receive(BeTrue()))

	return remoteCluster
}

func createRemoteClusterWithPod(eventChannel chan *Event) (*RemoteCluster, *v1.Pod) {
	testPod := NewPod(pod0)

	return createRemoteClusterWithObjects(eventChannel,
		&v1.PodList{Items: []v1.Pod{*testPod}}), testPod
}

func createRemoteClusterWithNetworkPolicy(eventChannel chan *Event) (*RemoteCluster, *v1net.NetworkPolicy) {
	testNetworkPolicy := NewNetworkPolicy(np0)

	return createRemoteClusterWithObjects(eventChannel,
		&v1net.NetworkPolicyList{Items: []v1net.NetworkPolicy{*testNetworkPolicy}}), testNetworkPolicy
}

const (
	testNamespace         = "default"
	testPodName           = "pod1"
	testPodNameOld        = "pod1-old"
	testNetworkPolicyName = "np1"
	testUID               = "ff3b5269-1201-4e2c-95f5-46fc69ff6c63"
)

func NewPod(name string) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      name,
			UID:       testUID,
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	return pod
}

func NewNetworkPolicy(name string) *v1net.NetworkPolicy {
	np := &v1net.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			UID:       testUID,
		},
	}

	return np
}

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coastguard: RemoteCluster suite")
}

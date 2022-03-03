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
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const defaultResyncTime = time.Hour * 24

type RemoteCluster struct {
	stopCh chan struct{}

	ClusterID             string
	ClientSet             kubernetes.Interface
	podInformer           cache.SharedIndexInformer
	networkPolicyInformer cache.SharedIndexInformer

	eventChanMutex *sync.Mutex
	eventChan      chan *Event
}

type EventType string

const (
	AddEvent    EventType = "Added"
	DeleteEvent EventType = "Deleted"
	UpdateEvent EventType = "Updated"
)

type ObjectType string

const (
	NetworkPolicy ObjectType = "np"
	Pod           ObjectType = "pod"
)

type Event struct {
	Cluster *RemoteCluster
	Type    EventType
	ObjType ObjectType
	Objs    []interface{}
	ObjID   string
}

func (ev *Event) ToUpdatedFrom(oldObj interface{}) *Event {
	if ev.Type == AddEvent {
		ev.Objs = append([]interface{}{oldObj}, ev.Objs[0])
		ev.Type = UpdateEvent
	} else {
		klog.Fatal("only AddEvents can be converted to UpdateEvents")
	}

	return ev
}

func (ev *Event) ToAdded() *Event {
	if ev.Type == UpdateEvent {
		ev.Objs = []interface{}{ev.Objs[1]}
		ev.Type = AddEvent
	} else {
		klog.Fatal("only UpdateEvents can be converted to AddEvents")
	}

	return ev
}

func New(clusterID string, clientSet kubernetes.Interface) *RemoteCluster {
	factory := informers.NewSharedInformerFactory(clientSet, defaultResyncTime)
	podInformer := factory.Core().V1().Pods().Informer()
	networkPolicyInformer := factory.Networking().V1().NetworkPolicies().Informer()

	resourceWatcher := &RemoteCluster{
		stopCh:                make(chan struct{}),
		ClusterID:             clusterID,
		ClientSet:             clientSet,
		podInformer:           podInformer,
		networkPolicyInformer: networkPolicyInformer,
		eventChanMutex:        &sync.Mutex{},
	}

	podInformer.AddEventHandler(resourceWatcher)
	networkPolicyInformer.AddEventHandler(resourceWatcher)

	return resourceWatcher
}

func (rc *RemoteCluster) HasSynced() bool {
	return rc.podInformer.HasSynced() &&
		rc.networkPolicyInformer.HasSynced()
}

// Stop will stop the running informers.
func (rc *RemoteCluster) Stop() {
	close(rc.stopCh)
}

func (rc *RemoteCluster) Stopped() bool {
	select {
	case _, ok := <-rc.stopCh:
		// channel is not ok to read, means it's closed now
		return !ok
	default:
		// channel was ok to read, but had no messages
		return false
	}
}

func (rc *RemoteCluster) Run(onSyncDoneFunc func(resourceWatcher *RemoteCluster)) {
	go rc.podInformer.Run(rc.stopCh)
	go rc.networkPolicyInformer.Run(rc.stopCh)

	go func() {
		if !cache.WaitForCacheSync(rc.stopCh, rc.podInformer.HasSynced) {
			klog.Warning("Timed out waiting for pod informer to sync")
		}

		if !cache.WaitForCacheSync(rc.stopCh, rc.networkPolicyInformer.HasSynced) {
			klog.Warning("Timed out waiting for NetworkPolicy informer to sync")
		}

		if onSyncDoneFunc != nil {
			onSyncDoneFunc(rc)
		}
	}()
}

func (rc *RemoteCluster) GetPods() []interface{} {
	return rc.podInformer.GetStore().List()
}

func (rc *RemoteCluster) GetNetworkPolicies() []interface{} {
	return rc.networkPolicyInformer.GetStore().List()
}

func (rc *RemoteCluster) SetEventChannel(eventChan chan *Event) {
	rc.eventChanMutex.Lock()
	rc.eventChan = eventChan
	rc.eventChanMutex.Unlock()
}

func (rc *RemoteCluster) OnAdd(obj interface{}) {
	rc.enqueueEvent(rc.NewAddEvent(obj))
}

func (rc *RemoteCluster) OnDelete(obj interface{}) {
	rc.enqueueEvent(rc.NewDeleteEvent(obj))
}

func (rc *RemoteCluster) OnUpdate(oldObj, newObj interface{}) {
	rc.enqueueEvent(rc.NewUpdateEvent(oldObj, newObj))
}

func (rc *RemoteCluster) enqueueEvent(event *Event) {
	if event == nil {
		return
	}

	// lock used as memory barrier to make sure eventChannel ref
	// is synchronized between threads
	rc.eventChanMutex.Lock()
	defer rc.eventChanMutex.Unlock()

	if rc.eventChan != nil {
		rc.eventChan <- event
	}
}

func (rc *RemoteCluster) NewAddEvent(objInterface interface{}) *Event {
	event := Event{
		Cluster: rc,
		Type:    AddEvent,
		Objs:    []interface{}{objInterface},
	}

	return rc.extractEventDetails(objInterface, &event)
}

func (rc *RemoteCluster) NewUpdateEvent(objInterface, newObjInterface interface{}) *Event {
	event := Event{
		Cluster: rc,
		Type:    UpdateEvent,
		Objs:    []interface{}{objInterface, newObjInterface},
	}

	return rc.extractEventDetails(newObjInterface, &event)
}

func (rc *RemoteCluster) NewDeleteEvent(objInterface interface{}) *Event {
	event := Event{
		Cluster: rc,
		Type:    DeleteEvent,
		Objs:    []interface{}{objInterface},
	}

	return rc.extractEventDetails(objInterface, &event)
}

func (rc *RemoteCluster) extractEventDetails(objInterface interface{}, event *Event) *Event {
	switch obj := objInterface.(type) {
	case *v1.Pod:
		event.ObjType = Pod
		event.ObjID = ObjID(rc.ClusterID, obj.Namespace, obj.Name, obj.UID)
	case *v1net.NetworkPolicy:
		event.ObjType = NetworkPolicy
		event.ObjID = ObjID(rc.ClusterID, obj.Namespace, obj.Name, obj.UID)
	case cache.DeletedFinalStateUnknown:
		return rc.extractEventDetails(obj.Obj, event)
	default:
		klog.Errorf("%s for unexpected type object: %v", event.Type, objInterface)
		return nil
	}

	return event
}

func ObjID(clusterID, ns, name string, uid types.UID) string {
	return fmt.Sprintf("%s:%s/%s/%s", clusterID, ns, name, string(uid))
}

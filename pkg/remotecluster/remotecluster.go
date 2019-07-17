package remotecluster

import (
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

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
}

func New(clusterID string, clientSet kubernetes.Interface) *RemoteCluster {

	factory := informers.NewSharedInformerFactory(clientSet, time.Hour*24)
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

func (rw *RemoteCluster) HasSynced() bool {
	return rw.podInformer.HasSynced() &&
		rw.networkPolicyInformer.HasSynced()
}

// Stop will stop the running informers
func (rw *RemoteCluster) Stop() {
	close(rw.stopCh)
}

func (rw *RemoteCluster) Stopped() bool {
	select {
	case _, ok := <-rw.stopCh:
		// channel is not ok to read, means it's closed now
		return !ok
	default:
		// channel was ok to read, but had no messages
		return false
	}
}

func (rw *RemoteCluster) Run(onSyncDoneFunc func(resourceWatcher *RemoteCluster)) {
	go rw.podInformer.Run(rw.stopCh)
	go rw.networkPolicyInformer.Run(rw.stopCh)

	go func() {
		if !cache.WaitForCacheSync(rw.stopCh, rw.podInformer.HasSynced) {
			klog.Warning("Timed out waiting for pod informer to sync")
		}

		if !cache.WaitForCacheSync(rw.stopCh, rw.networkPolicyInformer.HasSynced) {
			klog.Warning("Timed out waiting for NetworkPolicy informer to sync")
		}

		if onSyncDoneFunc != nil {
			onSyncDoneFunc(rw)
		}
	}()
}

func (rw *RemoteCluster) GetPods() []interface{} {
	return rw.podInformer.GetStore().List()
}

func (rw *RemoteCluster) GetNetworkPolicies() []interface{} {
	return rw.networkPolicyInformer.GetStore().List()
}

func (rw *RemoteCluster) SetEventChannel(eventChan chan *Event) {
	rw.eventChanMutex.Lock()
	rw.eventChan = eventChan
	rw.eventChanMutex.Unlock()
}

func (rw *RemoteCluster) enqueueEvent(event *Event) {

	// lock used as memory barrier to make sure eventChannel ref
	// is synchronized between threads
	rw.eventChanMutex.Lock()
	if rw.eventChan != nil {
		rw.eventChan <- event
	}
	rw.eventChanMutex.Unlock()
}

func (rw *RemoteCluster) OnAdd(obj interface{}) {

	var objType ObjectType

	if pod, ok := obj.(*v1.Pod); ok {
		klog.Infof("Pod discovered in %s: %s", rw.ClusterID, pod.GetSelfLink())
		objType = Pod
	} else if np, ok := obj.(*v1net.NetworkPolicy); ok {
		klog.Infof("NetworkPolicy discovered in %s: %s", rw.ClusterID, np.GetSelfLink())
		objType = NetworkPolicy
	} else {
		klog.Errorf("OnAdd for unexpected object type: %v", obj)
		return
	}

	rw.enqueueEvent(&Event{
		Cluster: rw,
		Type:    AddEvent,
		ObjType: objType,
		Objs:    []interface{}{obj},
	})
}

func (rw *RemoteCluster) OnDelete(obj interface{}) {

	var objType ObjectType

	if pod, ok := obj.(*v1.Pod); ok {
		klog.Infof("Pod deleted in %s: %s", rw.ClusterID, pod.GetSelfLink())
		objType = Pod
	} else if np, ok := obj.(*v1net.NetworkPolicy); ok {
		klog.Infof("NetworkPolicy deleted in %s: %s", rw.ClusterID, np.GetSelfLink())
		objType = NetworkPolicy
	} else {
		klog.Errorf("OnDelete for unexpected object type: %v", obj)
		return
	}

	rw.enqueueEvent(&Event{
		Cluster: rw,
		Type:    DeleteEvent,
		ObjType: objType,
		Objs:    []interface{}{obj},
	})
}

func (rw *RemoteCluster) OnUpdate(oldObj interface{}, newObj interface{}) {

	var objType ObjectType

	if newPod, ok := newObj.(*v1.Pod); ok {
		klog.Infof("Pod updated in %s: %s", rw.ClusterID, newPod.GetSelfLink())
		objType = Pod
	} else if newNP, ok := newObj.(*v1net.NetworkPolicy); ok {
		klog.Infof("NetworkPolicy updated in %s: %s", rw.ClusterID, newNP.GetSelfLink())
		objType = NetworkPolicy
	} else {
		klog.Errorf("OnUpdate for unexpected object type: %v", newObj)
		return
	}

	rw.enqueueEvent(&Event{
		Cluster: rw,
		Type:    UpdateEvent,
		ObjType: objType,
		Objs:    []interface{}{oldObj, newObj},
	})
}

package controller

import (
	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	"k8s.io/klog"

	"github.com/submariner-io/coastguard/pkg/networkpolicy"
	"github.com/submariner-io/coastguard/pkg/remotecluster"
)

func (c *CoastguardController) onClusterFinishedSyncing(cluster *remotecluster.RemoteCluster) {

	c.processingMutex.Lock()
	defer c.processingMutex.Unlock()

	klog.Infof("Cluster %s finished syncing", cluster.ClusterID)
	c.syncedClusters[cluster.ClusterID] = cluster
}

func (c *CoastguardController) processLoop(stopCh <-chan struct{}) {
	for {
		select {
		case event := <-c.clusterEvents:
			c.processEvent(event)
		case <-stopCh:
			klog.Info("exited process loop")
			return
		}
	}
}

func (c *CoastguardController) processEvent(event *remotecluster.Event) {

	if event == nil {
		klog.Error("processEvent received nil remotecluster.Event")
		return
	}

	klog.Infof("%s\t%s\t%s", event.Type, event.ObjType, event.ObjID)
	switch event.ObjType {
	case remotecluster.NetworkPolicy:
		c.processNetworkPolicyEvent(event)
	case remotecluster.Pod:
		c.processPodEvent(event)
	}
}

func (c *CoastguardController) processNetworkPolicyEvent(event *remotecluster.Event) {
	switch event.Type {
	case remotecluster.AddEvent:
		c.addedRemoteNetworkPolicy(event)
	case remotecluster.UpdateEvent:
		c.updateRemoteNetworkPolicy(event)
	case remotecluster.DeleteEvent:
		c.deleteRemoteNetworkPolicy(event)
	}
}

func (c *CoastguardController) processPodEvent(event *remotecluster.Event) {

	switch event.Type {
	case remotecluster.AddEvent:
		c.addedPod(event)
	case remotecluster.UpdateEvent:
		c.updatePod(event)
	case remotecluster.DeleteEvent:
		c.deletePod(event)
	}
}

func (c *CoastguardController) addedRemoteNetworkPolicy(event *remotecluster.Event) {

	if rnp, exists := c.remoteNetworkPolicies[event.ObjID]; !exists {
		np := event.Objs[0].(*v1net.NetworkPolicy)
		c.remoteNetworkPolicies[event.ObjID] = networkpolicy.NewRemoteNetworkPolicy(np, event.Cluster, event.ObjID, c.remotePods)
	} else {
		event.Objs = append([]interface{}{rnp.Np}, event.Objs[0])
		c.updateRemoteNetworkPolicy(event)
	}
}

func (c *CoastguardController) updateRemoteNetworkPolicy(event *remotecluster.Event) {

	if _, exists := c.remoteNetworkPolicies[event.ObjID]; exists {
		np := event.Objs[1].(*v1net.NetworkPolicy)
		rnp := networkpolicy.NewRemoteNetworkPolicy(np, event.Cluster, event.ObjID, c.remotePods)
		c.remoteNetworkPolicies[event.ObjID] = rnp
	} else {
		c.addedRemoteNetworkPolicy(event)
	}
}

func (c *CoastguardController) deleteRemoteNetworkPolicy(event *remotecluster.Event) {

	if _, exists := c.remoteNetworkPolicies[event.ObjID]; exists {
		delete(c.remoteNetworkPolicies, event.ObjID)
	} else {
		klog.Warningf("A deleteNetworkPolicy event was received for a np not in our cache: %s", event.ObjID)

	}
}

func (c *CoastguardController) addedPod(event *remotecluster.Event) {
	pod := event.Objs[0].(*v1.Pod)
	if rp, exists := c.remotePods[event.ObjID]; !exists {
		c.remotePods[event.ObjID] = networkpolicy.NewRemotePod(pod, event.Cluster, event.ObjID)
		for _, np := range c.remoteNetworkPolicies {
			np.AddedPod(event)
		}
	} else {
		klog.Warningf("An addPod event was received for a pod already in our cache: %s, updating instead", event.ObjID)
		c.updatePod(event.ToUpdatedFrom(rp.Pod))
	}
}

func (c *CoastguardController) updatePod(event *remotecluster.Event) {
	if _, exists := c.remotePods[event.ObjID]; exists {
		pod := event.Objs[0].(*v1.Pod)
		c.remotePods[event.ObjID] = networkpolicy.NewRemotePod(pod, event.Cluster, event.ObjID)
		for _, np := range c.remoteNetworkPolicies {
			np.UpdatedPod(event)
		}
	} else {
		klog.Warningf("An updatePod event was received for a pod not in our cache: %s, adding instead", event.ObjID)
		c.addedPod(event.ToAdded())
	}
}

func (c *CoastguardController) deletePod(event *remotecluster.Event) {

	if _, exists := c.remotePods[event.ObjID]; exists {
		for _, np := range c.remoteNetworkPolicies {
			np.DeletedPod(event)
		}
		delete(c.remotePods, event.ObjID)
	} else {
		klog.Warningf("An deletePod event was received for a pod not in our cache: %s", event.ObjID)
	}
}

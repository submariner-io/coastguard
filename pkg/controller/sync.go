package controller

import (
	"k8s.io/klog"

	"github.com/submariner-io/coastguard/pkg/remotecluster"
)

func (c *CoastguardController) onClusterFinishedSyncing(cluster *remotecluster.RemoteCluster) {

	c.processingMutex.Lock()
	defer c.processingMutex.Unlock()

	klog.Infof("Cluster %s finished syncing", cluster.ClusterID)
	c.syncedClusters[cluster.ClusterID] = cluster

	cluster.SetEventChannel(c.clusterEvents)
}

func (c *CoastguardController) processLoop(stopCh <-chan struct{}) {
	for {
		select {
		case event := <-c.clusterEvents:
			c.processingMutex.Lock()
			klog.Infof("event: %v", event)
			c.processingMutex.Unlock()
		case <-stopCh:
			klog.Info("exited process loop")
			return
		}
	}
}

package controller

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"github.com/submariner-io/coastguard/pkg/remotecluster"
)

func (c *CoastguardController) OnAdd(clusterID string, kubeConfig *rest.Config) {
	klog.Infof("adding cluster: %s", clusterID)

	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("error creating clientset for cluster %s: %s", clusterID, err.Error())
		return
	}

	c.addCluster(clusterID, clientSet)
}

func (c *CoastguardController) addCluster(clusterID string, clientSet kubernetes.Interface) {
	rc := remotecluster.New(clusterID, clientSet)
	rc.SetEventChannel(c.clusterEvents)
	c.processingMutex.Lock()
	c.remoteClusters[clusterID] = rc
	c.processingMutex.Unlock()
	rc.Run(c.onClusterFinishedSyncing)
}

func (c *CoastguardController) OnUpdate(clusterID string, kubeConfig *rest.Config) {

	klog.Infof("updating cluster: %s", clusterID)
	klog.Fatalf("Not implemented yet")
}

func (c *CoastguardController) OnRemove(clusterID string) {

	klog.Infof("removing cluster: %s", clusterID)
	klog.Fatalf("Not implemented yet")
}

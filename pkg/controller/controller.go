package controller

import (
	"sync"

	"github.com/submariner-io/admiral/pkg/federate"
	"k8s.io/klog"

	"github.com/submariner-io/coastguard/pkg/healthz"
	"github.com/submariner-io/coastguard/pkg/remotecluster"
)

// arbitrary event channel size
const eventChannelSize = 1000

type CoastguardController struct {
	federate.ClusterEventHandler

	// remoteClusters is a map of remote clusters, which have been discovered
	remoteClusters map[string]*remotecluster.RemoteCluster
	// syncedClusters is a map of remote clusters, which have been discovered,
	// and also our local cache is in sync with them
	syncedClusters map[string]*remotecluster.RemoteCluster

	// clusterEvents is the channel to receive events from all the
	// existing remote clusters
	clusterEvents chan *remotecluster.Event

	// processingMutex is used to avoid synchronization issues when handling
	// objects inside the controller, it's a generalistic lock, although
	// later in time we can come up with a more granular implementation.
	processingMutex *sync.Mutex
}

func New() *CoastguardController {

	return &CoastguardController{
		remoteClusters:  make(map[string]*remotecluster.RemoteCluster),
		syncedClusters:  make(map[string]*remotecluster.RemoteCluster),
		processingMutex: &sync.Mutex{},
		clusterEvents:   make(chan *remotecluster.Event, eventChannelSize),
	}
}

func (c *CoastguardController) Run(stopCh <-chan struct{}) {

	go c.processLoop(stopCh)

	healthzServer := healthz.New(":8080")
	go healthzServer.Run(stopCh)

	// we stop here until the stopCh channel is closed
	<-stopCh

	klog.Info("Stopping remote cluster informers")

	// ensure other go routines have stopped too
	for _, remoteCluster := range c.remoteClusters {
		remoteCluster.Stop()
	}

	close(c.clusterEvents)
}

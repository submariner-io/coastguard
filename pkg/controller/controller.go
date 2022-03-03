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
	"sync"

	"github.com/submariner-io/coastguard/pkg/healthz"
	"github.com/submariner-io/coastguard/pkg/networkpolicy"
	"github.com/submariner-io/coastguard/pkg/remotecluster"
	"k8s.io/klog"
)

// arbitrary event channel size.
const eventChannelSize = 1000

type CoastguardController struct {
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

	remoteNetworkPolicies    map[string]*networkpolicy.RemoteNetworkPolicy
	remoteGenNetworkPolicies map[string]*remoteGeneratedNetworkPolicy
	remotePods               map[string]*networkpolicy.RemotePod
}

func New() *CoastguardController {
	return &CoastguardController{
		remoteClusters:           make(map[string]*remotecluster.RemoteCluster),
		syncedClusters:           make(map[string]*remotecluster.RemoteCluster),
		processingMutex:          &sync.Mutex{},
		clusterEvents:            make(chan *remotecluster.Event, eventChannelSize),
		remoteNetworkPolicies:    make(map[string]*networkpolicy.RemoteNetworkPolicy),
		remoteGenNetworkPolicies: make(map[string]*remoteGeneratedNetworkPolicy),
		remotePods:               make(map[string]*networkpolicy.RemotePod),
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

func (c *CoastguardController) AllClustersSynced() bool {
	c.processingMutex.Lock()
	defer c.processingMutex.Unlock()

	return len(c.syncedClusters) == len(c.remoteClusters)
}

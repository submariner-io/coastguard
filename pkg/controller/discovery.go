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
	"github.com/submariner-io/coastguard/pkg/remotecluster"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
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

func (c *CoastguardController) OnUpdate(clusterID string, _ *rest.Config) {
	klog.Infof("updating cluster: %s", clusterID)
	klog.Fatalf("Not implemented yet")
}

func (c *CoastguardController) OnRemove(clusterID string) {
	klog.Infof("removing cluster: %s", clusterID)
	klog.Fatalf("Not implemented yet")
}

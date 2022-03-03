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
	"github.com/submariner-io/coastguard/pkg/networkpolicy"
	"github.com/submariner-io/coastguard/pkg/remotecluster"
	v1net "k8s.io/api/networking/v1"
	"k8s.io/klog"
)

// RemoteGeneratedNetworkPolicy represents a policy that we generated and sent
// to a remote cluster, that we received back via events. We used to track the
// remote Cluster where it came from.
type remoteGeneratedNetworkPolicy struct {
	// Cluster is the origin of the network policy
	cluster *remotecluster.RemoteCluster

	// Np is the NetworkPolicy origin of this tracking
	np *v1net.NetworkPolicy
}

func (c *CoastguardController) syncGeneratedPolicies() {
	if !c.AllClustersSynced() {
		klog.Info("Skipping generated policy sync until all clusters sync has finished")
		return
	}

	c.processPoliciesNeedingDistribution()
	c.processPoliciesNeedingDelete()
}

func (c *CoastguardController) processPoliciesNeedingDistribution() {
	for objID, rnp := range c.remoteNetworkPolicies {
		if rnp.GeneratedPolicy != nil {
			genPolicyReceived, exists := c.remoteGenNetworkPolicies[objID]
			if !exists || networkpolicy.ArePolicyRulesDifferent(genPolicyReceived.np, rnp.GeneratedPolicy) {
				err := rnp.Cluster.Distribute(rnp.GeneratedPolicy)
				if err != nil {
					klog.Errorf("An error happened trying to distribute a generated policy: %s", err)
				}
			}
		}
	}
}

func (c *CoastguardController) processPoliciesNeedingDelete() {
	for objID, rnp := range c.remoteNetworkPolicies {
		if rnp.GeneratedPolicy != nil {
			continue
		}

		if rgp, exists := c.remoteGenNetworkPolicies[objID]; exists {
			if err := rnp.Cluster.Delete(rgp.np); err != nil {
				klog.Errorf("There was an error deleting a NetworkPolicy %s", err)
			}
		}
	}

	for objID, rgnp := range c.remoteGenNetworkPolicies {
		if _, exists := c.remoteNetworkPolicies[objID]; !exists {
			if err := rgnp.cluster.Delete(rgnp.np); err != nil {
				klog.Errorf("There was an error deleting a NetworkPolicy: %s", err)
			}
		}
	}
}

// processGeneratedNetworkPolicyEvent processes events related to NetworkPolicies that we
// have generated ourselves and that show up on the remote clusters. We should not generate
// new policies based on those, but we should track them.
func (c *CoastguardController) processGeneratedNetworkPolicyEvent(event *remotecluster.Event) {
	switch event.Type {
	case remotecluster.AddEvent:
		c.addedGeneratedNetworkPolicy(event)
	case remotecluster.UpdateEvent:
		c.updatedGeneratedNetworkPolicy(event)
	case remotecluster.DeleteEvent:
		c.deletedGeneratedNetworkPolicy(event)
	}
}

func (c *CoastguardController) addedGeneratedNetworkPolicy(event *remotecluster.Event) {
	np := event.Objs[0].(*v1net.NetworkPolicy)
	origObjID := networkpolicy.OriginatingObjID(np)

	if existingNp, exists := c.remoteGenNetworkPolicies[origObjID]; !exists {
		c.remoteGenNetworkPolicies[origObjID] = &remoteGeneratedNetworkPolicy{cluster: event.Cluster, np: np}
	} else {
		c.updatedGeneratedNetworkPolicy(event.ToUpdatedFrom(existingNp))
	}
}

func (c *CoastguardController) updatedGeneratedNetworkPolicy(event *remotecluster.Event) {
	np := event.Objs[1].(*v1net.NetworkPolicy)
	origObjID := networkpolicy.OriginatingObjID(np)

	if _, exists := c.remoteGenNetworkPolicies[origObjID]; exists {
		c.remoteGenNetworkPolicies[origObjID] = &remoteGeneratedNetworkPolicy{cluster: event.Cluster, np: np}
	} else {
		c.addedGeneratedNetworkPolicy(event.ToAdded())
	}
}

func (c *CoastguardController) deletedGeneratedNetworkPolicy(event *remotecluster.Event) {
	np := event.Objs[0].(*v1net.NetworkPolicy)
	origObjID := networkpolicy.OriginatingObjID(np)

	if _, exists := c.remoteGenNetworkPolicies[origObjID]; exists {
		delete(c.remoteGenNetworkPolicies, origObjID)
	} else {
		klog.Warningf("A deleteNetworkPolicy event was received for a np not in our cache: %s", event.ObjID)
	}
}

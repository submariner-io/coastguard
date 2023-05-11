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
	"context"

	"github.com/pkg/errors"
	v1net "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (rc *RemoteCluster) Distribute(np *v1net.NetworkPolicy) error {
	npClient := rc.ClientSet.NetworkingV1().NetworkPolicies(np.Namespace)

	_, err := npClient.Update(context.TODO(), np, v1.UpdateOptions{})

	if err == nil {
		return nil
	} else if !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "error updating NetworkPolicy %s for cluster %s", np.Name, rc.ClusterID)
	}

	_, err = npClient.Create(context.TODO(), np, v1.CreateOptions{})

	return errors.Wrapf(err, "error creating NetworkPolicy %s for cluster %s", np.Name, rc.ClusterID)
}

func (rc *RemoteCluster) Delete(np *v1net.NetworkPolicy) error {
	npClient := rc.ClientSet.NetworkingV1().NetworkPolicies(np.Namespace)

	return errors.Wrapf(npClient.Delete(context.TODO(), np.Name, v1.DeleteOptions{}),
		"error deleting NetworkPolicy %s from cluster %s", np.Name, rc.ClusterID)
}

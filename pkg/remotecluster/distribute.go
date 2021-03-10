/*
© 2020 Red Hat, Inc.

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
	v1net "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func (r *RemoteCluster) Distribute(np *v1net.NetworkPolicy) error {
	npClient := r.ClientSet.NetworkingV1().NetworkPolicies(np.Namespace)

	_, err := npClient.Update(np)

	klog.Infof("distributing NetworkPolicy %s to cluster %s", np.Name, r.ClusterID)

	if err == nil {
		return nil
	} else if !errors.IsNotFound(err) {
		return err
	}

	_, err = npClient.Create(np)
	return err
}

func (r *RemoteCluster) Delete(np *v1net.NetworkPolicy) error {

	npClient := r.ClientSet.NetworkingV1().NetworkPolicies(np.Namespace)
	klog.Infof("deleting NetworkPolicy %s from cluster %s", np.Name, r.ClusterID)
	return npClient.Delete(np.Name, &v1.DeleteOptions{})

}

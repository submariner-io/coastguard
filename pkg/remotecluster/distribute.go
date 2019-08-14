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

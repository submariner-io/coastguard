# Building and testing

This is a temporary README instructing how to test coastguard deployment manually.
Changes will be made so that coastguard will be deployed in kind using e2e scripts.

This patch builds and packages coastguard locally. The built image then needs to be imported in cluster 1 of kind deployment.
To do the same, follow below steps:

1. Run make build which builds coastguard binary named coastguard-controller
2. Run make package which builds coastguard docker image named submariner-io/coastguard-controller:dev using the Dockerfile
3. Import this image to deployed kind cluster 1 (broker cluster)

    docker tag submariner-io/coastguard-controller:dev coastguard-controller:local
    echo "Loading submariner images in to cluster1..."
    kind --name cluster1 load docker-image coastguard-controller:local

4. Now deploy coastguard in broker cluster with kubectl apply.

    kubectl config use-context cluster1
    kubectl apply -f coastguard_deployment.yaml

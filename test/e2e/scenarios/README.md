# e2e tests

These e2e tests validate the use cases as described in this
[Multicluster NetworkPolicy
Proposal](https://docs.google.com/document/d/1_QzuasJPiQ-4t8tUGODoRS2E-Q3KLJajqMyCvLPH2tY/edit#heading=h.4igbcpajqich).

Each test corresponds to each use case from the above mentioned proposal.

To run these e2e tests, follow the steps as described
[here](https://github.com/submariner-io/submariner#testing).

There are two type of tests, dataplane and controlplane.
Dataplane tests check for connectivity between pods after network
policies are created.
Controlplane tests check for IP of target pods in Coastguard managed
NetworkPolicies, on the cluster where the network policy is created,
that instead of podSelectors use ipBlocks to reference remote pods.

For controlplane tests, only listener pods are created everywhere
which will wait indefinitely until they are killed (as the test
finishes) and for dataplane tests, both listener + connector pods are
created that will both exit once the tests are finished and the
network policies created by coastguard will also be deleted.

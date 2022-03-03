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

package main

import (
	"flag"
	"os"

	"github.com/submariner-io/admiral/pkg/federate"
	"github.com/submariner-io/admiral/pkg/federate/kubefed"
	"github.com/submariner-io/coastguard/pkg/controller"
	"github.com/submariner-io/submariner/pkg/signals"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	kubeConfig string
	masterURL  string
)

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", os.Getenv("KUBECONFIG"),
		"Path to kubeconfig containing embedded authinfo.")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.V(2).Info("Starting coastguard-network-policy-sync")

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	runStoppedCh := make(chan struct{})

	coastGuardController := controller.New()

	go func() {
		defer close(runStoppedCh)
		coastGuardController.Run(stopCh)
	}()

	if err := buildKubeFedFederator(stopCh).WatchClusters(coastGuardController); err != nil {
		klog.Fatalf("Error watching federation clusters: %s", err.Error())
	}

	<-runStoppedCh
	klog.Info("All controllers stopped or exited. Stopping main loop")
}

func buildKubeFedFederator(stopCh <-chan struct{}) federate.Federator {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		klog.Fatalf("Error attempting to load kubeconfig: %s", err.Error())
	}

	federator, err := kubefed.New(kubeConfig, stopCh)
	if err != nil {
		klog.Fatalf("Error creating kubefed federator: %s", err.Error())
	}

	return federator
}

package main

import (
	"flag"
	"sync"

	"github.com/submariner-io/submariner/pkg/signals"
	"k8s.io/klog"

	"github.com/submariner-io/coastguard/pkg/clusters/discovery"
	"github.com/submariner-io/coastguard/pkg/controller"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.V(2).Info("Starting coastguard-network-policy-sync")

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	var wg sync.WaitGroup

	wg.Add(1)

	coastGuardController := controller.New()

	go func() {
		defer wg.Done()
		coastGuardController.Run(stopCh)
	}()

	discovery.Start(coastGuardController)

	wg.Wait()
	klog.Fatal("All controllers stopped or exited. Stopping main loop")
}

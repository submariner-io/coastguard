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
package healthz

import (
	"context"
	"net/http"
	"time"

	"k8s.io/klog"
)

type HealthzServer struct {
	httpServer *http.Server
}

func New(address string) *HealthzServer {
	healthServer := &HealthzServer{
		httpServer: &http.Server{Addr: address},
	}
	healthServer.httpServer.Handler = healthServer

	return healthServer
}

func (hs *HealthzServer) Run(stop <-chan struct{}) {

	listenAndServeFailed := make(chan struct{})

	go func() {
		err := hs.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			close(listenAndServeFailed)
			klog.Errorf("Unable to start healthz server: %s", err.Error())
		}
	}()

	<-stop
	select {
	case <-listenAndServeFailed:
	// the server didn't start, no need to shutdown
	default:
		hs.shutdown()
	}
}

func (hs *HealthzServer) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := hs.httpServer.Shutdown(ctx)
	if err != nil {
		klog.Errorf("Error shutting down healthz server: %s", err.Error())
	}
}

func (hs *HealthzServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.EscapedPath() == "/healthz" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

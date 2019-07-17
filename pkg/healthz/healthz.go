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

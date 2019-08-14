package healthz

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/klog"
)

var _ = Describe("Coastguard Healthz server", func() {
	klog.InitFlags(nil)

	Context("Server", func() {
		It("Should start and stop", func() {
			healthz := New("127.0.0.1:8123")
			stopCh := make(chan struct{})
			finished := make(chan bool)
			go func() {
				healthz.Run(stopCh)
				finished <- true
			}()

			Expect(finished).ShouldNot(Receive())
			close(stopCh)
			Eventually(finished).Should(Receive(Equal(true)))

		})
	})

	Context("Response handler", func() {

		It("It should respond to GET /healthz request", func() {

			resp := runHealthzRequest("GET", "/healthz")

			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body.String()).To(Equal("OK"))
		})

		It("It should respond to unexpected GET request", func() {
			resp := runHealthzRequest("GET", "/unexpected")

			Expect(resp.Code).To(Equal(http.StatusNotFound))
			Expect(resp.Body.String()).ToNot(Equal("OK"))
		})

	})
})

func runHealthzRequest(method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	resp := httptest.NewRecorder()
	healthz := &HealthzServer{}
	healthz.ServeHTTP(resp, req)
	return resp
}

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coastguard: Healthz suite")
}

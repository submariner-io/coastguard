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

package healthz

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
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
	healthz := &Server{}
	healthz.ServeHTTP(resp, req)

	return resp
}

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coastguard: Healthz suite")
}

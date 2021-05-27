module github.com/submariner-io/coastguard

go 1.12

require (
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.13.0
	github.com/submariner-io/admiral v0.0.0-20190910105841-6188b5ef1d22
	github.com/submariner-io/submariner v0.0.2-0.20190828132721-a11a9a84c90d
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery v0.0.0-20190629003722-e20a3a656cff
	k8s.io/client-go v0.0.0-20190521190702-177766529176
	k8s.io/klog v0.0.0-20181108234604-8139d8cb77af
)

replace github.com/bronze1man/goStrongswanVici => github.com/bronze1man/goStrongswanVici v0.0.0-20190921045355-4c81bd8d0bd5

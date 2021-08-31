module github.com/kubevirt-ui/kube-gateway-operator

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/openshift/api v0.0.0-20210309190949-7d6cac66d2a4
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

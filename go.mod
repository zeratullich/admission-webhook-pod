module admission-webhook-pod

go 1.20

require (
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/pflag v1.0.5
	github.com/wI2L/jsondiff v0.4.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.15
	k8s.io/apimachinery v0.20.15
	k8s.io/client-go v0.20.15
)

require sigs.k8s.io/yaml v1.3.0 // indirect

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.13.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/term v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace k8s.io/api => k8s.io/api v0.20.15

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.15

replace k8s.io/apimachinery => k8s.io/apimachinery v0.20.16-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.20.15

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.15

replace k8s.io/client-go => k8s.io/client-go v0.20.15

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.15

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.15

replace k8s.io/code-generator => k8s.io/code-generator v0.20.16-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.20.15

replace k8s.io/component-helpers => k8s.io/component-helpers v0.20.15

replace k8s.io/controller-manager => k8s.io/controller-manager v0.20.15

replace k8s.io/cri-api => k8s.io/cri-api v0.20.16-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.15

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.15

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.15

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.15

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.15

replace k8s.io/kubectl => k8s.io/kubectl v0.20.15

replace k8s.io/kubelet => k8s.io/kubelet v0.20.15

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.15

replace k8s.io/metrics => k8s.io/metrics v0.20.15

replace k8s.io/mount-utils => k8s.io/mount-utils v0.20.16-rc.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.15

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.15

replace k8s.io/sample-controller => k8s.io/sample-controller v0.20.15

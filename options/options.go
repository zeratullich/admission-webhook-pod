package options

import "github.com/spf13/pflag"

type OptionsParams struct {
	Port       int32
	LogLevel   int32
	Service    string
	Namespace  string
	Kubeconfig string
	IsDebug    bool
	URL        string
}

const (
	MutatePath = "/mutate"
)

func (opts *OptionsParams) FlagParse() {
	pflag.StringVar(&opts.Service, "service", "admission-webhook-pod", "Service in k8s")
	pflag.StringVar(&opts.Namespace, "namespace", "admission-webhook-pod", "Namespace in k8s")
	pflag.StringVar(&opts.Kubeconfig, "kubeconfig", "/root/.kube/config", "K8s configuration file")
	pflag.Int32VarP(&opts.LogLevel, "logLevel", "l", 5, "Log level")
	pflag.BoolVar(&opts.IsDebug, "isDebug", false, "Whether to enable development mode")
	pflag.Int32Var(&opts.Port, "port", 2443, "Webhook service port number")
	pflag.StringVar(&opts.URL, "url", "192.168.0.117", "The development machine address of the webhook")

	pflag.Parse()
}

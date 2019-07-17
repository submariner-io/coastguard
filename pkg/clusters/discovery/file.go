package discovery

import (
	"flag"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/submariner-io/admiral/pkg/federate"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
)

type contextArray []string

var (
	kubeConfig   string
	kubeContexts contextArray
)

func (contexts *contextArray) String() string {
	return strings.Join(*contexts, ",")
}

func (contexts *contextArray) Set(value string) error {
	*contexts = append(*contexts, value)
	return nil
}

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", os.Getenv("KUBECONFIG"),
		"Path to kubeconfig containing embedded authinfo.")
	flag.Var(&kubeContexts, "dp-context", "kubeconfig context for dataplane clusters (use several times).")
}

func Start(callback federate.ClusterEventHandler) {

	for _, context := range kubeContexts {
		restClient, _, err := loadConfig(kubeConfig, context)
		if err != nil {
			klog.Fatalf("Error loading config for context %s: %s", context, err)
		}
		klog.Infof("Discovered cluster %s from parameters", context)
		callback.OnAdd(context, restClient)
	}
}

func loadConfig(configPath, context string) (*restclient.Config, *clientcmdapi.Config, error) {

	errs := []string{}

	for _, config := range strings.Split(configPath, ":") {
		rest_config, client_config, err := loadSingleConfig(config, context)
		if err == nil {
			return rest_config, client_config, nil
		}
		errs = append(errs, err.Error())
	}

	return nil, nil, errors.Errorf("error loading any kubeConfig %s for context %s: [%v]",
		configPath, context, errs)
}

func loadSingleConfig(configPath, context string) (*restclient.Config, *clientcmdapi.Config, error) {
	c, err := clientcmd.LoadFromFile(configPath)

	if err != nil {
		return nil, nil, errors.Errorf("error loading kubeConfig %s: %v", configPath, err.Error())
	}
	if context != "" {
		c.CurrentContext = context
	}

	cfg, err := clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, nil, errors.Errorf("error creating default client config: %v", err.Error())
	}
	return cfg, c, nil
}

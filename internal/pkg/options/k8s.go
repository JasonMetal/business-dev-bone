package options

import (
	"business-dev-bone/pkg/component-base/util/homedir"
	"path/filepath"

	"github.com/spf13/pflag"
)

// K8sOptions contains configuration items related to kubernetes.
type K8sOptions struct {
	KubeConfig             string                 `json:"kubeconfig" mapstructure:"kube-config"` // 默认用户根目录~/.kube/config
	ApiServerAddr          string                 `json:"apiserver-addr" mapstructure:"apiserver-addr"`
	Namespace              string                 `json:"namespace" mapstructure:"namespace"`
	ServerNamespace        string                 `json:"server-namespace" mapstructure:"server-namespace"`
	NodeSelector           map[string]interface{} `json:"node-selector" mapstructure:"node-selector"`
	NodeIPMap              map[string]interface{} `json:"node-ip-map,omitempty" mapstructure:"node-ip-map"`
	DisableWatchGameserver bool                   `json:"disable-watch-gameserver" mapstructure:"disable-watch-gameserver"`
	DisableK8s             bool                   `json:"disable-k8s" mapstructure:"disable-k8s"`
}

type NodeSelectorConfig struct {
	LabelName  string `json:"label-name" mapstructure:"label-name"`
	LabelValue string `json:"label-value" mapstructure:"label-value"`
}

func NewK8sOptions() *K8sOptions {
	home := homedir.HomeDir()
	return &K8sOptions{
		KubeConfig:      filepath.Join(home, ".kube", "config"),
		ApiServerAddr:   "https://kubernetes.default.svc",
		Namespace:       "mmos-gameserver",
		ServerNamespace: "mmos",
		NodeSelector: map[string]interface{}{
			"mmos": "project",
		},
		DisableWatchGameserver: false,
		DisableK8s:             false,
	}
}

func (k *K8sOptions) Validate() []error {
	errs := []error{}
	return errs
}

func (k *K8sOptions) AddFlags(fs *pflag.FlagSet) {

}

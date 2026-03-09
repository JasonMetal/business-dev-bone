package options

import (
	"encoding/json"

	genericoptions "business-dev-bone/internal/pkg/options"
	"business-dev-bone/internal/pkg/server"
	cliflag "business-dev-bone/pkg/component-base/cli/flag"
	"business-dev-bone/pkg/component-base/log"
)

// Options 仅包含通用 HTTP 服务与日志所需项，无业务配置
type Options struct {
	Env                     string                                 `json:"env"    mapstructure:"env"`
	GenericServerRunOptions *genericoptions.ServerRunOptions       `json:"server"   mapstructure:"server"`
	InsecureServing         *genericoptions.InsecureServingOptions `json:"insecure" mapstructure:"insecure"`
	SecureServing           *genericoptions.SecureServingOptions   `json:"secure"   mapstructure:"secure"`
	Log                     *log.Options                           `json:"log"      mapstructure:"log"`
	FeatureOptions          *genericoptions.FeatureOptions         `json:"feature"  mapstructure:"feature"`
	JwtOptions              *genericoptions.JwtOptions             `json:"jwt"      mapstructure:"jwt"`
	MySQLOptions            *genericoptions.MySQLOptions           `json:"mysql"    mapstructure:"mysql"`
	RedisOptions            *genericoptions.RedisOptions           `json:"redis"    mapstructure:"redis"`
	K8sOptions              *genericoptions.K8sOptions             `json:"k8s"      mapstructure:"k8s"`
}

// NewOptions 默认值：HTTP 8080，HTTPS 关闭，适合做新应用起点
func NewOptions() *Options {
	sec := genericoptions.NewSecureServingOptions()
	sec.Required = false
	sec.BindPort = 0
	return &Options{
		Env:                     "dev",
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		InsecureServing:         genericoptions.NewInsecureServingOptions(),
		SecureServing:           sec,
		MySQLOptions:            genericoptions.NewMySQLOptions(),
		RedisOptions:            genericoptions.NewRedisOptions(),
		JwtOptions:              genericoptions.NewJwtOptions(),
		K8sOptions:              genericoptions.NewK8sOptions(),
		Log:                     log.NewOptions(),
		FeatureOptions:          genericoptions.NewFeatureOptions(),
	}
}

// ApplyTo 实现可选的应用层扩展，骨架中无需
func (o *Options) ApplyTo(_ *server.Config) error {
	return nil
}

// Flags 只暴露通用服务、端口、日志、特性开关
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.GenericServerRunOptions.AddFlags(fss.FlagSet("server"))
	o.InsecureServing.AddFlags(fss.FlagSet("insecure"))
	o.SecureServing.AddFlags(fss.FlagSet("secure"))
	o.FeatureOptions.AddFlags(fss.FlagSet("feature"))
	o.Log.AddFlags(fss.FlagSet("log"))
	return fss
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)
	return string(data)
}

// Complete 骨架内可在此补默认值
func (o *Options) Complete() error {
	return nil
}

// Validate 只校验通用项
func (o *Options) Validate() []error {
	var errs []error
	errs = append(errs, o.GenericServerRunOptions.Validate()...)
	errs = append(errs, o.InsecureServing.Validate()...)
	errs = append(errs, o.SecureServing.Validate()...)
	errs = append(errs, o.Log.Validate()...)
	errs = append(errs, o.FeatureOptions.Validate()...)
	return errs
}

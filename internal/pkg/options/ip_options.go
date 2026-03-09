package options

import (
	"github.com/spf13/pflag"
)

type NetworkOptions struct {
	Preferred string `json:"preferred" mapstructure:"preferred"`
}

// NewNetworkOptions create a `zero` value instance.
func NewNetworkOptions() *NetworkOptions {
	return &NetworkOptions{
		Preferred: "10",
	}
}

// Validate verifies flags passed to MySQLOptions.
func (o *NetworkOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *NetworkOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Preferred, "network.preferred", o.Preferred, ""+
		"指定 ip 前缀")

}

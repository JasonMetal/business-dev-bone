package options

import (
	"github.com/spf13/pflag"
)

// MySQLOptions defines options for mysql database.
type KafkaOptions struct {
	Servers                   []string `json:"servers,omitempty"                     mapstructure:"servers"`
	ChannelMessageNotifyTopic string   `json:"channel-message-notify-topic,omitempty"   mapstructure:"channel-message-notify-topic"`
}

// NewMySQLOptions create a `zero` value instance.
func NewKafkaOptions() *KafkaOptions {
	return &KafkaOptions{
		Servers:                   []string{},
		ChannelMessageNotifyTopic: "",
	}
}

// Validate verifies flags passed to MySQLOptions.
func (o *KafkaOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *KafkaOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&o.Servers, "kafka.servers", o.Servers, ""+
		"Kafka service host address. If left blank, the following related kafka options will be ignored.")

	fs.StringVar(&o.ChannelMessageNotifyTopic, "kafka.channel-message-notify-topic", o.ChannelMessageNotifyTopic, ""+
		"Channel message notify topic.")
}

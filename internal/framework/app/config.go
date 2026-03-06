package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"business-dev-bone/pkg/component-base/util/homedir"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configFlagName = "config"
	envFlagName    = "env"
)

var (
	cfgFile string
	envName string
)

// nolint: gochecknoinits
func init() {
	pflag.StringVarP(&cfgFile, "config", "c", cfgFile, "Read configuration from specified `FILE`, "+
		"support JSON, TOML, YAML, HCL, or Java properties formats.")
	pflag.StringVarP(&envName, "env", "e", "dev", "Specify environment: dev, test, release, prod")
}

// addConfigFlag adds flags for a specific server to the specified FlagSet
// object.
func addConfigFlag(basename string, fs *pflag.FlagSet) {
	fs.AddFlag(pflag.Lookup(configFlagName))
	fs.AddFlag(pflag.Lookup(envFlagName))

	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.Replace(strings.ToUpper(basename), "-", "_", -1))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			// 如果指定了配置文件，直接使用
			viper.SetConfigFile(cfgFile)
		} else {
			// 否则按照环境查找配置文件
			// 1. 当前目录
			viper.AddConfigPath(".")

			// 2. configs/环境 目录
			viper.AddConfigPath(filepath.Join("configs", envName))

			// 3. /etc/服务名/环境 目录
			if names := strings.Split(basename, "-"); len(names) > 1 {
				viper.AddConfigPath(filepath.Join(homedir.HomeDir(), "."+names[0], envName))
				viper.AddConfigPath(filepath.Join("/etc", names[0], envName))
			}

			// 设置配置文件名
			viper.SetConfigName(basename)
		}

		if err := viper.ReadInConfig(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: failed to read configuration file(%s): %v\n", cfgFile, err)
			os.Exit(1)
		}
	})
}

func printConfig() {
	if keys := viper.AllKeys(); len(keys) > 0 {
		fmt.Printf("%v Configuration items:\n", progressMessage)
		table := uitable.New()
		table.Separator = " "
		table.MaxColWidth = 80
		table.RightAlign(0)
		for _, k := range keys {
			table.AddRow(fmt.Sprintf("%s:", k), viper.Get(k))
		}
		fmt.Printf("%v", table)
	}
}

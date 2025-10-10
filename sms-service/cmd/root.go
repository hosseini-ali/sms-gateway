package cmd

import (
	"fmt"
	"notif/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(consumeCmd)

}

var (
	// rootCMD represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:              "notif",
		Short:            "A sample application that can be used",
		Long:             "",
		PersistentPreRun: preRun,
	}
)

func preRun(_ *cobra.Command, _ []string) {
	fmt.Println("PRe run")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	if err = viper.Unmarshal(&config.C); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

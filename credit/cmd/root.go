package cmd

import (
	"credit/internal/conf"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:                "credits",
	Short:              "A sample application that can be used",
	PersistentPreRun:   preRun,
	PersistentPostRunE: postRun,
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(persistCmd)
}

func preRun(_ *cobra.Command, _ []string) {
	fmt.Println("pre")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	if err = viper.Unmarshal(&conf.Cfg); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func postRun(_ *cobra.Command, _ []string) error {
	return nil
}

func Execute() error {
	return rootCmd.Execute()
}

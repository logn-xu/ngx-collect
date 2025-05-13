package cmd

import (
	"fmt"
	"os"

	"github.com/logn/ngx-collect/internal/collector"
	"github.com/logn/ngx-collect/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "ngx-collect",
	Short: "ngx-collect is a tool to collect nginx configurations from multiple machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, machine := range config.AppConfig.Machines {
			collecter := collector.NewCollecter(machine)
			if err := collecter.Connect(); err != nil {
				log.Errorf("ssh connect faild: %v", err)
				continue
			}
			if err := collecter.Fetch(); err != nil {
				log.Errorf("ssh fetch faild: %v", err)
				continue
			}
			defer collecter.Close()

		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "config/config.yaml", "config file (default is config/config.yaml)")
}

// InitConfig 初始化配置
func InitConfig() {
	log.Info("configPath: ", configPath)
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("mapstructure")
		viper.AddConfigPath("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read mapstructure file: %v", err)
	}

	if err := viper.Unmarshal(&config.AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal mapstructure file: %v", err)
	}

}

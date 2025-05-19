package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/logn/ngx-collect/internal/collector"
	"github.com/logn/ngx-collect/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/semaphore"
)

var (
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "ngx-collect",
	Short: "ngx-collect is a tool to collect nginx configurations from multiple machines",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.Unmarshal(&config.AppConfig); err != nil {
			log.Fatalf("Failed to unmarshal mapstructure file: %v", err)
			return err
		}
		log.Debug("config: ", config.AppConfig)
		log.Debug("batch size: ", config.AppConfig.BatchSize)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		batchSize := config.AppConfig.BatchSize
		log.Info("batch size: ", batchSize)
		log.Info("timeout: ", config.AppConfig.Timeout)

		wg := sync.WaitGroup{}
		// init semaphore
		sem := semaphore.NewWeighted(int64(batchSize))

		for _, machine := range config.AppConfig.Machines {
			wg.Add(1)
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Errorf("semaphore acquire faild: %v", err)
				return err
			}

			go func(m config.MachineConfig) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.AppConfig.Timeout)*time.Second)
				defer func() {
					sem.Release(1)
					wg.Done()
					cancel()
				}()

				start := time.Now()
				log.Infof("collecting %s-%s", m.Alias, m.Host)

				done := collector.GetNginxConfig(ctx, m)

				select {
				case <-ctx.Done():
					log.Errorf("collecting %s-%s timeout", m.Alias, m.Host)
				case <-done:
					end := time.Now()
					log.Infof("collecting %s-%s done, cost %s", m.Alias, m.Host, end.Sub(start))
				}

			}(machine)

		}
		wg.Wait()

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
	rootCmd.PersistentFlags().IntVarP(&config.AppConfig.BatchSize, "batch-size", "b", 5, "batch size (default is 5)")
	rootCmd.PersistentFlags().IntVarP(&config.AppConfig.Timeout, "timeout", "t", 5, "timeout (default is 5s)")
}

// InitConfig 初始化配置
func InitConfig() {
	log.Info("configPath: ", configPath)
	viper.SetDefault("batch_size", 5)
	viper.SetDefault("timeout", 5)

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

	if err := viper.BindPFlag("batch_size", rootCmd.PersistentFlags().Lookup("batch-size")); err != nil {
		log.Fatalf("Failed to bind flag: %v", err)
	}

	if err := viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout")); err != nil {
		log.Fatalf("Failed to bind flag: %v", err)
	}

	if err := viper.Unmarshal(&config.AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal mapstructure file: %v", err)
	}

}

package config

type MachineConfig struct {
	Alias            string `mapstructure:"alias"`
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	User             string `mapstructure:"user"`
	AuthMethod       string `mapstructure:"auth_method"`
	KeyPath          string `mapstructure:"key_path"`
	Password         string `mapstructure:"password"`
	NginxExecBin     string `mapstructure:"nginx_exec_bin"`
	NginxMainConfig  string `mapstructure:"nginx_main_config"`
	LocalDestination string `mapstructure:"local_destination"`
}

type Config struct {
	Machines  []MachineConfig `mapstructure:"machines"`
	BatchSize int             `mapstructure:"batch_size"`
	Timeout   int             `mapstructure:"timeout"`
}

var AppConfig Config

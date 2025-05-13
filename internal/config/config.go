package config

type MachineConfig struct {
	Alias            string   `mapstructure:"alias"`
	Host             string   `mapstructure:"host"`
	Port             int      `mapstructure:"port"`
	User             string   `mapstructure:"user"`
	AuthMethod       string   `mapstructure:"auth_method"`
	KeyPath          string   `mapstructure:"key_path"`
	Password         string   `mapstructure:"password"`
	RemotePaths      []string `mapstructure:"remote_paths"`
	LocalDestination string   `mapstructure:"local_destination"`
}

type Config struct {
	Machines []MachineConfig `mapstructure:"machines"`
}

var AppConfig Config

package config

import (
	"time"
)

type AppConfig struct {
	ServerConfig ServerConfig `mapstructure:"server"`
	LoggerConfig loggerConfig `mapstructure:"logger"`
	DBConfig     dbConfig     `mapstructure:"db_config"`
	RetrysConfig RetrysConfig `mapstructure:"retry_strategy"`
	GinConfig    ginConfig    `mapstructure:"gin"`
}

type RetrysConfig struct {
	Attempts int           `mapstructure:"attempts" default:"3"`
	Delay    time.Duration `mapstructure:"delay" default:"1s"`
	Backoffs float64       `mapstructure:"backoffs" default:"2"`
}

type ginConfig struct {
	Mode string `mapstructure:"mode" default:"debug"`
}

type ServerConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
	Port int    `mapstructure:"port" default:"8080"`
}

type loggerConfig struct {
	Level string `mapstructure:"level" default:"info"`
}

type postgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string
	Password string
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode" default:"disable"`
}

type dbConfig struct {
	Master          postgresConfig   `mapstructure:"postgres"`
	Slaves          []postgresConfig `mapstructure:"slaves"`
	MaxOpenConns    int              `mapstructure:"max_open_conns"`
	MaxIdleConns    int              `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration    `mapstructure:"conn_max_lifetime"`
}

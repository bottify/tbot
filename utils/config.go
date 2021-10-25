package utils

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel      string `yaml:"log_level"`
	SuperUsers    string `yaml:"super_users"`
	CommandPrefix string `yaml:"cmd_prefix"`
	RuntimePath   string `yaml:"runtime_path"`
	DBFile        string `yaml:"db_file"`

	superUsers []string `yaml:"-"`
}

var config Config

func GetConfig() *Config {
	return &config
}

func (cfg *Config) Init(file string) {
	cfg_str, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("open [%v] failed! %v", file, err)
	} else {
		err = yaml.Unmarshal(cfg_str, cfg)
		if err != nil {
			log.Errorf("parse [%v] failed! %v", file, err)
		}
	}

	// fix default value
	if len(cfg.CommandPrefix) == 0 {
		cfg.CommandPrefix = "%"
	}
	if len(cfg.LogLevel) == 0 {
		cfg.LogLevel = "debug"
	}
	cfg.superUsers = strings.Split(cfg.SuperUsers, ",")
}

func (cfg *Config) GetSuperUsers() []string {
	return cfg.superUsers
}

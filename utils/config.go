package utils

import (
	"fmt"
	"os"
	"path/filepath"
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
	ApiTimeout    int64  `yaml:"api_timeout"`

	superUsers []string `yaml:"-"`
	inited     bool
}

var config Config

func GetConfig() *Config {
	return &config
}

func (cfg *Config) Init(file string) error {
	cfg_str, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("open [%v] failed! %v", file, err)
		return err
	} else {
		err = yaml.Unmarshal(cfg_str, cfg)
		if err != nil {
			log.Errorf("parse [%v] failed! %v", file, err)
			return err
		}
	}

	// fix default value
	if len(cfg.CommandPrefix) == 0 {
		cfg.CommandPrefix = "%"
	}
	if len(cfg.LogLevel) == 0 {
		cfg.LogLevel = "debug"
	}
	if cfg.ApiTimeout == 0 {
		cfg.ApiTimeout = 30
	}
	cfg.superUsers = strings.Split(cfg.SuperUsers, ",")

	log.Info("will mkdir: ", fmt.Sprintf("%v/data/tbot", cfg.RuntimePath))
	err = os.MkdirAll(fmt.Sprintf("%v/data/tbot", cfg.RuntimePath), os.FileMode(0770))
	if err != nil {
		log.Error("mkdir: ", err)
	}
	cfg.inited = true
	return nil
}

func (cfg *Config) GetDataPath(filename string) string {
	// incase of corrupt filesystem
	if !cfg.inited {
		panic("config not inited!")
	}
	path, err := filepath.Abs(fmt.Sprintf("%v/data/tbot/%v", cfg.RuntimePath, filename))
	if err != nil {
		panic(err)
	}
	return path
}

func (cfg *Config) GetSuperUsers() []string {
	return cfg.superUsers
}

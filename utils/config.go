package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	EpicQuota     string `yaml:"epic_quota"`

	epicQuota  map[int64]int64 `yaml:"-"`
	superUsers []string        `yaml:"-"`
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
	cfg.epicQuota = make(map[int64]int64)

	for _, group_quota := range strings.Split(cfg.EpicQuota, ",") {
		group_quota_pair := strings.Split(group_quota, ":")
		if len(group_quota_pair) != 2 {
			log.Errorf("invalid epic quota [%v]", group_quota)
			continue
		}
		group, err := strconv.ParseInt(group_quota_pair[0], 10, 64)
		if err != nil {
			log.Errorf("invalid epic quota [%v]", group_quota)
			continue
		}
		quota, err := strconv.ParseInt(group_quota_pair[1], 10, 64)
		if err != nil {
			log.Errorf("invalid epic quota [%v]", group_quota)
			continue
		}
		cfg.epicQuota[group] = quota
	}
	log.Info("epic group quota: ", cfg.epicQuota)

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

func (cfg *Config) GetGroupEpicQuota(group int64) (int64, bool) {
	if !cfg.inited {
		panic("config not inited!")
	}
	quota, ok := cfg.epicQuota[group]
	return quota, ok
}

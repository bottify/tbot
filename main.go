package main

import (
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"gopkg.in/yaml.v3"

	_ "tbot/cmd/roll"
)

type Config struct {
	SuperUsers string `yaml:"super_users"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	zero.OnCommand("ping").Handle(func(ctx *zero.Ctx) {
		ctx.Send("pong")
	})

	cfg_str, err := os.ReadFile("config.yaml")
	cfg := &Config{}
	if err != nil {
		log.Error("config.yaml not exists! using default! ", err)
	} else {
		err = yaml.Unmarshal(cfg_str, cfg)
		if err != nil {
			log.Error("parse config.yaml failed! using default! ", err)
		}
	}
	log.Infof("using config: %+v", cfg)

	// TODO: read config from cfg
	zero.Run(zero.Config{
		NickName:      []string{"tbot"},
		CommandPrefix: "%",
		SuperUsers:    strings.Split(cfg.SuperUsers, ","),
		Driver: []zero.Driver{
			driver.NewWebSocketClient("ws://127.0.0.1:6700", "wTgZb5TsiTqmaOYT"),
		},
	})
	select {}
}

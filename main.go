package main

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	_ "tbot/plugin/epicture"
	_ "tbot/plugin/maimai"
	_ "tbot/plugin/roll"
	"tbot/utils"
	"tbot/utils/db"
)

var startAt time.Time

func main() {
	startAt = time.Now()
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	zero.OnCommand("ping").Handle(func(ctx *zero.Ctx) {
		ctx.Send("pong")
	})
	zero.OnCommand("pingex").Handle(func(ctx *zero.Ctx) {
		ctx.Send(fmt.Sprint("pong, bot started at ", startAt.Format("2006-01-02 15:04:05")))
	})

	cfg := utils.GetConfig()
	cfg.Init("config.yaml")
	db.InitDB()

	log.Infof("using config: %+v", cfg)
	lv, _ := log.ParseLevel(cfg.LogLevel)
	log.SetLevel(lv)

	zero.Run(zero.Config{
		NickName:      []string{"tbot"},
		CommandPrefix: cfg.CommandPrefix,
		SuperUsers:    cfg.GetSuperUsers(),
		Driver: []zero.Driver{
			driver.NewWebSocketClient("ws://127.0.0.1:6700", "wTgZb5TsiTqmaOYT", cfg.ApiTimeout),
		},
	})
	select {}
}

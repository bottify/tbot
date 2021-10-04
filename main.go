package main

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	_ "tbot/cmd/roll"
)

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

	zero.Run(zero.Config{
		NickName:      []string{"tbot"},
		CommandPrefix: "%",
		SuperUsers:    []string{"876472013"},
		Driver: []zero.Driver{
			driver.NewWebSocketClient("ws://127.0.0.1:6700", "wTgZb5TsiTqmaOYT"),
		},
	})
	select {}
}

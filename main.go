package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"gopkg.in/yaml.v3"

	_ "tbot/plugin/roll"
	"tbot/utils/msg"
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

	zero.OnCommand("b40").Handle(func(ctx *zero.Ctx) {
		args := ctx.State["args"].(string)
		var cmd *exec.Cmd
		if len(args) > 0 {
			cmd = exec.Command("python", "lib3rd/maimai_b40/entry.py", "username", args)
		} else {
			cmd = exec.Command("python", "lib3rd/maimai_b40/entry.py", "qq", fmt.Sprint(ctx.Event.Sender.ID))
		}
		bo := &bytes.Buffer{}
		cmd.Stdout = bo
		err := cmd.Run()
		if err != nil {
			log.Error("spwan python error: ", err)
			ctx.Send("执行出错，请稍后再试...")
			return
		}
		if len(bo.Bytes()) < 50 {
			ctx.Send("获取信息失败, 请确认已绑定用户并导入成绩")
			return
		}
		ctx.Send(msg.New().ImageBytes(bo.Bytes()))
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
		CommandPrefix: "!",
		SuperUsers:    strings.Split(cfg.SuperUsers, ","),
		Driver: []zero.Driver{
			driver.NewWebSocketClient("ws://127.0.0.1:6700", "wTgZb5TsiTqmaOYT"),
		},
	})
	select {}
}

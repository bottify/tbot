package rules

import (
	"fmt"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func CommandWithReply(cmd string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		has_reply := false
		for _, msg := range ctx.Event.Message {
			if msg.Type == "reply" {
				has_reply = true
			}
			if msg.Type == "reply" || msg.Type == "at" {
				continue
			}
			if msg.Type == "text" && has_reply {
				text := msg.Data["text"]
				if len(strings.TrimSpace(text)) == 0 {
					continue
				}
				return strings.Fields(text)[0] == fmt.Sprint(zero.BotConfig.CommandPrefix, cmd)
			}
			return false
		}
		return false
	}
}

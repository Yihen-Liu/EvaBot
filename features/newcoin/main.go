/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-05-09
 */
package newcoin

import (
	"github.com/riversgo007/EvaBot/common/log"
	"github.com/riversgo007/EvaBot/core"
	"os"
	"sync"
	"time"
)

var passedUsers = sync.Map{}

func RunService() {
	log.InitLog(log.InfoLog, os.Stdout, log.PATH)

	bot, err := core.NewBotAPI("1723000333:AAFeplI74Q2ZP7IVoddxlXr0ODelLzKjez4")
	if err != nil {
		log.Error(err)
	}

	bot.Debug = true

	log.Infof("Authorized on account %s", bot.Self.UserName)

	u := core.NewUpdate(0)
	u.Timeout = 60

	u.AllowedUpdates = []string{core.UpdateTypeChatMember, core.UpdateTypeMessage}

	updates := bot.GetUpdatesChan(u)

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	go broadcast(bot)

	for update := range updates {
		if update.ChatMember !=nil{
			if update.ChatMember.From.IsBot && update.ChatMember.From.UserName=="newcoin"{
				log.Infof("chat.id:%d, username:%s",update.ChatMember.Chat.ID, update.ChatMember.From.UserName)

				passedUsers.Store(update.ChatMember.Chat.ID, true)
			}
		}
		if update.Message == nil {
			continue
		}

		//update.Message.
		if update.Message.IsCommand() {
			msg := core.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "help":
				msg.Text = "type /newcoin [ito/ido]"
				_, _ = bot.Send(msg)

			case "newcoin":
				arg:= update.Message.CommandArguments()

				if arg=="ito"{
					msg.Text = "evanesco ito will coming soon."
				}

				if arg == "ido"{
					msg.Text = "evanesco ito will coming soon."
				}

				if arg == ""{
					msg.Text = "please type /newcoin [ito/ido]; for example: /newcoin ito"
				}
				_, _ = bot.Send(msg)

			default:
				msg.Text = "I don't know that command"
				_, _ = bot.Send(msg)
			}
		}

	}
}

func broadcast(bot *core.BotAPI) {
	t := time.NewTicker(time.Second*10)
	for {
		select {
		case <-t.C:
			passedUsers.Range(func(key, value interface{}) bool {
				if value.(bool)==true{
					msg := core.NewMessage(key.(int64), "broadcat from newcoin bot")
					_, _ = bot.Send(msg)
				}
				return true
			})
		}
	}
}

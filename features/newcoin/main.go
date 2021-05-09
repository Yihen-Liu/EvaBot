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
	"time"
)

func RunService() {
	log.InitLog(log.InfoLog, os.Stdout, log.PATH)

	bot, err := core.NewBotAPI("1706227378:AAEAjAhUS6IAoDLiR8UAAcwH2sJM1qWSnQY")
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


	for update := range updates {

		if update.Message == nil {
			continue
		}

		log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
		log.Infof("chat-id:%d, chat-name:%s", update.Message.Chat.ID, update.Message.Chat.UserName)

		//update.Message.
		if update.Message.IsCommand() {
			msg := core.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "help":
				msg.Text = "type /url or /count"
				_, _ = bot.Send(msg)

			case "withArgument":
				msg.Text = "You supplied the following argument: " + update.Message.CommandArguments()
				_, _ = bot.Send(msg)

			case "newcoin":
				msg.Text = "bot is private, don't support this group."
				_, _ = bot.Send(msg)

			default:
				msg.Text = "I don't know that command"
				_, _ = bot.Send(msg)
			}
		}

	}
}

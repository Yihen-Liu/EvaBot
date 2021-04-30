/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-04-30
 */
package main
import (
	"encoding/json"
	"github.com/riversgo007/EvaBot/core"
	"log"
	"time"
)

func main() {
	bot, err := core.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := core.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			msg := core.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "help":
				msg.Text = "type /sayhi or /status."
			case "sayhi":
				msg.Text = "Hi :)"
			case "status":
				msg.Text = "I'm ok."
			case "withArgument":
				msg.Text = "You supplied the following argument: " + update.Message.CommandArguments()
			case "url":
				response, err:=bot.Request(core.NewCreateChatInviteLink(update.Message.Chat.ID))
				if err!=nil{
					log.Println("")
					continue
				}

				data, err:=response.Result.MarshalJSON()

				var resp struct {
					Result *core.ChatInviteLink
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					continue
				}


				msg.Text = resp.Result.InviteLink
			default:
				msg.Text = "I don't know that command"
			}
			bot.Send(msg)
		}

	}
}
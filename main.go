/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-04-30
 */
package main

import (
	"encoding/json"
	"fmt"
	"github.com/riversgo007/EvaBot/common/log"
	"github.com/riversgo007/EvaBot/core"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"time"
)

var (
	DB *gorm.DB
)

type URL struct {
	ID     		int    `json:"id"`
	URLValue 	string `json:"url_value"`
	Status 		int   `json:"status"`
	UserName	string `json:"user_name"`
	UserId		string `json:"user_id"`
	CreateTime  int64   `json:"create_time"`
	UpdateTime 	int64 	`json:"update_time"`
}

func initMysql() (err error) {
	//初始化数据库连接
	dns := "eva:evanetwork.org@(localhost:3306)/telegram?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		return
	} else {
		log.Info("create mysql connection successed.")
	}
	return
}

func main() {
	log.InitLog(log.InfoLog, os.Stdout, log.PATH)

	bot, err := core.NewBotAPI("1706227378:AAE1i8R1CgpG_WJd7wr6y_bned_ggt0NJ7o")
	if err != nil {
		log.Error(err)
	}

	bot.Debug = true

	log.Infof("Authorized on account %s", bot.Self.UserName)

	u := core.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	if err:=initMysql();err!=nil{
		log.Fatal("init mysql err:", err.Error())
		return
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			msg := core.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "help":
				msg.Text = "type /url or /count."
				_, _ = bot.Send(msg)

			case "withArgument":
				msg.Text = "You supplied the following argument: " + update.Message.CommandArguments()
				_, _ = bot.Send(msg)

			case "url":
				go func() {

					var u URL
					if err := DB.Where("user_name=?", update.Message.From.UserName).First(&u).Error; err == nil {
						msg.Text = fmt.Sprintf("%s has been DONE, url:%s, don't repeated.", update.Message.From, u.URLValue)
						return
					}

					response, err := bot.Request(core.NewCreateChatInviteLink(update.Message.Chat.ID))
					if err != nil {
						log.Error("bot.Rrequest error:", err.Error())
						return
					}

					var inviteLink core.ChatInviteLink
					if err := json.Unmarshal(response.Result, &inviteLink); err != nil {
						return
					}

					msg.Text = fmt.Sprintf("allocate to:%s, url:%s", update.Message.From, inviteLink.InviteLink)

					url := URL{
						URLValue:   inviteLink.InviteLink,
						UserName:   update.Message.From.UserName,
						Status:     1,
						UserId:     string(update.Message.From.ID),
						CreateTime: time.Now().Unix(),
						UpdateTime: time.Now().Unix(),
					}

					if err := DB.Create(&url).Error; err != nil {
						msg.Text = "server timeout"
						log.Error("create url err:", err.Error(), ",user:", url.UserName)
					}

					_, _ = bot.Send(msg)

				}()
			default:
				msg.Text = "I don't know that command"
				_, _ = bot.Send(msg)
			}
		}

	}
}
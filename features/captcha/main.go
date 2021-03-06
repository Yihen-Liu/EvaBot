package captcha

import (
	"fmt"
	"github.com/riversgo007/EvaBot/features/invite"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/net/proxy"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Config struct for toml config file
type Config struct {
	ButtonText          string `mapstructure:"button_text"`
	WelcomeMessage      string `mapstructure:"welcome_message"`
	AfterSuccessMessage string `mapstructure:"after_success_message"`
	AfterFailMessage    string `mapstructure:"after_fail_message"`
	PrintSuccessAndFail string `mapstructure:"print_success_and_fail_messages_strategy"`
	WelcomeTimeout      string `mapstructure:"welcome_timeout"`
	BanDurations        string `mapstructure:"ban_duration"`
	UseSocks5Proxy      string `mapstructure:"use_socks5_proxy"`
	Socks5Address       string `mapstructure:"socks5_address"`
	Socks5Port          string `mapstructure:"socks5_port"`
	Socks5Login         string `mapstructure:"socks5_login"`
	Socks5Password      string `mapstructure:"socks5_password"`
}

var config Config
var passedUsers = sync.Map{}
var bot *tb.Bot
var tgtoken = "TGTOKEN"
var configPath = "./"

func initConfig() {
	err := readConfig()
	if err != nil {
		log.Fatalf("Cannot read config file. Error: %v", err)
	}
}

func RunService() {
	initConfig()

	token, err := getToken(tgtoken)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Telegram Bot Token [%v] successfully obtained from env variable $TGTOKEN\n", token)

	var httpClient *http.Client
	if config.UseSocks5Proxy == "yes" {
		var err error
		httpClient, err = initSocks5Client()
		if err != nil {
			log.Fatalln(err)
		}
	}

	bot, err = tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		Client: httpClient,
	})
	if err != nil {
		log.Fatalf("Cannot start bot. Error: %v\n", err)
	}

	bot.Handle(tb.OnUserJoined, removeRestrict)
	//bot.Handle(tb.OnCallback, passChallenge)
	//bot.Handle(tb.OnUserLeft, leaveAway)
	bot.Handle("/healthz", func(m *tb.Message) {
		msg := "I'm OK"
		if _, err := bot.Send(m.Chat, msg); err != nil {
			log.Println(err)
		}
		log.Printf("Healthz request from user: %v\n in chat: %v", m.Sender, m.Chat)
	})

	log.Println("Bot started!")
	go func() {
		bot.Start()
	}()
}

func leaveAway(m *tb.Message) {
	if m.UserLeft.ID != m.Sender.ID {
		return
	}

	log.Printf("User: %v leave away the chat: %v", m.UserLeft, m.Chat)

	//update invitors database
	var invitor invite.Invitor
	err := invite.DB.Where("group_name=? and user_id=?",m.Chat.Title, fmt.Sprintf("%d",m.Sender.ID)).First(&invitor).Error

	if err == nil {
		if _err := invite.DB.Model(&invitor).Updates(map[string]interface{}{"status": int8(0),"update_time": time.Now().Unix()}).Error; _err != nil {
			log.Println("update use invited success err:",_err.Error())
		}
	}
}

func removeRestrict(m *tb.Message) {
	if m.UserJoined.ID != m.Sender.ID {
		return
	}
	log.Printf("User: %v joined the chat: %v", m.UserJoined, m.Chat)

	if member, err := bot.ChatMemberOf(m.Chat, m.UserJoined); err == nil {
		if member.Role == tb.Restricted {
			log.Printf("User: %v already restricted in chat: %v", m.UserJoined, m.Chat)
			//if there is somebody was ban forever, open the door again.
			newChatMember := tb.ChatMember{User: m.UserJoined, RestrictedUntil: tb.Forever(), Rights: tb.Rights{CanSendMessages: true}}
			err := bot.Promote(m.Chat, &newChatMember)
			if err != nil {
				log.Println(err)
			}
			//return
		}
	}
	return
}

func challengeUser(m *tb.Message) {
	if m.UserJoined.ID != m.Sender.ID {
		return
	}
	log.Printf("User: %v joined the chat: %v", m.UserJoined, m.Chat)

	if member, err := bot.ChatMemberOf(m.Chat, m.UserJoined); err == nil {
		if member.Role == tb.Restricted {
			log.Printf("User: %v already restricted in chat: %v", m.UserJoined, m.Chat)
			//if there is somebody was ban forever, open the door again.
			newChatMember := tb.ChatMember{User: m.UserJoined, RestrictedUntil: tb.Forever(), Rights: tb.Rights{CanSendMessages: true}}
			err := bot.Promote(m.Chat, &newChatMember)
			if err != nil {
				log.Println(err)
			}
			//return
		}
	}

	newChatMember := tb.ChatMember{User: m.UserJoined, RestrictedUntil: tb.Forever(), Rights: tb.Rights{CanSendMessages: false}}
	err := bot.Restrict(m.Chat, &newChatMember)
	if err != nil {
		log.Println(err)
	}

	inlineKeys := [][]tb.InlineButton{{tb.InlineButton{
		Unique: "challenge_btn",
		Text:   m.UserJoined.Username+config.ButtonText,
	}}}
	challengeMsg, _ := bot.Reply(m, config.WelcomeMessage, &tb.ReplyMarkup{InlineKeyboard: inlineKeys})

	n, err := strconv.ParseInt(config.WelcomeTimeout, 10, 64)
	if err != nil {
		log.Println(err)
	}
	time.AfterFunc(time.Duration(n)*time.Second, func() {
		_, passed := passedUsers.Load(m.UserJoined.ID)
		if !passed {
			banDuration, e := getBanDuration()
			if e != nil {
				log.Println(e)
			}
			chatMember := tb.ChatMember{User: m.UserJoined, RestrictedUntil: banDuration}
			err := bot.Ban(m.Chat, &chatMember)
			if err != nil {
				log.Println(err)
			}

			if config.PrintSuccessAndFail == "show" {
				_, err := bot.Edit(challengeMsg, m.UserJoined.Username+config.AfterFailMessage)
				if err != nil {
					log.Println(err)
				}
			} else if config.PrintSuccessAndFail == "del" {
				err := bot.Delete(m)
				if err != nil {
					log.Println(err)
				}
				err = bot.Delete(challengeMsg)
				if err != nil {
					log.Println(err)
				}
			}

			log.Printf("User: %v was banned in chat: %v for: %v minutes", m.UserJoined, m.Chat, config.BanDurations)
		}
		passedUsers.Delete(m.UserJoined.ID)
	})
}

// passChallenge is used when user passed the validation
func passChallenge(c *tb.Callback) {
	if c.Message == nil || c.Message.ReplyTo==nil|| c.Message.ReplyTo.Sender==nil || c.Sender==nil{
		err := bot.Respond(c, &tb.CallbackResponse{Text: "This button is incorrect for you"})
		if err != nil {
			log.Println(err)
		}
		return
	}

	if c.Message.ReplyTo.Sender.ID != c.Sender.ID {
		err := bot.Respond(c, &tb.CallbackResponse{Text: "This button isn't for you"})
		if err != nil {
			log.Println(err)
		}
		return
	}
	passedUsers.Store(c.Sender.ID, struct{}{})

	if config.PrintSuccessAndFail == "show" {
		_, err := bot.Edit(c.Message, c.Sender.Username+config.AfterSuccessMessage)
		if err != nil {
			log.Println(err)
		}
	} else if config.PrintSuccessAndFail == "del" {
		err := bot.Delete(c.Message)
		if err != nil {
			log.Println(err)
		}
	}

	log.Printf("User: %v passed the challenge in chat: %v", c.Sender, c.Message.Chat)
	newChatMember := tb.ChatMember{User: c.Sender, RestrictedUntil: tb.Forever(), Rights: tb.Rights{CanSendMessages: true}}
	err := bot.Promote(c.Message.Chat, &newChatMember)
	if err != nil {
		log.Println(err)
	}

	//update invitors database
	var invitor invite.Invitor
	err = invite.DB.Where("group_name=? and user_id=?",c.Message.Chat.Title, fmt.Sprintf("%d",c.Sender.ID)).First(&invitor).Error

	if err == nil {
		if _err := invite.DB.Model(&invitor).Updates(map[string]interface{}{"status": int8(1),"update_time": time.Now().Unix()}).Error; _err != nil {
			log.Println("update use invited success err:",_err.Error())
		}
	}

	err = bot.Respond(c, &tb.CallbackResponse{Text: "Validation passed!"})
	if err != nil {
		log.Println(err)
	}
}

func readConfig() (err error) {
	v := viper.New()
	path, ok := os.LookupEnv(configPath)
	if ok {
		v.SetConfigName("config")
		v.AddConfigPath(path)
	}
	v.SetConfigName("config")
	v.AddConfigPath(".")

	if err = v.ReadInConfig(); err != nil {
		return err
	}
	if err = v.Unmarshal(&config); err != nil {
		return err
	}
	return
}

func getToken(key string) (string, error) {
	token, ok := os.LookupEnv(key)
	if !ok {
		err := errors.Errorf("Env variable %v isn't set!", key)
		return "", err
	}
	match, err := regexp.MatchString(`^[0-9]+:.*$`, token)
	if err != nil {
		return "", err
	}
	if !match {
		err := errors.Errorf("Telegram Bot Token [%v] is incorrect. Token doesn't comply with regexp: `^[0-9]+:.*$`. Please, provide a correct Telegram Bot Token through env variable TGTOKEN", token)
		return "", err
	}
	return token, nil
}

func getBanDuration() (int64, error) {
	if config.BanDurations == "forever" {
		return tb.Forever(), nil
	}

	n, err := strconv.ParseInt(config.BanDurations, 10, 64)
	if err != nil {
		return 0, err
	}

	return time.Now().Add(time.Duration(n) * time.Minute).Unix(), nil
}

func initSocks5Client() (*http.Client, error) {
	addr := fmt.Sprintf("%s:%s", config.Socks5Address, config.Socks5Port)
	dialer, err := proxy.SOCKS5("tcp", addr, &proxy.Auth{User: config.Socks5Login, Password: config.Socks5Password}, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("cannot init socks5 proxy client dialer: %w", err)
	}

	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial

	return httpClient, nil
}

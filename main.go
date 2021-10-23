package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/crisp-im/go-crisp-api/crisp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/l0sted/Crisp_Telegram_bot/utils"
	"github.com/spf13/viper"
)

var bot *tgbotapi.BotAPI
var client *crisp.Client

// var redisClient *redis.Client
var config *viper.Viper

// CrispMessageInfo stores the original message
type CrispMessageInfo struct {
	WebsiteID string
	SessionID string
}

// MarshalBinary serializes CrispMessageInfo into binary
func (s *CrispMessageInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalBinary deserializes CrispMessageInfo into struct
func (s *CrispMessageInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

//from crisp to tg
func sendMsgToAdmins(text string, WebsiteID string, SessionID string) {
	for _, id := range config.Get("admins").([]interface{}) {
		msg := tgbotapi.NewMessage(int64(id.(int)), text)
		msg.ParseMode = "Markdown"
		sent, _ := bot.Send(msg)
		log.Println(strconv.Itoa(sent.MessageID))
	}
}

func main() {
	config = utils.GetConfig()

	var chat_prefix = config.GetString("prefix")

	var err error
	log.Printf("Initializing Bot...")

	bot, err = tgbotapi.NewBotAPI(config.GetString("telegram.key"))

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.GetBool("debug")
	bot.RemoveWebhook()

	log.Printf("Authorized on account %s", bot.Self.UserName)

	log.Printf("Initializing Crisp Listner")
	client = crisp.New()
	// Set authentication parameters
	// client.Authenticate(config.GetString("crisp.identifier"), config.GetString("crisp.key"))
	client.AuthenticateTier("plugin", config.GetString("crisp.identifier"), config.GetString("crisp.key"))

	// Connect to realtime events backend and listen (only to 'message:send' namespace)
	client.Events.Listen(
		[]string{
			"message:send",
		},

		func(reg *crisp.EventsRegister) {
			// Socket is connected: now listening for events

			// Notice: if the realtime socket breaks at any point, this function will be called again upon reconnect (to re-bind events)
			// Thus, ensure you only use this to register handlers

			// Register handler on 'message:send/text' namespace
			reg.On("message:send/text", func(evt crisp.EventsReceiveTextMessage) {
				text := fmt.Sprintf(`(%s) *%s(%s): *%s`, chat_prefix, *evt.User.Nickname, *evt.User.UserID, *evt.Content)
				sendMsgToAdmins(text, *evt.WebsiteID, *evt.SessionID)
			})

			// Register handler on 'message:send/file' namespace
			reg.On("message:send/file", func(evt crisp.EventsReceiveFileMessage) {
				text := fmt.Sprintf(`(%s) *%s(%s): *[File](%s)`, chat_prefix, *evt.User.Nickname, *evt.User.UserID, evt.Content.URL)
				sendMsgToAdmins(text, *evt.WebsiteID, *evt.SessionID)
			})

			// Register handler on 'message:send/animation' namespace
			reg.On("message:send/animation", func(evt crisp.EventsReceiveAnimationMessage) {
				text := fmt.Sprintf(`(%s) *%s(%s): *[Animation](%s)`, chat_prefix, *evt.User.Nickname, *evt.User.UserID, evt.Content.URL)
				sendMsgToAdmins(text, *evt.WebsiteID, *evt.SessionID)
			})
		},

		func() {
			log.Printf("Crisp listener disconnected, reconnecting...")
		},

		func() {
			log.Fatal("Crisp listener error, check your API key or internet connection?")
		},
	)
	for {
		time.Sleep(1 * time.Second)
	}
}

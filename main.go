package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	token string
	bot   *tgbotapi.BotAPI
)

func main() {
	ParseArgs()

	log.Printf("Use telegram token `%s`", ArgBotToken)
	var err error
	bot, err = tgbotapi.NewBotAPI(ArgBotToken)
	if err != nil {
		log.Panic(err)
	}

	if err := dbOpen(); err != nil {
		log.Panic(err)
	}
	defer dbClose()

	// bot.Debug = true

	log.Printf("Authorized on account `%s`", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 180

	cmds := []tgbotapi.BotCommand{
		{
			Command:     "help",
			Description: "/help – list all commands"},
		{
			Command:     "photo",
			Description: "/photo - get last image"},
		{
			Command:     "list",
			Description: "/list – all registered chats"},
		{
			Command:     "register",
			Description: "/register token – register current chat for notification updates"},
		{
			Command:     "leave",
			Description: "/leave – do not receive further notifications"},
		{
			Command:     "remove",
			Description: "/remove chatid – notification abo"},
		{
			Command:     "admin",
			Description: "/admin chatid – toggle admin state"},
		{
			Command:     "token",
			Description: "/token – generate and show token"},
	}

	bot.SetMyCommands(cmds)

	genToken()
	go RunServer()
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = `Commands:
/photo - get last photo
/register accesstoken - register new chat with accesstoken
/list - list all registered chats
/leave - leave current chat
/remove chat accesstoken - unregister chat with accesstoken 
`
			case "register":
				msg.Text = cmdRegister(update)
			case "list":
				// msg.ParseMode = tgbotapi.ModeMarkdownV2
				msg.Text = cmdListUsers(update)
			case "leave":
				msg.Text = cmdLeave(update)
			case "remove":
				msg.Text = cmdRemove(update)
			case "photo":
				msg.Text = cmdSendPhoto(update)
			case "admin":
				msg.Text = cmdAdminToggle(update)
			case "token":
				msg.Text = cmdToken(update)
			default:
				msg.Text = "available commands: /help, /register, /list, /leave, /remove"
			}
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}

func genToken() string {

	token = generatePassword(16, 0, 2, 2)
	log.Printf("current access token: `%s`", token)

	return token
}

func cmdListUsers(update tgbotapi.Update) string {
	return dbListUsers()
}

func cmdRegister(update tgbotapi.Update) string {
	if update.Message.CommandArguments() == token {
		if err := dbRegisterChat(update.Message.Chat.ID, update.Message.Chat.UserName, update.Message.Chat.FirstName, update.Message.Chat.LastName, token); err != nil {
			return "Error: " + err.Error()
		}
		genToken()
		return fmt.Sprintf("Successfully registered: %xd `%s` (%s, %s)",
			update.Message.Chat.ID,
			update.Message.Chat.UserName, update.Message.Chat.FirstName, update.Message.Chat.LastName)
	}
	return "Invalid access token!"
}

func cmdToken(update tgbotapi.Update) string {
	if dbIsAdmin(update.Message.Chat.ID) {
		genToken()
		return token
	}
	return "Kein Admin - kein Token!"
}

func cmdAdminToggle(update tgbotapi.Update) string {
	if dbIsAdmin(update.Message.Chat.ID) {
		if chatid, err := strconv.ParseInt(update.Message.CommandArguments(), 10, 64); err != nil {
			return "Invalid chat id"
		} else if err = dbToggleAdmin(chatid); err != nil {
			return err.Error()
		} else {
			return cmdListUsers(update)
		}
	}
	return "Kein Admin - keine Aktion"
}

func cmdLeave(update tgbotapi.Update) string {
	return dbLeave(update.Message.Chat.ID)
}

func cmdRemove(update tgbotapi.Update) string {
	if err := dbIsRegisteredChat(update.Message.Chat.ID); err != nil {
		return err.Error()
	}
	args := strings.Split(update.Message.CommandArguments(), " ")
	if len(args) < 2 {
		return "Need args: ACCESSTOKEN CHATID"
	}
	if args[0] != token {
		return "Invalid access token!"
	}
	if chatid, err := strconv.ParseInt(args[1], 10, 64); err != nil {
		return "Invalid chat id"
	} else {
		return dbLeave(chatid)
	}
}

func notifyTelegram(msg string) string {
	chats := dbGetChats()

	for _, chat := range chats {
		msg := tgbotapi.NewMessage(chat, msg)
		bot.Send(msg)

	}

	return "OK"
}

func cmdSendPhoto(update tgbotapi.Update) string {
	if err := dbIsRegisteredChat(update.Message.Chat.ID); err != nil {
		return err.Error()
	}

	matches, err := filepath.Glob("/data/output/Camera1/*/*.jpg")
	if err != nil {
		return "Error: " + err.Error()
	}

	imgFile := ArgImage
	if len(matches) > 0 {
		sort.Strings(matches)
		imgFile = matches[len(matches)-1]
	}

	pho := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, imgFile)
	pho.ReplyToMessageID = update.Message.MessageID
	pho.Caption = imgFile
	if _, err := bot.Send(pho); err != nil {
		return "Error: " + err.Error()
	}

	return "Photo send"
}

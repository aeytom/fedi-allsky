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

	motionGetCameras()
	for c, n := range cameras {
		log.Printf("cam %s := '%s'", c, n)
	}

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
			Command:     "register",
			Description: "/register token – register current chat for notification updates"},
		{
			Command:     "photo",
			Description: "/photo - get last image"},
		{
			Command:     "snapshot",
			Description: "/snapshot - do a new snapshot"},
		{
			Command:     "abo",
			Description: "/abo – automatically send new cropped images"},
		{
			Command:     "leave",
			Description: "/leave – do not receive further notifications"},
		{
			Command:     "admin",
			Description: "/admin chatid – toggle admin state"},
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
				msg.Text = `User Commands:
/leave - leave current chat
/photo - get last photo
/register accesstoken - register new chat with accesstoken
`
			case "register":
				msg.Text = cmdRegister(update)
			}

			user := dbFindUser(update.Message.Chat.ID)
			if user != nil {
				switch update.Message.Command() {
				case "leave":
					msg.Text = cmdLeave(user)
				case "photo":
					msg.Text = cmdSendPhoto(update, user)
				case "snapshot":
					msg.Text = cmdSendSnapshot(update, user)
				case "abo":
					msg.Text = cmdToggleAbo(update, user)
				case "camera":
					msg.Text = cmdSetCamera(update, user)
				case "list":
					// msg.ParseMode = tgbotapi.ModeMarkdownV2
					msg.Text = cmdAdminListUsers(update, user)
				case "remove":
					msg.Text = cmdAdminRemove(update, user)
				case "admin":
					msg.Text = cmdAdminToggle(update, user)
				case "token":
					msg.Text = cmdAdminToken(update, user)
				}
			}

			if msg.Text != "" {
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			}
		}
	}
}

func genToken() string {

	token = generatePassword(16, 0, 2, 2)
	log.Printf("current access token: `%s`", token)

	return token
}

func cmdAdminListUsers(update tgbotapi.Update, user *Chat) string {
	if !user.Admin {
		return "Kein Admin - keine Aktion!"
	}
	return dbListUsers()
}

func cmdAdminToken(update tgbotapi.Update, user *Chat) string {
	if !user.Admin {
		return "Kein Admin - keine Aktion!"
	}
	genToken()
	return token
}

func cmdAdminToggle(update tgbotapi.Update, user *Chat) string {
	if !user.Admin {
		return "Kein Admin - keine Aktion!"
	}

	if update.Message.CommandArguments() == "" {
		return `Admin Commands:
/list – list users
/admin [ID] – toggle user admin state
/remove ID – remove user
/token – show token for /register`
	} else if chatid, err := strconv.ParseInt(update.Message.CommandArguments(), 10, 64); err != nil {
		return "Invalid chat id"
	} else if err = dbToggleAdmin(chatid); err != nil {
		return err.Error()
	} else {
		return cmdAdminListUsers(update, user)
	}
}

func cmdAdminRemove(update tgbotapi.Update, user *Chat) string {
	if !user.Admin {
		return "Kein Admin - keine Aktion!"
	}

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
		ruser := dbFindUser(chatid)
		return dbLeave(ruser)
	}
}

func cmdLeave(user *Chat) string {
	return dbLeave(user)
}

func notifyTelegram(msg string) string {
	chats := dbGetChats()

	for _, chat := range chats {
		msg := tgbotapi.NewMessage(chat, msg)
		bot.Send(msg)

	}

	return "OK"
}

func telegramSendCroppedImage(imgFile string) string {

	for _, user := range chats {
		if !user.AboPhoto {
			continue
		}
		pho := tgbotapi.NewPhotoUpload(user.ChatID, imgFile)
		pho.Caption = imgFile
		if _, err := bot.Send(pho); err != nil {
			return "Error: " + err.Error()
		}
	}

	return "OK"
}

func cmdSendSnapshot(update tgbotapi.Update, user *Chat) string {
	if _, err := motionAction(user.Camera, "snapshot"); err != nil {
		return err.Error()
	}
	return cmdSendPhoto(update, user)
}

func cmdSendPhoto(update tgbotapi.Update, user *Chat) string {

	target_dir, err := motionConfigGet(user.Camera, "target_dir")
	if err != nil {
		return err.Error()
	}

	glob := target_dir + "/**/*.jpg"
	matches, err := filepath.Glob(glob)
	if err != nil {
		return "Error: " + err.Error()
	}

	imgFile := ArgImage
	if len(matches) > 0 {
		sort.Strings(matches)
		imgFile = matches[len(matches)-1]
	}

	pho := tgbotapi.NewPhotoUpload(user.ChatID, imgFile)
	pho.ReplyToMessageID = update.Message.MessageID
	pho.Caption = imgFile
	if _, err := bot.Send(pho); err != nil {
		return "Error: " + err.Error()
	}

	return "Photo send"
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

func cmdSetCamera(update tgbotapi.Update, user *Chat) string {
	if err := motionGetCameras(); err != nil {
		return err.Error()
	}

	arg := update.Message.CommandArguments()
	if arg == "" {
		clist := "Cameras:\n"
		for c, n := range cameras {
			clist += fmt.Sprintf("/camera %s := %s\n", c, n)
		}
		return clist
	} else {
		for c, n := range cameras {
			if c == arg {
				user.Camera = c
				dbFlush()
				return "Use camera: " + n
			}
		}
	}
	return fmt.Sprintf("Can't use camera: '%s'", arg)
}

func cmdToggleAbo(update tgbotapi.Update, user *Chat) string {
	user.AboPhoto = !user.AboPhoto
	dbFlush()
	if user.AboPhoto {
		return "You get new cropped images automatically."
	} else {
		return "You don't get new images automatically."
	}
}

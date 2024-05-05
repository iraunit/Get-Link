package whatsapp

import (
	"fmt"
	chatbot "github.com/green-api/whatsapp-chatbot-golang"
)

type Whatsapp interface {
	Init()
}

type WhatsappImpl struct {
	bot *chatbot.Bot
}

func NewWhatsappImpl() *WhatsappImpl {
	return &WhatsappImpl{}
}

func (impl *WhatsappImpl) init() {
	token := "EAAJwsyyYK4IBOx2zs9spnWawWbsPQBqYFHXe4J3e2SYowr1KPOIZC7inPIT4e98mL3rjB5nZBbL3pMEGrvrE4f12B1Y5VgJZCD6RgFf0cqWJ03YLccvZC62ZCESeLatlXyEgN9M4RRxQTDFeYK6lA29BTTvGN6Mlk2rbTRWZBGWRpw034uZB8DHruFI51nwoQtCpI7UZBC3U2OXoMvDt"
	phone := "190677170785803"
	bot := chatbot.NewBot(phone, token)
	bot.IncomingMessageHandler(func(message *chatbot.Notification) {
		fmt.Println("Notification: ", message)
	})
}

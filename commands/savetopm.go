package commands

import (
	"fmt"

	"github.com/NexonSU/telegram-go-chatbot/utils"
	tele "gopkg.in/telebot.v3"
)

//Resend post on user request
func SaveToPM(context tele.Context) error {
	if context.Message() == nil || context.Message().ReplyTo == nil {
		return context.Reply("Пример использования:\n/savetopm в ответ на какое-либо сообщение")
	}
	_, err := utils.Bot.Forward(context.Sender(), context.Message().ReplyTo)
	if err != nil {
		return context.Reply(fmt.Sprintf("Не удалось отправить сообщение в ЛС:\n<code>%v</code>", err.Error()))
	}
	return context.Reply("Отправил сообщение тебе в личку, проверяй.")
}

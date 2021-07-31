package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/NexonSU/telegram-go-chatbot/app/utils"
	"gopkg.in/tucnak/telebot.v3"
	"gorm.io/gorm/clause"
)

//Kill user on /kill
func Kill(context telebot.Context) error {
	var text = strings.Split(context.Text(), " ")
	if (context.Message().ReplyTo == nil && len(text) != 2) || (context.Message().ReplyTo != nil && len(text) != 1) {
		return context.Reply("Пример использования: <code>/kill {ID или никнейм}</code>\nИли отправь в ответ на какое-либо сообщение <code>/kill</code>")
	}
	target, _, err := utils.FindUserInMessage(context)
	if err != nil {
		return context.Reply(fmt.Sprintf("Не удалось определить пользователя:\n<code>%v</code>", err.Error()))
	}
	ChatMember, err := utils.Bot.ChatMemberOf(context.Chat(), &target)
	if err != nil {
		return context.Reply(fmt.Sprintf("Ошибка определения пользователя чата:\n<code>%v</code>", err.Error()))
	}
	var duelist utils.Duelist
	result := utils.DB.Model(utils.Duelist{}).Where(target.ID).First(&duelist)
	if result.RowsAffected == 0 {
		duelist.UserID = target.ID
		duelist.Kills = 0
		duelist.Deaths = 0
	}
	duelist.Deaths++
	result = utils.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(duelist)
	if result.Error != nil {
		return err
	}
	ChatMember.RestrictedUntil = time.Now().Add(time.Second * time.Duration(600*duelist.Deaths)).Unix()
	err = utils.Bot.Restrict(context.Chat(), ChatMember)
	if err != nil {
		return err
	}
	return context.Send(fmt.Sprintf("💥 %v пристрелил %v.\n%v отправился на респавн на %v0 минут.", utils.UserFullName(context.Sender()), utils.UserFullName(&target), utils.UserFullName(&target), duelist.Deaths))
}

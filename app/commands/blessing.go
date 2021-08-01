package commands

import (
	"fmt"
	"time"

	"github.com/NexonSU/telegram-go-chatbot/app/utils"
	"gopkg.in/tucnak/telebot.v3"
	"gorm.io/gorm/clause"
)

//Kill user on /blessing, /suicide
func Blessing(context telebot.Context) error {
	ChatMember, err := utils.Bot.ChatMemberOf(context.Chat(), context.Sender())
	if err != nil {
		return err
	}
	if ChatMember.Role == "administrator" || ChatMember.Role == "creator" {
		return context.Reply(fmt.Sprintf("<code>👻 %v возродился у костра.</code>", utils.UserFullName(context.Sender())))
	}
	var duelist utils.Duelist
	result := utils.DB.Model(utils.Duelist{}).Where(context.Sender().ID).First(&duelist)
	if result.RowsAffected == 0 {
		duelist.UserID = context.Sender().ID
		duelist.Kills = 0
		duelist.Deaths = 0
	}
	duelist.Deaths++
	result = utils.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(duelist)
	if result.Error != nil {
		return result.Error
	}
	ChatMember.RestrictedUntil = time.Now().Add(time.Second * time.Duration(600*duelist.Deaths)).Unix()
	err = utils.Bot.Restrict(context.Chat(), ChatMember)
	if err != nil {
		return err
	}
	err = context.Send(fmt.Sprintf("<code>💥 %v выбрал лёгкий путь.\nРеспавн через %v0 минут.</code>", utils.UserFullName(context.Sender()), duelist.Deaths))
	if err != nil {
		return err
	}
	return utils.Bot.Delete(context.Message())
}

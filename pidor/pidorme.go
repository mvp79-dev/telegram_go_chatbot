package pidor

import (
	"time"

	"github.com/NexonSU/telegram-go-chatbot/utils"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	tele "gopkg.in/telebot.v3"
)

//Send DB stats on /pidorme
func Pidorme(context tele.Context) error {
	// prt will replace fmt package to format text according plurals defined in utils package
	// If no plural rule matched it will be ignored and processed as usual formatting
	prt := message.NewPrinter(language.Russian)

	var pidor utils.PidorStats
	var countYear int64
	var countAlltime int64
	pidor.UserID = context.Sender().ID
	utils.DB.Model(&utils.PidorStats{}).Where(pidor).Where("date BETWEEN ? AND ?", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Local), time.Now()).Count(&countYear)
	utils.DB.Model(&utils.PidorStats{}).Where(pidor).Count(&countAlltime)
	thisYear := prt.Sprintf("В этом году ты был пидором дня — %d раз", countYear)
	total := prt.Sprintf("За всё время ты был пидором дня — %d раз!", countAlltime)
	return context.Reply(prt.Sprintf("%s\n%s", thisYear, total))
}

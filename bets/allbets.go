package bets

import (
	"fmt"
	"html"
	"strconv"
	"time"

	"github.com/NexonSU/telegram-go-chatbot/utils"
	tele "gopkg.in/telebot.v3"
)

// List all bets
func AllBets(context tele.Context) error {
	var betlist string
	var bet utils.Bets
	var user tele.User
	var i = 0
	var from int64
	var to int64
	if len(context.Args()) > 0 {
		if context.Args()[0] == "all" {
			from = 0
		}
	}
	from = time.Now().Local().Truncate(24 * time.Hour).Unix()
	to = time.Now().Local().Add(43800 * time.Hour).Unix()
	result, _ := utils.DB.Model(&utils.Bets{}).Where("timestamp > ? AND timestamp < ?", from, to).Order("timestamp ASC").Rows()
	for result.Next() {
		err := utils.DB.ScanRows(result, &bet)
		if err != nil {
			return err
		}
		i++
		user, err = utils.GetUserFromDB(strconv.FormatInt(bet.UserID, 10))
		if err != nil {
			return err
		}
		betlist += fmt.Sprintf("%v, %v:\n<pre>%v</pre>\n", time.Unix(bet.Timestamp, 0).Format("02.01.2006"), utils.UserFullName(&user), html.EscapeString(bet.Text))
		if len(betlist) > 3900 {
			return context.Reply(betlist)
		}
	}
	return context.Reply(betlist)
}

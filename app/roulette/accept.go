package roulette

import (
	"fmt"
	"github.com/NexonSU/telegram-go-chatbot/app/utils"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm/clause"
	"time"
)

func RouletteAccept(busy map[string]bool) func(*tb.Callback) {
	return func(c *tb.Callback) {
		err := utils.Bot.Respond(c, &tb.CallbackResponse{})
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		message := c.Message
		victim := c.Message.Entities[0].User
		if victim.ID != c.Sender.ID {
			err := utils.Bot.Respond(c, &tb.CallbackResponse{})
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			return
		}
		player := c.Message.Entities[1].User
		busy["russianroulette"] = false
		busy["russianroulettePending"] = false
		busy["russianrouletteInProgress"] = true
		defer func() { busy["russianrouletteInProgress"] = false }()
		success := []string{"%v остаётся в живых. Хм... может порох отсырел?", "В воздухе повисла тишина. %v остаётся в живых.", "%v сегодня заново родился.", "%v остаётся в живых. Хм... я ведь зарядил его?", "%v остаётся в живых. Прикольно, а давай проверим на ком-нибудь другом?"}
		invincible := []string{"пуля отскочила от головы %v и улетела в другой чат.", "%v похмурил брови и отклеил расплющенную пулю со своей головы.", "но ничего не произошло. %v взглянул на револьвер, он был неисправен.", "пуля прошла навылет, но не оставила каких-либо следов на %v."}
		fail := []string{"мозги %v разлетелись по чату!", "%v упал со стула и его кровь растеклась по месседжу.", "%v замер и спустя секунду упал на стол.", "пуля едва не задела кого-то из участников чата! А? Что? А, %v мёртв, да.", "и в воздухе повисла тишина. Все начали оглядываться, когда %v уже был мёртв."}
		prefix := fmt.Sprintf("Дуэль! %v против %v!\n", utils.MentionUser(player), utils.MentionUser(victim))
		_, err = utils.Bot.Edit(message, fmt.Sprintf("%vЗаряжаю один патрон в револьвер и прокручиваю барабан.", prefix), &tb.SendOptions{ReplyMarkup: nil})
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		time.Sleep(time.Second * 2)
		_, err = utils.Bot.Edit(message, fmt.Sprintf("%vКладу револьвер на стол и раскручиваю его.", prefix))
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		time.Sleep(time.Second * 2)
		if utils.RandInt(1, 360)%2 == 0 {
			player, victim = victim, player
		}
		_, err = utils.Bot.Edit(message, fmt.Sprintf("%vРевольвер останавливается на %v, первый ход за ним.", prefix, utils.MentionUser(victim)))
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		bullet := utils.RandInt(1, 6)
		for i := 1; i <= bullet; i++ {
			time.Sleep(time.Second * 2)
			prefix = fmt.Sprintf("Дуэль! %v против %v, раунд %v:\n%v берёт револьвер, приставляет его к голове и...\n", utils.MentionUser(player), utils.MentionUser(victim), i, utils.MentionUser(victim))
			_, err := utils.Bot.Edit(message, prefix)
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			if bullet != i {
				time.Sleep(time.Second * 2)
				_, err := utils.Bot.Edit(message, fmt.Sprintf("%v🍾 %v", prefix, fmt.Sprintf(success[utils.RandInt(0, len(success)-1)], utils.MentionUser(victim))))
				if err != nil {
					utils.ErrorReporting(err, c.Message)
					return
				}
				player, victim = victim, player
			}
		}
		time.Sleep(time.Second * 2)
		PlayerChatMember, err := utils.Bot.ChatMemberOf(c.Message.Chat, player)
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		VictimChatMember, err := utils.Bot.ChatMemberOf(c.Message.Chat, victim)
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		if (PlayerChatMember.Role == "creator" || PlayerChatMember.Role == "administrator") && (VictimChatMember.Role == "creator" || VictimChatMember.Role == "administrator") {
			_, err = utils.Bot.Edit(message, fmt.Sprintf("%vПуля отскакивает от головы %v и летит в голову %v.", prefix, utils.MentionUser(victim), utils.MentionUser(player)))
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			time.Sleep(time.Second * 2)
			_, err = utils.Bot.Edit(message, fmt.Sprintf("%vПуля отскакивает от головы %v и летит в голову %v.", prefix, utils.MentionUser(player), utils.MentionUser(victim)))
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			time.Sleep(time.Second * 2)
			_, err = utils.Bot.Edit(message, fmt.Sprintf("%vПуля отскакивает от головы %v и летит в мою голову... блять.", prefix, utils.MentionUser(victim)))
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			busy["bot_is_dead"] = true
			return
		}
		if utils.StringInSlice(victim.Username, utils.Config.Telegram.Admins) {
			_, err = utils.Bot.Edit(message, fmt.Sprintf("%v😈 Наводит револьвер на %v и стреляет.", prefix, utils.MentionUser(player)))
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			time.Sleep(time.Second * 3)
			var duelist utils.Duelist
			result := utils.DB.Model(utils.Duelist{}).Where(player.ID).First(&duelist)
			if result.RowsAffected == 0 {
				duelist.UserID = player.ID
				duelist.Kills = 0
				duelist.Deaths = 0
			}
			duelist.Deaths++
			result = utils.DB.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(duelist)
			if result.Error != nil {
				utils.ErrorReporting(result.Error, c.Message)
				return
			}
			PlayerChatMember.RestrictedUntil = time.Now().Add(time.Second * time.Duration(600*duelist.Deaths)).Unix()
			err = utils.Bot.Restrict(c.Message.Chat, PlayerChatMember)
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			_, err = utils.Bot.Edit(message, fmt.Sprintf("%v😈 Наводит револьвер на %v и стреляет.\nЯ хз как это объяснить, но %v победитель!\n%v отправился на респавн на %v0 минут.", prefix, utils.MentionUser(player), utils.MentionUser(victim), utils.MentionUser(player), duelist.Deaths))
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			return
		}
		if VictimChatMember.Role == "creator" || VictimChatMember.Role == "administrator" {
			prefix = fmt.Sprintf("%v💥 %v", prefix, fmt.Sprintf(invincible[utils.RandInt(0, len(invincible)-1)], utils.MentionUser(victim)))
			_, err := utils.Bot.Edit(message, prefix)
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			time.Sleep(time.Second * 2)
			_, err = utils.Bot.Edit(message, fmt.Sprintf("%v\nПохоже, у нас ничья.", prefix))
			if err != nil {
				utils.ErrorReporting(err, c.Message)
				return
			}
			return
		}
		prefix = fmt.Sprintf("%v💥 %v", prefix, fmt.Sprintf(fail[utils.RandInt(0, len(fail)-1)], utils.MentionUser(victim)))
		_, err = utils.Bot.Edit(message, prefix)
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		time.Sleep(time.Second * 2)
		var VictimDuelist utils.Duelist
		result := utils.DB.Model(utils.Duelist{}).Where(victim.ID).First(&VictimDuelist)
		if result.RowsAffected == 0 {
			VictimDuelist.UserID = victim.ID
			VictimDuelist.Kills = 0
			VictimDuelist.Deaths = 0
		}
		VictimDuelist.Deaths++
		result = utils.DB.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(VictimDuelist)
		if result.Error != nil {
			utils.ErrorReporting(result.Error, c.Message)
			return
		}
		VictimChatMember.RestrictedUntil = time.Now().Add(time.Second * time.Duration(600*VictimDuelist.Deaths)).Unix()
		err = utils.Bot.Restrict(c.Message.Chat, VictimChatMember)
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		_, err = utils.Bot.Edit(message, fmt.Sprintf("%v\nПобедитель дуэли: %v.\n%v отправился на респавн на %v0 минут.", prefix, utils.MentionUser(player), utils.MentionUser(victim), VictimDuelist.Deaths))
		if err != nil {
			utils.ErrorReporting(err, c.Message)
			return
		}
		var PlayerDuelist utils.Duelist
		result = utils.DB.Model(utils.Duelist{}).Where(victim.ID).First(&PlayerDuelist)
		if result.RowsAffected == 0 {
			PlayerDuelist.UserID = victim.ID
			PlayerDuelist.Kills = 0
			PlayerDuelist.Deaths = 0
		}
		PlayerDuelist.Kills++
		result = utils.DB.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(PlayerDuelist)
		if result.Error != nil {
			utils.ErrorReporting(result.Error, c.Message)
			return
		}
	}
}

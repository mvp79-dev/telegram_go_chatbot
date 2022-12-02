package duel

import (
	"time"

	"github.com/NexonSU/telegram-go-chatbot/utils"

	"golang.org/x/text/language"
	plurals "golang.org/x/text/message"
	tele "gopkg.in/telebot.v3"
	"gorm.io/gorm/clause"
)

func Accept(context tele.Context) error {
	// prt will replace fmt package to format text according plurals defined in utils package
	// If no plural rule matched it will be ignored and processed as usual formatting
	prt := plurals.NewPrinter(language.Russian)

	message := context.Message()
	victim := message.Entities[0].User
	if victim.ID != context.Sender().ID {
		return context.Respond(&tele.CallbackResponse{Text: utils.GetNope()})
	}
	err := utils.Bot.Respond(context.Callback(), &tele.CallbackResponse{})
	if err != nil {
		return err
	}
	player := message.Entities[1].User
	busy["russianroulette"] = false
	busy["russianroulettePending"] = false
	busy["russianrouletteInProgress"] = true
	defer func() { busy["russianrouletteInProgress"] = false }()
	success := []string{"%v остаётся в живых. Хм... может порох отсырел?", "В воздухе повисла тишина. %v остаётся в живых.", "%v сегодня заново родился.", "%v остаётся в живых. Хм... я ведь зарядил его?", "%v остаётся в живых. Прикольно, а давай проверим на ком-нибудь другом?"}
	invincible := []string{"пуля отскочила от головы %v и улетела в другой чат.", "%v похмурил брови и отклеил расплющенную пулю со своей головы.", "но ничего не произошло. %v взглянул на револьвер, он был неисправен.", "пуля прошла навылет, но не оставила каких-либо следов на %v."}
	fail := []string{"мозги %v разлетелись по чату!", "%v упал со стула и его кровь растеклась по месседжу.", "%v замер и спустя секунду упал на стол.", "пуля едва не задела кого-то из участников чата! А? Что? А, %v мёртв, да.", "и в воздухе повисла тишина. Все начали оглядываться, когда %v уже был мёртв."}
	prefix := prt.Sprintf("Дуэль! %v против %v!\n", utils.MentionUser(player), utils.MentionUser(victim))
	_, err = utils.Bot.Edit(message, prt.Sprintf("%vЗаряжаю один патрон в револьвер и прокручиваю барабан.", prefix), &tele.SendOptions{ReplyMarkup: nil})
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 2)
	_, err = utils.Bot.Edit(message, prt.Sprintf("%vКладу револьвер на стол и раскручиваю его.", prefix))
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 2)
	if utils.RandInt(1, 360)%2 == 0 {
		player, victim = victim, player
	}
	_, err = utils.Bot.Edit(message, prt.Sprintf("%vРевольвер останавливается на %v, первый ход за ним.", prefix, utils.MentionUser(victim)))
	if err != nil {
		return err
	}
	bullet := utils.RandInt(1, 5)
	for i := 1; i <= bullet; i++ {
		time.Sleep(time.Second * 2)
		prefix = prt.Sprintf("Дуэль! %v против %v, раунд %v:\n%v берёт револьвер, приставляет его к голове и...\n", utils.MentionUser(player), utils.MentionUser(victim), i, utils.MentionUser(victim))
		_, err := utils.Bot.Edit(message, prefix)
		if err != nil {
			return err
		}
		if bullet != i {
			time.Sleep(time.Second * 2)
			_, err := utils.Bot.Edit(message, prt.Sprintf("%v🍾 %v", prefix, prt.Sprintf(success[utils.RandInt(0, len(success)-1)], utils.MentionUser(victim))))
			if err != nil {
				return err
			}
			player, victim = victim, player
		}
	}
	time.Sleep(time.Second * 2)
	PlayerChatMember, err := utils.Bot.ChatMemberOf(context.Message().Chat, player)
	if err != nil {
		return err
	}
	VictimChatMember, err := utils.Bot.ChatMemberOf(context.Message().Chat, victim)
	if err != nil {
		return err
	}
	if (PlayerChatMember.Role == "creator" || PlayerChatMember.Role == "administrator") && (VictimChatMember.Role == "creator" || VictimChatMember.Role == "administrator") {
		_, err = utils.Bot.Edit(message, prt.Sprintf("%vПуля отскакивает от головы %v и летит в голову %v.", prefix, utils.MentionUser(victim), utils.MentionUser(player)))
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 2)
		_, err = utils.Bot.Edit(message, prt.Sprintf("%vПуля отскакивает от головы %v и летит в голову %v.", prefix, utils.MentionUser(player), utils.MentionUser(victim)))
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 2)
		var ricochetVictim *tele.ChatMember
		var lastMessage utils.Message
		for i := 1; i < 100; i++ {
			result := utils.DB.Where(utils.Message{ChatID: context.Chat().ID}).Order("id desc").Offset(i).Last(&lastMessage)
			if result.Error != nil {
				continue
			}
			ricochetVictim, err = utils.Bot.ChatMemberOf(context.Chat(), &tele.User{ID: lastMessage.UserID})
			if err != nil {
				continue
			}
			if ricochetVictim.Role == "member" {
				VictimChatMember = ricochetVictim
				victim = ricochetVictim.User
				break
			}
		}
	}
	if utils.IsAdmin(victim.ID) {
		_, err = utils.Bot.Edit(message, prt.Sprintf("%v😈 Наводит револьвер на %v и стреляет.", prefix, utils.MentionUser(player)))
		if err != nil {
			return err
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
		}).Create(&duelist)
		if result.Error != nil {
			return err
		}
		PlayerChatMember.RestrictedUntil = time.Now().Add(time.Second * time.Duration(60*duelist.Deaths)).Unix()
		err = utils.Bot.Restrict(context.Message().Chat, PlayerChatMember)
		if err != nil {
			return err
		}
		_, err = utils.Bot.Edit(message, prt.Sprintf("%v😈 Наводит револьвер на %v и стреляет.\nЯ хз как это объяснить, но %v победитель!\n%v отправился на респавн на %d мин.", prefix, utils.MentionUser(player), utils.MentionUser(victim), utils.MentionUser(player), duelist.Deaths))
		if err != nil {
			return err
		}
		return err
	}
	if VictimChatMember.Role == "creator" || VictimChatMember.Role == "administrator" {
		prefix = prt.Sprintf("%v💥 %v", prefix, prt.Sprintf(invincible[utils.RandInt(0, len(invincible)-1)], utils.MentionUser(victim)))
		_, err := utils.Bot.Edit(message, prefix)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 2)
		_, err = utils.Bot.Edit(message, prt.Sprintf("%v\nПохоже, у нас ничья.", prefix))
		if err != nil {
			return err
		}
		return err
	}
	prefix = prt.Sprintf("%v💥 %v", prefix, prt.Sprintf(fail[utils.RandInt(0, len(fail)-1)], utils.MentionUser(victim)))
	_, err = utils.Bot.Edit(message, prefix)
	if err != nil {
		return err
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
	}).Create(&VictimDuelist)
	if result.Error != nil {
		return err
	}
	VictimChatMember.RestrictedUntil = time.Now().Add(time.Second * time.Duration(60*VictimDuelist.Deaths)).Unix()
	err = utils.Bot.Restrict(context.Message().Chat, VictimChatMember)
	if err != nil {
		return err
	}
	_, err = utils.Bot.Edit(message, prt.Sprintf("%v\nПобедитель дуэли: %v.\n%v отправился на респавн на %d мин.", prefix, utils.MentionUser(player), utils.MentionUser(victim), VictimDuelist.Deaths))
	if err != nil {
		return err
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
	}).Create(&PlayerDuelist)
	if result.Error != nil {
		return result.Error
	}
	return err
}

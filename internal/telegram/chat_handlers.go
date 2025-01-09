package telegram

import (
	"fmt"

	"telega_chess/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCreateChatInstruction(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, roomID string) {
	// Здесь мы не создаём группу автоматически (Telegram API не даёт).
	// Просто даём инструкцию.
	instructionText := `
Чтобы создать новый групповой чат:
1) Выйдите в главное меню Telegram → «Новая группа»
	*пожалуйсиа, постарайтесь создать простую группу(где будите только Вы)
2) Добавьте меня (@TelegaChessBot) в группу
3) Перейдите в настройки группы и назначьте меня администратором (минимум с правами «Change group info» и «Invite users»)
	3.1 Для полноценного взаимодействия с чат-комнатой - рекомендую закончить простые настройки группы
4) Готово! Я автоматически переименую группу и подготовлю приглошение для второго игрока.
`
	// Подставим имя бота
	formattedText := fmt.Sprintf(instructionText, bot.Self.UserName)

	bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, formattedText))

	hint := tgbotapi.NewMessage(
		query.Message.Chat.ID,
		fmt.Sprintf("потребуется для настройки комнаты```\n/setroom %d\n```", roomID))
	hint.ParseMode = tgbotapi.ModeMarkdownV2
	bot.Send(hint)
}

func handleManageRoomMenu(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	// Показываем 2-3 кнопки:
	// 1) "Продолжить настройку"
	// 2) "Отмена" (или "Назад")

	continueBtn := tgbotapi.NewInlineKeyboardButtonData("Продолжить настройку", "continue_setup")
	cancelBtn := tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel_setup")

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(continueBtn),
		tgbotapi.NewInlineKeyboardRow(cancelBtn),
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Выберите действие:")
	msg.ReplyMarkup = kb
	bot.Send(msg)
}

func handleContinueSetup(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Проверим, есть ли уже room, привязанная к этому chatID
	room, err := db.GetRoomByChatID(chatID)
	if err != nil {
		// Нет привязки
		text := `
Пока к этой группе не привязана никакая комната.
Введите команду /setroom <room_id> для привязки:
Например: /setroom 546e81dc-5aff-463a-9681-3e41627b8df2
`
		bot.Send(tgbotapi.NewMessage(chatID, text))
		return
	}

	// Если есть, проверим, есть ли второй игрок
	if room.Player2 == nil {
		// Предлагаем сгенерировать invite-link
		linkCfg := tgbotapi.ChatInviteLinkConfig{
			ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
		}
		link, err := bot.GetInviteLink(linkCfg)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при создании invite-link: "+err.Error()))
			return
		}

		text := fmt.Sprintf("Комната уже привязана к room_id=%s, но пока нет второго игрока.\n"+
			"Пригласите его ссылкой:\n%s", room.RoomID, link)
		bot.Send(tgbotapi.NewMessage(chatID, text))
	} else {
		// Есть 2 игрока => "Игра началась!" (или уже идёт)
		newTitle := makeFinalTitle(room)
		tryRenameGroup(bot, chatID, newTitle)

		notifyGameStarted(bot, room)
	}
}

func handleRetryRename(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, newTitle string) {
	// Просто заново вызываем tryRenameGroup
	// chatID = query.Message.Chat.ID
	tryRenameGroup(bot, query.Message.Chat.ID, newTitle)
}

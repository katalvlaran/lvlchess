package telegram

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleNewChatMembers(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chat := update.Message.Chat // Информация о чате (Type, ID, Title, ...)
	newMembers := update.Message.NewChatMembers

	for _, member := range newMembers {
		if member.IsBot && member.ID == bot.Self.ID {
			// 1. Бот добавлен в новый чат => попробуем переименовать
			newTitle := fmt.Sprintf("tChess:%d", time.Now().Unix())

			renameConfig := tgbotapi.SetChatTitleConfig{
				ChatID: chat.ID,
				Title:  newTitle,
			}
			if _, err := bot.Request(renameConfig); err != nil {
				log.Printf("Не удалось переименовать группу: %v", err)
			}

			// Предлагаем управление
			manageButton := tgbotapi.NewInlineKeyboardButtonData(
				"Управление комнатой",
				"manage_room",
			)
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(manageButton),
			)
			msg := tgbotapi.NewMessage(chat.ID,
				"Привет! Нажмите 'Управление комнатой' для дальнейших действий.")
			msg.ReplyMarkup = kb
			bot.Send(msg)

			// 2. Создадим приглашение
			linkCfg := tgbotapi.ChatInviteLinkConfig{
				ChatConfig: tgbotapi.ChatConfig{ChatID: chat.ID},
				// Можно указать Name: "Приглашение для tChess",
				// ExpireDate, MemberLimit, CreatesJoinRequest и т.д.
			}
			chatLink, err := bot.GetInviteLink(linkCfg)
			if err != nil {
				log.Printf("Ошибка генерации InviteLink: %v", err)
				bot.Send(tgbotapi.NewMessage(chat.ID,
					"Не смог создать ссылку-приглашение. Проверьте, есть ли у меня права."))
				return
			}

			text := fmt.Sprintf(
				"Группа успешно создана! Вот ссылка для второго игрока:\n%s",
				chatLink,
			)
			bot.Send(tgbotapi.NewMessage(chat.ID, text))

			// 3. (Опционально) связать chat.ID с нашей roomID
			//    Например, если user1 предварительно ввёл /setroom 123
			//    и мы сохранили это в контексте.
			//    Или если user1ID = X => находим комнату, где Player1ID=X
			//    room.ChatID = &chat.ID
			//    db.UpdateRoom(room)
		}
	}
}

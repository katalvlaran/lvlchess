package telegram

import (
	"fmt"
	"log"
	"time"

	"telega_chess/internal/db"
	"telega_chess/internal/game"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleNewChatMembers вызывается, когда в группе появляются новые участники (NewChatMembers)
func HandleNewChatMembers(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chat := update.Message.Chat
	newMembers := update.Message.NewChatMembers

	// Получим room, если он есть:
	room, err := db.GetRoomByChatID(string(chat.ID)) // Нужно написать метод в db, типа GetRoomByChatID
	var haveRoom bool
	if err == nil && room.RoomID != "" {
		haveRoom = true
	}

	for _, member := range newMembers {
		if member.IsBot && member.ID == bot.Self.ID {
			// Бот добавлен в новую группу → пытаемся переименовать, если нет прав, выдаём "Повторить..."
			tryRenameGroup(bot, chat.ID, fmt.Sprintf("tChess:%d", time.Now().Unix()))

			// Покажем кнопку "Управление комнатой"
			manageButton := tgbotapi.NewInlineKeyboardButtonData("Управление комнатой", "manage_room")
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(manageButton),
			)
			msg := tgbotapi.NewMessage(chat.ID,
				"Привет! Я бот Telega-Chess. Чтобы продолжить настройку комнаты, нажмите [Управление комнатой].")
			msg.ReplyMarkup = kb
			bot.Send(msg)

		} else {
			// Возможно, это второй игрок (или ещё кто-то).
			// Если у нас уже есть "привязанная" комната (haveRoom == true),
			// и room.Player2ID == nil => назначаем его вторым игроком
			if haveRoom && room.Player2ID == nil {
				user2ID := int64(member.ID)
				// Сохраним username (или fallback)
				user2Name := member.UserName
				if user2Name == "" {
					user2Name = member.FirstName
				}
				room.Player2ID = &user2ID
				room.Player2Username = &user2Name
				game.AssignRandomColors(&room) // назначили белые/чёрные, если ещё не назначены

				room.Status = "playing"
				if err := db.UpdateRoom(room); err != nil {
					bot.Send(tgbotapi.NewMessage(chat.ID, "Ошибка обновления комнаты: "+err.Error()))
					return
				}

				msgText := game.MakeGameStartedMessage(room)
				bot.Send(tgbotapi.NewMessage(chat.ID, msgText))

				// Переименуем в "tChess:@user1_⚔️_@user2"
				newTitle := makeFinalTitle(room)
				tryRenameGroup(bot, chat.ID, newTitle)

				// Сообщим: "Игра началась!"
				text := game.MakeGameStartedMessage(room)
				bot.Send(tgbotapi.NewMessage(chat.ID, text))

				// ASCII-доска
				fen := ""
				if room.BoardState != nil {
					fen = *room.BoardState
				}
				asciiBoard, _ := game.RenderBoardCustom(fen) // или RenderBoardFromFEN
				bot.Send(tgbotapi.NewMessage(chat.ID, asciiBoard))
			}
		}
	}
}

// tryRenameGroup обёртка, которая пытается переименовать группу.
// Если не хватает прав - выводит кнопку "Повторить переименование".
func tryRenameGroup(bot *tgbotapi.BotAPI, chatID int64, newTitle string) {
	renameConfig := tgbotapi.SetChatTitleConfig{
		ChatID: chatID,
		Title:  newTitle,
	}
	_, err := bot.Request(renameConfig)
	if err != nil {
		log.Printf("Не удалось переименовать группу (chatID=%d): %v", chatID, err)

		// Сообщим пользователю, что нужны права
		retryBtn := tgbotapi.NewInlineKeyboardButtonData(
			"Повторить переименование",
			fmt.Sprintf("retry_rename:%s", newTitle),
		)
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(retryBtn),
		)
		msg := tgbotapi.NewMessage(chatID,
			"У меня нет прав на изменение названия группы. Дайте права 'Change group info' и нажмите [Повторить переименование].")
		msg.ReplyMarkup = kb
		bot.Send(msg)
	}
}

func makeFinalTitle(r db.Room) string {
	if r.Player1Username == nil {
		return "tChess:????"
	}
	if r.Player2Username == nil {
		return fmt.Sprintf("tChess:@%s_⚔️_??", *r.Player1Username)
	}
	return fmt.Sprintf("tChess:@%s_⚔️_@%s", *r.Player1Username, *r.Player2Username)
}

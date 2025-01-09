package telegram

import (
	"fmt"
	"strings"
	"time"

	"telega_chess/internal/db"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

/*	main_handlers.go
	HandleUpdate
	handleMessage
	handleCallback
	handleNewChatMembers */

// HandleUpdate - универсальная точка входа
func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(bot, update)
	case update.CallbackQuery != nil:
		handleCallback(bot, update.CallbackQuery)
	case update.MyChatMember != nil:
		// ???
		// мне нравится и я хочу использовать, но как...?
		// какие примеры как можно круто обработать подобные события !?
		// ???
	case update.ChatMember != nil:
		handleNewChatMembers(bot, update)
	}
}

// handleMessage - обрабатываем сообщения/команды
func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := update.Message

	if msg.NewChatMembers != nil {
		handleNewChatMembers(bot, update)
	}

	// Проверка: если в группе и msg.IsCommand():
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if msg.IsCommand() {
			if msg.Command() == "setroom" {
				handleSetRoomCommand(bot, update)
			} else {
				// Отключаем остальные команды
				reply := tgbotapi.NewMessage(msg.Chat.ID, "Здесь команды не работают. Используйте /setroom <room_id> или кнопки.")
				bot.Send(reply)
			}
		} else {
			// Любой текст -> "Используйте кнопки..."
			//reply := tgbotapi.NewMessage(msg.Chat.ID, "Используйте кнопки (inline).")
			reply := tgbotapi.NewMessage(msg.Chat.ID, "🌚")
			bot.Send(reply)
		}
		return
	}

	// Если это личка
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			handleStartCommand(bot, update) // в logic.go
		case "create_room":
			handleCreateRoomCommand(bot, update)
		case "play_with_bot":
			handlePlayWithBotCommand(bot, update)
		case "game_list":
			handleGameListCommand(bot, update)
		default:
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Команда не распознана. Используйте кнопки или /start."))
		}
	} else {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "🌚"))
	}
}

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	data := query.Data

	switch {
	case data == "manage_room":
		handleManageRoomMenu(bot, query)
	case data == "continue_setup":
		handleContinueSetup(bot, query)
	case strings.HasPrefix(data, "retry_rename:"):
		newTitle := data[len("retry_rename:"):]
		handleRetryRename(bot, query, newTitle)
	case strings.HasPrefix(data, "create_chat_"):
		// пользователь нажал "Создать и перейти в Чат"
		roomID := data[len("create_chat_"):]
		handleCreateChatInstruction(bot, query, roomID)

	case strings.HasPrefix(data, "delete_"):
		roomID := data[7:]
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Комната "+roomID+" будет удалена (заглушка).")
		bot.Send(msg)
	default:
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Неизвестный callback: "+data)
		bot.Send(msg)
	}

	// Подтверждаем callback
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := bot.Request(callback); err != nil {
		utils.Logger.Error("😖 AnswerCallbackQuery error 👾", zap.Error(err))
	}
}

// HandleNewChatMembers вызывается, когда в группе появляются новые участники (NewChatMembers)
func handleNewChatMembers(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chat := update.Message.Chat
	newMembers := update.Message.NewChatMembers

	// Получим room, если он есть:
	room, err := db.GetRoomByChatID(chat.ID) // Нужно написать метод в db, типа GetRoomByChatID
	var haveRoom bool
	if err == nil && room.RoomID != "" {
		haveRoom = true
	}

	for _, member := range newMembers {
		if member.IsBot && member.ID == bot.Self.ID {
			// Бот добавлен в новую группу → пытаемся переименовать, если нет прав, выдаём "Повторить..."
			//tryRenameGroup(bot, chat.ID, fmt.Sprintf("tChess:%d", room.Player1.Username))
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
			if haveRoom && room.Player2 == nil {
				p2 := &db.User{
					ID:        member.ID,
					Username:  member.UserName,
					FirstName: member.FirstName,
					ChatID:    db.UnregisteredPrivateChat,
				}

				if err = db.CreateOrUpdateUser(p2); err != nil {
					bot.Send(tgbotapi.NewMessage(chat.ID, "Ошибка создания второго игрока: "+err.Error()))
					return
				}

				room.Player2 = p2
				game.AssignRandomColors(room) // назначили белые/чёрные, если ещё не назначены

				room.Status = "playing"
				if err := db.UpdateRoom(room); err != nil {
					bot.Send(tgbotapi.NewMessage(chat.ID, "Ошибка обновления комнаты: "+err.Error()))
					return
				}

				// Переименуем в "tChess:@user1_⚔️_@user2"
				newTitle := makeFinalTitle(room)
				tryRenameGroup(bot, chat.ID, newTitle)

				notifyGameStarted(bot, room)
				break
			}
		}
	}
}

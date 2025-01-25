package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"telega_chess/internal/db"
	"telega_chess/internal/db/models"
	"telega_chess/internal/db/repositories"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	CommandDelimiter   = ":"
	ActionMove         = "move"
	ActionChooseFigure = "choose_figure"
	CreateRoom         = "create_room"
	PlayWithBot        = "play_with_bot"
	SetupRoom          = "setup_room"
	RetryRename        = "retry_rename"
	SetupRoomWhite     = "setup_room_white"
	ContinueSetup      = "continue_setup"
	ManageRoom         = "manage_room"
	RoomID             = "roomID"
	JoinThisRoom       = "join_this_room"
	CreateChat         = "create_chat_"
	GameList           = "game_list"
	RoomEntrance       = "room_entrance"
	Delete             = "delete_"
)

var TelegramHandler *Handler

type Handler struct {
	Bot      *tgbotapi.BotAPI
	UserRepo *repositories.UsersRepository
	RoomRepo *repositories.RoomsRepository
}

func NewHandler(bot *tgbotapi.BotAPI) {
	TelegramHandler = &Handler{
		Bot:      bot,
		RoomRepo: db.GetRoomsRepo(),
		UserRepo: db.GetUsersRepo(),
	}
}

// HandleUpdate - универсальная точка входа
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		h.handleMessage(ctx, update)
	case update.CallbackQuery != nil:
		h.handleCallback(ctx, update.CallbackQuery)
	case update.MyChatMember != nil:
		// ???
		// мне нравится и я хочу использовать, но как...?
		// какие примеры как можно круто обработать подобные события !?
		// ???
	case update.ChatMember != nil:
		h.handleNewChatMembers(ctx, update)
	}
}

// handleMessage - обрабатываем сообщения/команды
func (h *Handler) handleMessage(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message

	if msg.NewChatMembers != nil {
		h.handleNewChatMembers(ctx, update)
	}

	// Проверка: если в группе и msg.IsCommand():
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if msg.IsCommand() {
			if msg.Command() == "setroom" {
				h.handleSetRoomCommand(ctx, update)
			} else {
				// Отключаем остальные команды
				reply := tgbotapi.NewMessage(msg.Chat.ID, "Здесь команды не работают. Используйте /setroom <room_id> или кнопки.")
				h.Bot.Send(reply)
			}
		} else {
			// Любой текст -> "Используйте кнопки..."
			//reply := tgbotapi.NewMessage(msg.Chat.ID, "Используйте кнопки (inline).")
			reply := tgbotapi.NewMessage(msg.Chat.ID, "🌚")
			h.Bot.Send(reply)
		}
		return
	}

	// Если это личка
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			h.handleStartCommand(ctx, update)
		default:
			h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Команда не распознана. Используйте кнопки или /start."))
		}
	} else {
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "🌚"))
	}
}

func (h *Handler) handleCallback(ctx context.Context, query *tgbotapi.CallbackQuery) {
	data := query.Data

	switch {
	case data == "create_room":
		h.handleCreateRoomCommand(ctx, query)
	case data == "play_with_bot":
		h.handlePlayWithBotCommand(ctx, query)
	case data == "game_list":
		h.handleGameListCommand(ctx, query)
	case data == "setup_room":
		h.handleAskWhoIsWhite(ctx, query)
	case strings.HasPrefix(data, "setup_room_white:"):
		choice := strings.TrimPrefix(data, "setup_room_white:")
		h.handleSetupRoomWhiteChoice(ctx, query, choice)
	case strings.HasPrefix(data, "choose_figure:"):
		h.handleChooseFigureCallback(ctx, query)
	case strings.HasPrefix(data, "move:"):
		h.handleMoveCallback(ctx, query)
	case data == "manage_room":
		h.handleManageRoomMenu(ctx, query)
	case data == "continue_setup":
		h.handleContinueSetup(ctx, query)
	case strings.HasPrefix(data, "roomID:"):
		roomID := data[len("roomID:"):]
		h.handleChooseRoom(ctx, query, roomID)
	case strings.HasPrefix(data, "join_this_room:"):
		rid := data[len("join_this_room:"):]
		h.handleJoinThisRoom(ctx, query, rid)
	case strings.HasPrefix(data, "retry_rename:"):
		newTitle := data[len("retry_rename:"):]
		h.handleRetryRename(ctx, query, newTitle)
	case strings.HasPrefix(data, "create_chat_"):
		// пользователь нажал "Создать и перейти в Чат"
		roomID := data[len("create_chat_"):]
		h.handleCreateChatInstruction(ctx, query, roomID)
	case strings.HasPrefix(data, "delete_"):
		roomID := data[7:]
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Комната "+roomID+" будет удалена (заглушка).")
		h.Bot.Send(msg)
	case strings.HasPrefix(data, "room_entrance:"):
		roomID := data[len("room_entrance:"):]
		h.handleRoomEntrance(ctx, query, roomID)
	default:
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Неизвестный callback: "+data)
		h.Bot.Send(msg)
	}

	// Подтверждаем callback
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := h.Bot.Request(callback); err != nil {
		utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
	}
}

// HandleNewChatMembers вызывается, когда в группе появляются новые участники (NewChatMembers)
func (h *Handler) handleNewChatMembers(ctx context.Context, update tgbotapi.Update) {
	chat := update.Message.Chat
	newMembers := update.Message.NewChatMembers

	// Получим room, если он есть:
	room, err := h.RoomRepo.GetRoomByChatID(ctx, chat.ID) // Нужно написать метод в db, типа GetRoomByChatID
	var haveRoom bool
	if err == nil && room.RoomID != "" {
		haveRoom = true
	}

	for _, member := range newMembers {
		if member.IsBot && member.ID == h.Bot.Self.ID {
			// Бот добавлен в новую группу → пытаемся переименовать, если нет прав, выдаём "Повторить..."
			//tryRenameGroup(ctx, chat.ID, fmt.Sprintf("tChess:%d", room.Player1.Username))
			h.tryRenameGroup(h.Bot, chat.ID, fmt.Sprintf("tChess:%d", time.Now().Unix()))

			// Покажем кнопку "Управление комнатой"
			manageButton := tgbotapi.NewInlineKeyboardButtonData("Управление комнатой", "manage_room")
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(manageButton),
			)
			msg := tgbotapi.NewMessage(chat.ID,
				"Привет! Я бот Telega-Chess. Чтобы продолжить настройку комнаты, нажмите [Управление комнатой].")
			msg.ReplyMarkup = kb
			h.Bot.Send(msg)

		} else {
			// Возможно, это второй игрок (или ещё кто-то).
			// Если у нас уже есть "привязанная" комната (haveRoom == true),
			// и room.Player2ID == nil => назначаем его вторым игроком
			if haveRoom && room.Player2ID == nil {
				p2 := &models.User{
					ID:        member.ID,
					Username:  member.UserName,
					FirstName: member.FirstName,
					ChatID:    models.UnregisteredPrivateChat,
				}

				if err = h.UserRepo.CreateOrUpdateUser(ctx, p2); err != nil {
					h.Bot.Send(tgbotapi.NewMessage(chat.ID, "Ошибка создания второго игрока: "+err.Error()))
					return
				}

				room.Player2ID = &p2.ID
				game.AssignRandomColors(room) // назначили белые/чёрные, если ещё не назначены

				room.Status = models.RoomStatusPlaying
				if err := h.RoomRepo.UpdateRoom(ctx, room); err != nil {
					h.Bot.Send(tgbotapi.NewMessage(chat.ID, "Ошибка обновления комнаты: "+err.Error()))
					return
				}

				// Переименуем в "tChess:@user1_⚔️_@user2"
				room.RoomTitle = h.MakeFinalTitle(ctx, room)
				h.tryRenameGroup(h.Bot, chat.ID, room.RoomTitle)

				h.notifyGameStarted(ctx, room)
				break
			}
		}
	}
}

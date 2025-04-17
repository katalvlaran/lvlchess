package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"lvlchess/config"
	"lvlchess/internal/db"
	"lvlchess/internal/db/models"
	"lvlchess/internal/db/repositories"
	"lvlchess/internal/game"
	"lvlchess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// CallbackAction constants to help us parse or handle user interactions in inline keyboards etc.
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

// TelegramHandler is a global-like reference, but ideally you'd keep it in your main
// and pass references. For now, we do a static var.
var TelegramHandler *Handler

// Handler is the core structure that has references to the bot, plus repos. This is used
// by all callback and command handlers. You register it in NewHandler().
type Handler struct {
	Bot                   *tgbotapi.BotAPI
	UserRepo              *repositories.UsersRepository
	RoomRepo              *repositories.RoomsRepository
	TournamentRepo        *repositories.TournamentRepository
	TournamentSettingRepo *repositories.TournamentSettingsRepository
}

// NewHandler initializes the global TelegramHandler with references
// to the repositories (taken from db.GetRoomsRepo() etc.).
func NewHandler(bot *tgbotapi.BotAPI) {
	TelegramHandler = &Handler{
		Bot:      bot,
		RoomRepo: db.GetRoomsRepo(),
		UserRepo: db.GetUsersRepo(),
		// If you want to handle tournaments here:
		TournamentRepo:        db.GetTournamentsRepo(),
		TournamentSettingRepo: db.GetTournamentSettingsRepo(),
	}
}

// HandleUpdate is the primary entrypoint for every incoming message/update from Telegram.
// We route them based on whether it's a message, callbackQuery, or chatMember event, etc.
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		// An incoming message or command in text form.
		h.handleMessage(ctx, update)

	case update.CallbackQuery != nil:
		// An inline keyboard callback was pressed.
		h.handleCallback(ctx, update.CallbackQuery)

	case update.MyChatMember != nil:
		// Called when the bot's status changes in a chat (e.g., promoted to admin).
		// Currently we do nothing. Potential expansions: detect if we lost admin privileges, etc.
		// h.handleBotChatMember(ctx, update)
		_ = update.MyChatMember

	case update.ChatMember != nil:
		// Called when a chat participant's status changes. If new players joined, we might see it here too.
		h.handleNewChatMembers(ctx, update)
	}
}

// handleMessage processes text-based messages in private or group chats.
// If it's a command, we dispatch to the relevant command (like /start).
// If it’s a non-command text in a group, we might ignore or respond with “Use the inline buttons.”
func (h *Handler) handleMessage(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message

	// If newChatMembers is set, we might have a new user or bot added to this chat.
	if msg.NewChatMembers != nil {
		h.handleNewChatMembers(ctx, update)
	}

	// If it's a group or supergroup:
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if msg.IsCommand() {
			if msg.Command() == "setroom" {
				h.handleSetRoomCommand(ctx, update)
			} else {
				// We can ignore all other commands in group context or warn user.
				reply := tgbotapi.NewMessage(msg.Chat.ID,
					"Commands in group chat are restricted. Use /setroom <room_id> or inline buttons.")
				h.Bot.Send(reply)
			}
		} else {
			// Non-command text in a group → optional minimal response.
			reply := tgbotapi.NewMessage(msg.Chat.ID, "🌚")
			h.Bot.Send(reply)
		}
		return
	}

	// If it's a private chat:
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			h.handleStartCommand(ctx, update)
		default:
			// If we get other commands we haven't recognized, just respond briefly.
			h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Unrecognized command. Use /start or inline buttons."))
		}
	} else {
		// A plain text message in private chat. Some implementations do a fallback or "Use /start".
		h.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "🌚"))
	}
}

// handleCallback processes all inline keyboard callbacks. data is typically something like
// "create_room", "move:e2-e4&roomID:123...", etc. We parse and dispatch logic accordingly.
func (h *Handler) handleCallback(ctx context.Context, query *tgbotapi.CallbackQuery) {
	data := query.Data
	utils.Logger.Info("handleCallback:", zap.Any("query", query))

	switch {
	case data == config.Cfg.GameShortName:
		// Формируем URL (можно добавить UTM или initData в query, но initData мы возьмём из WebView)
		gameURL := config.Cfg.GameURL
		// Отвечаем Telegram, чтобы открыть WebView
		//callback := tgbotapi.NewCallbackWithURL(query.ID, gameURL)
		callback := tgbotapi.NewCallback(query.ID, gameURL)
		if _, err := h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error:", zap.Error(err))
		}
	case data == "tournament_list":
		h.handleTournamentList(ctx, query)

	case data == "create_tournament":
		h.handleCreateTournament(ctx, query)

	case data == "join_tournament:<ID>":
		// Not fully implemented. We could parse the ID then call handleJoinTournament.
		// ...

	case data == "start_tournament:<ID>":
		// Similarly handle starting a tournament with the given ID.

	case data == CreateRoom:
		h.handleCreateRoomCommand(ctx, query)

	case data == PlayWithBot:
		h.handlePlayWithBotCommand(ctx, query)

	case data == GameList:
		h.handleGameListCommand(ctx, query)

	case data == SetupRoom:
		h.handleAskWhoIsWhite(ctx, query)

	case strings.HasPrefix(data, fmt.Sprintf("%s%s", SetupRoomWhite, CommandDelimiter)):
		choice := strings.TrimPrefix(data, fmt.Sprintf("%s%s", SetupRoomWhite, CommandDelimiter))
		h.handleSetupRoomWhiteChoice(ctx, query, choice)

	case strings.HasPrefix(data, fmt.Sprintf("%s%s", ActionChooseFigure, CommandDelimiter)):
		h.handleChooseFigureCallback(ctx, query)

	case strings.HasPrefix(data, fmt.Sprintf("%s%s", ActionMove, CommandDelimiter)):
		h.handleMoveCallback(ctx, query)

	case data == ManageRoom:
		h.handleManageRoomMenu(ctx, query)

	case data == ContinueSetup:
		h.handleContinueSetup(ctx, query)

	case strings.HasPrefix(data, fmt.Sprintf("%s%s", RoomID, CommandDelimiter)):
		roomID := data[len(fmt.Sprintf("%s%s", RoomID, CommandDelimiter)):]
		h.handleChooseRoom(ctx, query, roomID)

	case strings.HasPrefix(data, fmt.Sprintf("%s%s", JoinThisRoom, CommandDelimiter)):
		rid := data[len(fmt.Sprintf("%s%s", JoinThisRoom, CommandDelimiter)):]
		h.handleJoinThisRoom(ctx, query, rid)

	case strings.HasPrefix(data, fmt.Sprintf("%s%s", RetryRename, CommandDelimiter)):
		newTitle := data[len(fmt.Sprintf("%s%s", RetryRename, CommandDelimiter)):]
		h.handleRetryRename(ctx, query, newTitle)

	case strings.HasPrefix(data, CreateChat):
		// "create_chat_XXX" => show instructions to create group, add bot, etc.
		roomID := data[len(CreateChat):]
		h.handleCreateChatInstruction(ctx, query, roomID)

	case strings.HasPrefix(data, Delete):
		roomID := data[len(Delete):]
		msg := tgbotapi.NewMessage(query.Message.Chat.ID,
			"Room "+roomID+" will be deleted (placeholder).")
		h.Bot.Send(msg)

	case strings.HasPrefix(data, fmt.Sprintf("%s%s", RoomEntrance, CommandDelimiter)):
		roomID := data[len(fmt.Sprintf("%s%s", RoomEntrance, CommandDelimiter)):]
		h.handleRoomEntrance(ctx, query, roomID)

	default:
		// Unknown callback, just log or respond:
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Unknown callback: "+data)
		h.Bot.Send(msg)
	}

	// We always send an empty callback to confirm we received their action, removing the spinner in Telegram UI.
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := h.Bot.Request(callback); err != nil {
		utils.Logger.Error("Error in AnswerCallbackQuery: "+err.Error(), zap.Error(err))
	}
}

// handleNewChatMembers is triggered if the user or bot was added to a group chat, or if new users joined.
// We sometimes handle the scenario: if the bot is newly added, rename the group, or if the second player joined, do X, etc.
func (h *Handler) handleNewChatMembers(ctx context.Context, update tgbotapi.Update) {
	chat := update.Message.Chat
	newMembers := update.Message.NewChatMembers

	// Attempt to find if the chat is linked to a "room."
	room, err := h.RoomRepo.GetRoomByChatID(ctx, chat.ID)
	var haveRoom bool
	if err == nil && room.RoomID != "" {
		haveRoom = true
	}

	for _, member := range newMembers {
		if member.IsBot && member.ID == h.Bot.Self.ID {
			// The bot was just added to this group. Attempt to rename the group or show "manage room" button.
			h.tryRenameGroup(h.Bot, chat.ID, fmt.Sprintf("tChess:%d", time.Now().Unix()))

			manageButton := tgbotapi.NewInlineKeyboardButtonData("Manage Room", ManageRoom)
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(manageButton),
			)
			msg := tgbotapi.NewMessage(chat.ID,
				"Hello! I'm the lvlChess bot. Use [Manage Room] to continue room setup.")
			msg.ReplyMarkup = kb
			h.Bot.Send(msg)

		} else {
			// Possibly the second player joined. If we have a room & no player2, we can assign them.
			if haveRoom && room.Player2ID == nil {
				p2 := &models.User{
					ID:        member.ID,
					Username:  member.UserName,
					FirstName: member.FirstName,
					ChatID:    models.UnregisteredPrivateChat, // not a private chat, so 0 or custom
				}
				if err = h.UserRepo.CreateOrUpdateUser(ctx, p2); err != nil {
					h.Bot.Send(tgbotapi.NewMessage(chat.ID,
						"Error creating second player: "+err.Error()))
					return
				}
				room.Player2ID = &p2.ID
				game.AssignRandomColors(room)
				room.Status = models.RoomStatusPlaying

				if err := h.RoomRepo.UpdateRoom(ctx, room); err != nil {
					h.Bot.Send(tgbotapi.NewMessage(chat.ID,
						"Error updating room: "+err.Error()))
					return
				}

				// Attempt group rename based on player names, e.g. "tChess:@p1_⚔️_@p2"
				room.RoomTitle = h.MakeFinalTitle(ctx, room)
				h.tryRenameGroup(h.Bot, chat.ID, room.RoomTitle)

				h.notifyGameStarted(ctx, room)
				break
			}
		}
	}
}

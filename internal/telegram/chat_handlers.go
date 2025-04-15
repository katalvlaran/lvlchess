package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCreateChatInstruction is called when user clicks "Создать и перейти в Чат" inline button.
// We can't auto-create a Telegram group, so we show instructions to manually create or add the bot.
func (h *Handler) handleCreateChatInstruction(ctx context.Context, query *tgbotapi.CallbackQuery, roomID string) {
	instructionText := `
Чтобы создать новый групповой чат:
1) Выйдите в главное меню Telegram → «Новая группа»
   (Попробуйте создать простую группу, где вы единственный участник сначала)
2) Добавьте меня (@TelegaChessBot) в группу
3) Назначьте меня администратором ("Change group info", "Invite users")
4) Готово! Я переименую группу и приглашу второго игрока.
`
	// Insert the actual bot's username if needed
	formattedText := fmt.Sprintf(instructionText, h.Bot.Self.UserName)
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, formattedText))

	// Show them maybe how to use /setroom <room_id>
	hint := tgbotapi.NewMessage(
		query.Message.Chat.ID,
		fmt.Sprintf("Для привязки комнаты используйте ```\n/setroom %s\n```", roomID),
	)
	hint.ParseMode = tgbotapi.ModeMarkdownV2
	h.Bot.Send(hint)
}

// handleManageRoomMenu is triggered when user presses some "Manage Room" button in the group chat.
// We might show multiple next-step choices, e.g. "Continue setup," "Cancel," etc.
func (h *Handler) handleManageRoomMenu(ctx context.Context, query *tgbotapi.CallbackQuery) {
	continueBtn := tgbotapi.NewInlineKeyboardButtonData("Продолжить настройку", ContinueSetup)
	cancelBtn := tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel_setup")

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(continueBtn),
		tgbotapi.NewInlineKeyboardRow(cancelBtn),
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Выберите действие:")
	msg.ReplyMarkup = kb
	h.Bot.Send(msg)
}

// handleContinueSetup tries to see if the group is already linked to some room (via chatID).
// If so, we check whether a second player is present, and if not, generate an invite link.
func (h *Handler) handleContinueSetup(ctx context.Context, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	room, err := h.RoomRepo.GetRoomByChatID(ctx, chatID)
	if err != nil {
		// No linked room => instruct user to /setroom <room_id>.
		text := `
Пока к этой группе не привязана никакая комната.
Введите команду /setroom <room_id> для привязки:
Пример: /setroom 546e81dc-5aff-463a-9681-3e41627b8df2
`
		h.Bot.Send(tgbotapi.NewMessage(chatID, text))
		return
	}

	// If we do have a room, but no second player, we ask them to invite someone
	if room.Player2ID == nil {
		linkCfg := tgbotapi.ChatInviteLinkConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: chatID}}
		link, err := h.Bot.GetInviteLink(linkCfg)
		if err != nil {
			h.Bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при создании invite-link: "+err.Error()))
			return
		}
		text := fmt.Sprintf("Комната уже привязана к room_id=%s, но пока нет второго игрока.\nПриглашение:\n%s", room.RoomID, link)
		h.Bot.Send(tgbotapi.NewMessage(chatID, text))
	} else {
		// If second player is present, let's finalize the room (rename group, start game).
		room.RoomTitle = h.MakeFinalTitle(ctx, room)
		h.tryRenameGroup(h.Bot, chatID, room.RoomTitle)
		h.RoomRepo.UpdateRoom(ctx, room)

		h.notifyGameStarted(ctx, room)
	}
}

// handleRetryRename is an optional method that tries to rename the group again if we lacked permissions
// the first time. The user can press "retry" after giving the bot admin privileges.
func (h *Handler) handleRetryRename(ctx context.Context, query *tgbotapi.CallbackQuery, newTitle string) {
	h.tryRenameGroup(h.Bot, query.Message.Chat.ID, newTitle)
}

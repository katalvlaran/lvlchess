package telegram

import (
	"context"
	"fmt"

	"lvlchess/internal/db"
	"lvlchess/internal/db/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleTournamentList is a placeholder to display "all tournaments for the user" or a subset.
// The actual logic to retrieve such data is not fully implemented here.
func (h *Handler) handleTournamentList(ctx context.Context, query *tgbotapi.CallbackQuery) {
	// userID := query.From.ID
	// Possibly call db.GetTournamentsRepo().GetTournamentsByUser(userID), etc.
	text := "Список ваших турниров: (заглушка)\n(тут вывести названия, статусы, кнопки 'Присоединиться', 'Старт')"
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	h.Bot.Send(msg)
}

// handleCreateTournament is triggered if user clicks a "create tournament" button.
// We create a basic record with a Title, an initial Player array, and set status=planned.
func (h *Handler) handleCreateTournament(ctx context.Context, query *tgbotapi.CallbackQuery) {
	t := &models.Tournament{
		Title:   "Мой тестовый турнир",
		Prise:   "Приз для победителя: ...",
		Players: []int64{query.From.ID}, // the initiator
		Status:  models.TournamentStatusPlanned,
	}
	if err := db.GetTournamentsRepo().CreateTournament(ctx, t); err != nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"Ошибка создания турнира: "+err.Error()))
		return
	}

	text := fmt.Sprintf("Турнир создан! ID=%s, Название=%s", t.ID, t.Title)
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, text))
}

// handleJoinTournament is a stub for when user chooses to join an existing tournament.
// We add them to the "Players" array in that tournament record.
func (h *Handler) handleJoinTournament(ctx context.Context, query *tgbotapi.CallbackQuery, tournamentID string) {
	userID := query.From.ID
	err := db.GetTournamentsRepo().AddPlayer(ctx, tournamentID, userID)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Ошибка: "+err.Error()))
		return
	}
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Вы успешно присоединились к турниру!"))
}

// handleStartTournament sets tournament status to active=1 and sets start_at=NOW() in DB.
// Then we might create the first round of rooms, etc. (not fully shown here).
func (h *Handler) handleStartTournament(ctx context.Context, query *tgbotapi.CallbackQuery, tournamentID string) {
	err := db.GetTournamentsRepo().StartTournament(ctx, tournamentID)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Ошибка старта турнира: "+err.Error()))
		return
	}
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Турнир запущен!"))
}

// Additional methods might include handleTournamentBrackets, handleTournamentMatch, handleTournamentRounds, etc.

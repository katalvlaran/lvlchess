package telegram

import (
	"context"
	"fmt"

	"telega_chess/internal/db"
	"telega_chess/internal/db/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleTournamentList(ctx context.Context, query *tgbotapi.CallbackQuery) {
	_ = query.From.ID
	// идея: получить все турниры, где userID входит в t.Players
	// Для упрощения: пока нет готового метода "GetTournamentsByUser",
	// можно просто вывести заглушку или сделать SELECT * FROM tournament, а потом фильтровать
	// Либо вручную:
	// tournaments, _ := db.GetTournamentsRepo().GetAllTournaments(ctx) // <- метод нужно создать
	// ...

	text := "Список ваших турниров: (заглушка)\n(тут вывести названия, статусы, кнопки 'Присоединиться', 'Старт')"
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	h.Bot.Send(msg)
}

func (h *Handler) handleCreateTournament(ctx context.Context, query *tgbotapi.CallbackQuery) {
	// Пример: создаём турнир
	t := &models.Tournament{
		Title:   "Мой тестовый турнир",
		Prise:   "Приз для победителя: ...",
		Players: []int64{query.From.ID}, // создатель
		Status:  models.TournamentStatusPlanned,
	}
	if err := db.GetTournamentsRepo().CreateTournament(ctx, t); err != nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"Ошибка создания турнира: "+err.Error()))
		return
	}

	// Ответ
	text := fmt.Sprintf("Турнир создан! ID=%s, Название=%s", t.ID, t.Title)
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, text))
}

func (h *Handler) handleJoinTournament(ctx context.Context, query *tgbotapi.CallbackQuery, tournamentID string) {
	userID := query.From.ID
	// добавляем user в players
	err := db.GetTournamentsRepo().AddPlayer(ctx, tournamentID, userID)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Ошибка: "+err.Error()))
		return
	}
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Вы успешно присоединились к турниру!"))
}

func (h *Handler) handleStartTournament(ctx context.Context, query *tgbotapi.CallbackQuery, tournamentID string) {
	err := db.GetTournamentsRepo().StartTournament(ctx, tournamentID)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Ошибка старта турнира: "+err.Error()))
		return
	}
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Турнир запущен!"))
}

// и т.д.

package game

import (
	"fmt"
	"math/rand"
	"time"

	"telega_chess/internal/db"
)

// AssignRandomColors задаёт WhiteID и BlackID, если они ещё не были назначены.
// Если already assigned (оба не nil), функция ничего не меняет.
func AssignRandomColors(r *db.Room) {
	// Если уже есть белые/чёрные, ничего не делаем.
	if r.WhiteID != nil && r.BlackID != nil {
		return
	}

	// Убедимся, что у нас есть player1 и player2
	if r.Player2 == nil {
		// нет второго игрока — не будем назначать
		return
	}

	// Инициализируем seed (можно один раз в main(), но если не инициализировано, сделаем локально)
	rand.Seed(time.Now().UnixNano())

	if rand.Intn(2) == 0 {
		// player1 -> white, player2 -> black
		r.WhiteID = &r.Player1.ID
		r.BlackID = &r.Player2.ID
	} else {
		// player2 -> white, player1 -> black
		r.WhiteID = &r.Player2.ID
		black := &r.Player1.ID
		r.BlackID = black
	}
}

func MakeGameStartedMessage(r *db.Room) string {
	// Возьмём удобные "displayName" для player1 и player2
	p1Name := getDisplayName(r.Player1.Username, "Player1")
	var p2Name string
	if r.Player2.Username == "" {
		p2Name = "Player2"
	} else {
		p2Name = getDisplayName(r.Player2.Username, "Player2")
	}

	// Кто белые, кто чёрные?
	var whiteName, blackName string

	if r.WhiteID != nil && *r.WhiteID == r.Player1.ID {
		whiteName = p1Name
		blackName = p2Name
	} else if r.WhiteID != nil && r.Player2 != nil && *r.WhiteID == r.Player2.ID {
		whiteName = p2Name
		blackName = p1Name
	} else {
		// На случай, если не назначены
		whiteName = p1Name + "?"
		blackName = p2Name + "?"
	}

	// Соберём текст
	return fmt.Sprintf(
		"Игра началась!\n%s (белые ♙) vs %s (чёрные ♟)\n\nНачальная позиция:",
		whiteName, blackName,
	)
}

func getDisplayName(usernamePtr string, fallback string) string {
	if usernamePtr == "" || usernamePtr == "" {
		return fallback
	}
	username := usernamePtr
	// Проверим, есть ли уже "@" в начале
	if username[0] != '@' {
		username = "@" + username
	}
	return username
}

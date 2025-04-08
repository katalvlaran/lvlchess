package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

const (
	TournamentStatusPlanned  = 0
	TournamentStatusActive   = 1
	TournamentStatusFinished = 2

	TSStatusWaiting = 0
	TSStatusOngoing = 1
	TSStatusDone    = 2
)

// Tournament модель для таблицы "tournament"
type Tournament struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	Prise     string    `db:"prise"`
	Players   []int64   `db:"players"` // либо []int64, которые в PG храним как JSONB
	Status    int       `db:"status"`  // 0=planned,1=active,2=finished
	StartAt   time.Time `db:"start_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TournamentSettings — связка турнир <-> комната (раунд).
type TournamentSettings struct {
	TID    string `db:"t_id"`
	RID    string `db:"r_id"`
	Rank   int    `db:"rank"`   // например, номер раунда
	Status int    `db:"status"` // 0=waiting,1=ongoing,2=done
}

func (u *Tournament) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Title, validation.Required),
		validation.Field(&u.Prise, validation.Required),
		//validation.Field(&u.Players, validation.NilOrNotEmpty),
		validation.Field(&u.Status, validation.Required, validation.In(TournamentStatusPlanned, TournamentStatusActive, TournamentStatusFinished)),
		validation.Field(&u.StartAt, validation.Required),
	)
}

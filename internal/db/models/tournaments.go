package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

const (
	TournamentStatusPlanned  = 0 // Not started yet
	TournamentStatusActive   = 1 // Currently in progress
	TournamentStatusFinished = 2 // Concluded or finished

	TSStatusWaiting = 0 // For "tournament_settings", waiting
	TSStatusOngoing = 1 // Ongoing round
	TSStatusDone    = 2 // Round is finished
)

// Tournament is the main structure for managing a multi-player or multi-round event.
// The array 'Players' is stored as JSON in the DB, each int64 representing a user ID.
type Tournament struct {
	ID        string    `db:"id"`       // Unique ID (UUID)
	Title     string    `db:"title"`    // Name of the tournament
	Prise     string    `db:"prise"`    // Some string describing the prize or reward
	Players   []int64   `db:"players"`  // Array of user IDs as participants
	Status    int       `db:"status"`   // 0=planned,1=active,2=finished
	StartAt   time.Time `db:"start_at"` // The time the tournament is started
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TournamentSettings holds relationships between a specific Tournament and a Room, typically used per round.
type TournamentSettings struct {
	TID    string `db:"t_id"`   // The tournament ID
	RID    string `db:"r_id"`   // The room ID (each round uses a separate room)
	Rank   int    `db:"rank"`   // e.g., the round number or bracket position
	Status int    `db:"status"` // 0=waiting,1=ongoing,2=done
}

// Validate ensures required fields are present and valid values are used.
func (u *Tournament) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Title, validation.Required),
		validation.Field(&u.Prise, validation.Required),
		validation.Field(&u.Status,
			validation.Required,
			validation.In(TournamentStatusPlanned, TournamentStatusActive, TournamentStatusFinished)),
		validation.Field(&u.StartAt, validation.Required),
	)
}

func (u *TournamentSettings) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.TID, validation.Required),
		validation.Field(&u.RID, validation.Required),
		validation.Field(&u.Status,
			validation.Required,
			validation.In(TSStatusWaiting, TSStatusOngoing, TSStatusDone)),
	)
}

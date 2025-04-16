package repositories

import (
	"context"
	"fmt"

	"lvlchess/internal/db/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
TournamentSettingsRepository manages the link between Tournaments and Rooms.
For instance, you can add references to which room corresponds to which round.
*/
type TournamentSettingsRepository struct {
	pool *pgxpool.Pool
}

// NewTournamentSettingsRepository creates a new instance of TournamentSettingsRepository.
func NewTournamentSettingsRepository(pool *pgxpool.Pool) *TournamentSettingsRepository {
	return &TournamentSettingsRepository{pool: pool}
}

/*
LinkRoomToTournament establishes a record in "tournament_settings",
indicating that the given room (rid) is part of the tournament (tid) at a certain rank or round.
*/
func (r *TournamentSettingsRepository) LinkRoomToTournament(
	ctx context.Context,
	tid, rid string,
	rank int,
) error {
	sql := `
INSERT INTO tournament_settings (t_id, r_id, rank, status)
VALUES ($1, $2, $3, 0)  -- status=0 => waiting
`
	_, err := r.pool.Exec(ctx, sql, tid, rid, rank)
	if err != nil {
		return fmt.Errorf("LinkRoomToTournament: %w", err)
	}
	return nil
}

/*
UpdateTournamentRoomRank allows you to update the 'rank' and 'status'
fields for a particular (t_id, r_id) combination, e.g., to mark
a room as 'done' or move from round 1 to round 2, etc.
*/
func (r *TournamentSettingsRepository) UpdateTournamentRoomRank(
	ctx context.Context,
	tid, rid string,
	newRank int,
	newStatus int,
) error {
	sql := `
UPDATE tournament_settings
SET
  rank   = $1,
  status = $2
WHERE t_id = $3
  AND r_id = $4
`
	_, err := r.pool.Exec(ctx, sql, newRank, newStatus, tid, rid)
	if err != nil {
		return fmt.Errorf("UpdateTournamentRoomRank: %w", err)
	}
	return nil
}

/*
GetRoomsByTournament returns the list of all tournament_settings records
for a given tournament ID. Each record includes t_id, r_id, rank, status.
*/
func (r *TournamentSettingsRepository) GetRoomsByTournament(
	ctx context.Context,
	tid string,
) ([]models.TournamentSettings, error) {
	const sql = `
SELECT t_id, r_id, rank, status
FROM tournament_settings
WHERE t_id = $1
ORDER BY rank ASC
`
	rows, err := r.pool.Query(ctx, sql, tid)
	if err != nil {
		return nil, fmt.Errorf("GetRoomsByTournament: %w", err)
	}
	defer rows.Close()

	var results []models.TournamentSettings
	for rows.Next() {
		var ts models.TournamentSettings
		err := rows.Scan(
			&ts.TID,
			&ts.RID,
			&ts.Rank,
			&ts.Status,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, ts)
	}
	return results, nil
}

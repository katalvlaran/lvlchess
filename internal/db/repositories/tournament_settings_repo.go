package repositories

import (
	"context"
	"fmt"

	"telega_chess/internal/db/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TournamentSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewTournamentSettingsRepository(pool *pgxpool.Pool) *TournamentSettingsRepository {
	return &TournamentSettingsRepository{pool: pool}
}

func (r *TournamentSettingsRepository) LinkRoomToTournament(ctx context.Context, tid, rid string, rank int) error {
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

func (r *TournamentSettingsRepository) UpdateTournamentRoomRank(ctx context.Context, tid, rid string, newRank int, newStatus int) error {
	sql := `
UPDATE tournament_settings
SET rank=$1,
    status=$2
WHERE t_id=$3 AND r_id=$4
`
	_, err := r.pool.Exec(ctx, sql, newRank, newStatus, tid, rid)
	if err != nil {
		return fmt.Errorf("UpdateTournamentRoomRank: %w", err)
	}
	return nil
}

// Пример: получить все rooms у турнира
func (r *TournamentSettingsRepository) GetRoomsByTournament(ctx context.Context, tid string) ([]models.TournamentSettings, error) {
	sql := `
SELECT t_id, r_id, rank, status
FROM tournament_settings
WHERE t_id=$1
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
		err := rows.Scan(&ts.TID, &ts.RID, &ts.Rank, &ts.Status)
		if err != nil {
			return nil, err
		}
		results = append(results, ts)
	}
	return results, nil
}

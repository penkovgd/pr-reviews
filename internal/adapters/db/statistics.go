package db

import (
	"context"
	"fmt"

	"github.com/penkovgd/closer"
)

func (d *DB) GetUserAssignmentStats(ctx context.Context) (map[string]int, error) {
	query := `
        SELECT 
            u.id as user_id,
            COALESCE(COUNT(prr.user_id), 0) as assignment_count
        FROM users u
        LEFT JOIN pull_request_reviewers prr ON u.id = prr.user_id
        WHERE u.is_active = true
        GROUP BY u.id
        ORDER BY assignment_count DESC
    `

	rows, err := d.conn.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get user assignment stats: %w", err)
	}
	defer closer.CloseOrLog(d.log, rows)

	stats := make(map[string]int)
	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, fmt.Errorf("scan user assignment stat: %w", err)
		}
		stats[userID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user assignment stats: %w", err)
	}

	return stats, nil
}

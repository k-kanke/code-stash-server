package repository

import (
	"context"
	"database/sql"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type collectionPGRepository struct {
	db *sql.DB
}

func NewCollectionPGRepository(db *sql.DB) usecaseRepo.CollectionRepository {
	return &collectionPGRepository{db: db}
}

func (r *collectionPGRepository) List(ctx context.Context, userID string) ([]entity.Collection, error) {
	const query = `
SELECT
	c.id,
	c.user_id,
	c.name,
	COALESCE(c.description, ''),
	c.created_at,
	c.updated_at,
	COALESCE(notes.note_count, 0)
FROM collections c
LEFT JOIN (
	SELECT collection_id, COUNT(*) AS note_count
	FROM notes
	GROUP BY collection_id
) notes ON notes.collection_id = c.id
WHERE c.user_id = $1
ORDER BY c.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collections := make([]entity.Collection, 0)
	for rows.Next() {
		var col entity.Collection
		if err := rows.Scan(
			&col.ID,
			&col.UserID,
			&col.Name,
			&col.Description,
			&col.CreatedAt,
			&col.UpdatedAt,
			&col.NoteCount,
		); err != nil {
			return nil, err
		}
		collections = append(collections, col)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return collections, nil
}

package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type notePGRepository struct {
	db *sql.DB
}

func NewNotePGRepository(db *sql.DB) usecaseRepo.NoteRepository {
	return &notePGRepository{db: db}
}

func (r *notePGRepository) ListByCollection(ctx context.Context, userID, collectionID string) ([]entity.Note, error) {
	const query = `
SELECT
	n.id,
	n.user_id,
	n.collection_id,
	n.folder_id,
	n.title,
	n.language,
	n.tags,
	n.code,
	n.note,
	n.created_at,
	n.updated_at
FROM notes n
WHERE n.user_id = $1 AND n.collection_id = $2
ORDER BY n.updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]entity.Note, 0)
	for rows.Next() {
		var note entity.Note
		var folder sql.NullString
		var body sql.NullString
		var tags pq.StringArray

		if err := rows.Scan(
			&note.ID,
			&note.UserID,
			&note.CollectionID,
			&folder,
			&note.Title,
			&note.Language,
			&tags,
			&note.Code,
			&body,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if folder.Valid {
			note.FolderID = &folder.String
		}
		if body.Valid {
			note.Note = body.String
		}
		if tags != nil {
			note.Tags = append([]string(nil), tags...)
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

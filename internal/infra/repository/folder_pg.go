package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type folderPGRepository struct {
	db *sql.DB
}

func NewFolderPGRepository(db *sql.DB) usecaseRepo.FolderRepository {
	return &folderPGRepository{db: db}
}

func (r *folderPGRepository) ListByCollection(ctx context.Context, userID, collectionID string) ([]entity.Folder, error) {
	const query = `
SELECT
	f.id,
	c.user_id,
	f.collection_id,
	f.parent_folder_id,
	f.name,
	f.sort_order,
	f.created_at,
	f.updated_at
FROM folders f
INNER JOIN collections c ON c.id = f.collection_id
WHERE c.user_id = $1 AND f.collection_id = $2
ORDER BY f.sort_order, f.created_at`

	rows, err := r.db.QueryContext(ctx, query, userID, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := make([]entity.Folder, 0)
	for rows.Next() {
		var folder entity.Folder
		var parent sql.NullString

		if err := rows.Scan(
			&folder.ID,
			&folder.UserID,
			&folder.CollectionID,
			&parent,
			&folder.Name,
			&folder.SortOrder,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if parent.Valid {
			folder.ParentFolderID = &parent.String
		} else {
			folder.ParentFolderID = nil
		}

		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

func (r *folderPGRepository) Create(ctx context.Context, userID, collectionID string, parentFolderID *string, name string) error {
	const query = `
INSERT INTO folders (id, collection_id, parent_folder_id, name)
SELECT $1, $2, $3, $4
WHERE EXISTS (
	SELECT 1 FROM collections WHERE id = $2 AND user_id = $5
)`

	var parent sql.NullString
	if parentFolderID != nil && *parentFolderID != "" {
		parent.Valid = true
		parent.String = *parentFolderID
	}

	result, err := r.db.ExecContext(ctx, query, uuid.NewString(), collectionID, parent, name, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

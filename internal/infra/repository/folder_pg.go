package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type folderPGRepository struct {
	db *sql.DB
}

type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
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

	parentValue, parentIsNull := normalizeFolderParent(parentFolderID)

	conflict, err := hasNameConflict(ctx, r.db, collectionID, parentValue, parentIsNull, name)
	if err != nil {
		return err
	}
	if conflict {
		return usecaseRepo.ErrFolderNameConflict
	}

	result, err := r.db.ExecContext(ctx, query, uuid.NewString(), collectionID, parentValue, name, userID)
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

	return touchCollection(ctx, r.db, userID, collectionID)
}

func (r *folderPGRepository) Delete(ctx context.Context, userID, collectionID, folderID string) error {
	const folderTreeCTE = `
WITH RECURSIVE folder_tree AS (
	SELECT f.id
	FROM folders f
	INNER JOIN collections c ON c.id = f.collection_id
	WHERE f.id = $1 AND f.collection_id = $2 AND c.user_id = $3
	UNION ALL
	SELECT child.id
	FROM folders child
	INNER JOIN folder_tree ft ON child.parent_folder_id = ft.id
)`

	const deleteNotesQuery = folderTreeCTE + `
DELETE FROM notes
WHERE collection_id = $2
  AND user_id = $3
  AND folder_id IN (SELECT id FROM folder_tree)`

	const deleteFoldersQuery = folderTreeCTE + `
DELETE FROM folders
WHERE collection_id = $2
  AND id IN (SELECT id FROM folder_tree)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, deleteNotesQuery, folderID, collectionID, userID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, deleteFoldersQuery, folderID, collectionID, userID)
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

	if err := touchCollection(ctx, tx, userID, collectionID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func touchCollection(ctx context.Context, exec execContext, userID, collectionID string) error {
	const update = `UPDATE collections SET updated_at = now() WHERE user_id = $1 AND id = $2`
	_, err := exec.ExecContext(ctx, update, userID, collectionID)
	return err
}

func normalizeFolderParent(parentFolderID *string) (any, bool) {
	if parentFolderID == nil {
		return nil, true
	}
	if trimmed := strings.TrimSpace(*parentFolderID); trimmed != "" {
		return trimmed, false
	}
	return nil, true
}

func hasNameConflict(ctx context.Context, db *sql.DB, collectionID string, parent any, parentIsNull bool, name string) (bool, error) {
	const folderConflictQuery = `
SELECT 1 FROM folders
WHERE collection_id = $1
  AND (
    ($3 AND parent_folder_id IS NULL)
    OR parent_folder_id = $2
  )
  AND name = $4
LIMIT 1`

	const noteConflictQuery = `
SELECT 1 FROM notes
WHERE collection_id = $1
  AND (
    ($3 AND folder_id IS NULL)
    OR folder_id = $2
  )
  AND title = $4
LIMIT 1`

	if exists, err := recordExists(ctx, db, folderConflictQuery, collectionID, parent, parentIsNull, name); err != nil || exists {
		return exists, err
	}

	return recordExists(ctx, db, noteConflictQuery, collectionID, parent, parentIsNull, name)
}

func recordExists(ctx context.Context, db *sql.DB, query string, args ...any) (bool, error) {
	var dummy int
	err := db.QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

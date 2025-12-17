package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
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

func (r *notePGRepository) Get(ctx context.Context, userID, noteID string) (*entity.Note, error) {
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
WHERE n.user_id = $1 AND n.id = $2`

	var note entity.Note
	var folder sql.NullString
	var body sql.NullString
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, userID, noteID).Scan(
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
	)
	if err != nil {
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

	return &note, nil
}

func (r *notePGRepository) Create(ctx context.Context, userID, collectionID string, folderID *string, title, code, language, noteBody string, tags []string) (string, error) {
	const query = `
INSERT INTO notes (id, collection_id, folder_id, user_id, title, code, language, note, tags)
SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
WHERE EXISTS (
	SELECT 1 FROM collections WHERE id = $2 AND user_id = $4
) AND (
	($3)::uuid IS NULL OR EXISTS (
		SELECT 1 FROM folders WHERE id = ($3)::uuid AND collection_id = $2
	)
)`

	parentValue, parentIsNull := normalizeFolderParent(folderID)

	conflict, err := hasNameConflict(ctx, r.db, collectionID, parentValue, parentIsNull, title)
	if err != nil {
		return "", err
	}
	if conflict {
		return "", usecaseRepo.ErrNoteTitleConflict
	}

	var folder sql.NullString
	if !parentIsNull {
		if str, ok := parentValue.(string); ok {
			folder.Valid = true
			folder.String = str
		}
	}

	noteID := uuid.NewString()
	result, err := r.db.ExecContext(
		ctx,
		query,
		noteID,
		collectionID,
		folder,
		userID,
		title,
		code,
		language,
		noteBody,
		pq.Array(tags),
	)
	if err != nil {
		return "", err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if rows == 0 {
		return "", sql.ErrNoRows
	}

	if err := touchCollection(ctx, r.db, userID, collectionID); err != nil {
		return "", err
	}
	return noteID, nil
}

func (r *notePGRepository) Update(ctx context.Context, userID, noteID string, update usecaseRepo.NoteUpdate) error {
	setClauses := []string{"updated_at = now()"}
	args := make([]any, 0, 8)
	idx := 1

	if update.Title != nil {
		setClauses = append(setClauses, "title = $"+strconv.Itoa(idx))
		args = append(args, *update.Title)
		idx++
	}
	if update.Language != nil {
		setClauses = append(setClauses, "language = $"+strconv.Itoa(idx))
		args = append(args, *update.Language)
		idx++
	}
	if update.Code != nil {
		setClauses = append(setClauses, "code = $"+strconv.Itoa(idx))
		args = append(args, *update.Code)
		idx++
	}
	if update.Note != nil {
		setClauses = append(setClauses, "note = $"+strconv.Itoa(idx))
		args = append(args, *update.Note)
		idx++
	}
	if update.Tags != nil {
		setClauses = append(setClauses, "tags = $"+strconv.Itoa(idx))
		args = append(args, pq.Array(*update.Tags))
		idx++
	}
	if update.FolderID != nil {
		var folder sql.NullString
		if trimmed := strings.TrimSpace(*update.FolderID); trimmed != "" {
			folder.Valid = true
			folder.String = trimmed
		}
		setClauses = append(setClauses, "folder_id = $"+strconv.Itoa(idx))
		args = append(args, folder)
		idx++
	}

	if len(setClauses) == 1 {
		return nil
	}

	query := `
UPDATE notes
SET ` + strings.Join(setClauses, ", ") + `
WHERE user_id = $` + strconv.Itoa(idx) + ` AND id = $` + strconv.Itoa(idx+1)

	args = append(args, userID, noteID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return touchCollectionByNote(ctx, r.db, userID, noteID)
}

func (r *notePGRepository) Delete(ctx context.Context, userID, noteID string) error {
	const query = `DELETE FROM notes WHERE user_id = $1 AND id = $2`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := touchCollectionByNote(ctx, tx, userID, noteID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, query, userID, noteID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func touchCollectionByNote(ctx context.Context, exec execContext, userID, noteID string) error {
	const query = `
UPDATE collections AS c
SET updated_at = now()
FROM notes n
WHERE n.id = $1
  AND n.user_id = $2
  AND n.collection_id = c.id
  AND c.user_id = $2`

	result, err := exec.ExecContext(ctx, query, noteID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

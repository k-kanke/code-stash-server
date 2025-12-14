package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type noteCommentPGRepository struct {
	db *sql.DB
}

func NewNoteCommentPGRepository(db *sql.DB) usecaseRepo.NoteCommentRepository {
	return &noteCommentPGRepository{db: db}
}

func (r *noteCommentPGRepository) ListByNote(ctx context.Context, userID, noteID string) ([]entity.NoteComment, error) {
	const query = `
SELECT
	c.id,
	c.note_id,
	c.author_id,
	c.body,
	c.line_start,
	c.line_end,
	c.resolved,
	c.created_at,
	c.updated_at
FROM note_comments c
JOIN notes n ON c.note_id = n.id
WHERE n.user_id = $1 AND c.note_id = $2
ORDER BY c.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, userID, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]entity.NoteComment, 0)
	for rows.Next() {
		var comment entity.NoteComment
		var lineStart sql.NullInt32
		var lineEnd sql.NullInt32
		if err := rows.Scan(
			&comment.ID,
			&comment.NoteID,
			&comment.AuthorID,
			&comment.Body,
			&lineStart,
			&lineEnd,
			&comment.Resolved,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if lineStart.Valid {
			v := int(lineStart.Int32)
			comment.LineStart = &v
		}
		if lineEnd.Valid {
			v := int(lineEnd.Int32)
			comment.LineEnd = &v
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *noteCommentPGRepository) Create(ctx context.Context, in usecaseRepo.CreateNoteCommentInput) (*entity.NoteComment, error) {
	const query = `
INSERT INTO note_comments (id, note_id, author_id, body, line_start, line_end)
SELECT $1, $2, $3, $4, $5, $6
FROM notes n
WHERE n.id = $2 AND n.user_id = $3
RETURNING id, note_id, author_id, body, line_start, line_end, resolved, created_at, updated_at`

	lineStart := sql.NullInt32{}
	if in.LineStart != nil {
		lineStart.Valid = true
		lineStart.Int32 = int32(*in.LineStart)
	}
	lineEnd := sql.NullInt32{}
	if in.LineEnd != nil {
		lineEnd.Valid = true
		lineEnd.Int32 = int32(*in.LineEnd)
	}

	commentID := uuid.NewString()

	var comment entity.NoteComment
	var insertedLineStart sql.NullInt32
	var insertedLineEnd sql.NullInt32

	err := r.db.QueryRowContext(
		ctx,
		query,
		commentID,
		in.NoteID,
		in.UserID,
		in.Body,
		lineStart,
		lineEnd,
	).Scan(
		&comment.ID,
		&comment.NoteID,
		&comment.AuthorID,
		&comment.Body,
		&insertedLineStart,
		&insertedLineEnd,
		&comment.Resolved,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if insertedLineStart.Valid {
		v := int(insertedLineStart.Int32)
		comment.LineStart = &v
	}
	if insertedLineEnd.Valid {
		v := int(insertedLineEnd.Int32)
		comment.LineEnd = &v
	}

	return &comment, nil
}

func (r *noteCommentPGRepository) Update(ctx context.Context, in usecaseRepo.UpdateNoteCommentInput) (*entity.NoteComment, error) {
	setClauses := []string{"updated_at = now()"}
	args := make([]any, 0, 6)
	idx := 1

	if in.Body != nil {
		setClauses = append(setClauses, "body = $"+strconv.Itoa(idx))
		args = append(args, *in.Body)
		idx++
	}
	if in.LineStart != nil {
		setClauses = append(setClauses, "line_start = $"+strconv.Itoa(idx))
		args = append(args, *in.LineStart)
		idx++
	}
	if in.LineEnd != nil {
		setClauses = append(setClauses, "line_end = $"+strconv.Itoa(idx))
		args = append(args, *in.LineEnd)
		idx++
	}
	if in.Resolved != nil {
		setClauses = append(setClauses, "resolved = $"+strconv.Itoa(idx))
		args = append(args, *in.Resolved)
		idx++
	}

	setClause := strings.Join(setClauses, ", ")

	query := `
UPDATE note_comments AS c
SET ` + setClause + `
FROM notes n
WHERE c.note_id = n.id
  AND n.user_id = $` + strconv.Itoa(idx) + `
  AND c.id = $` + strconv.Itoa(idx+1) + `
RETURNING c.id, c.note_id, c.author_id, c.body, c.line_start, c.line_end, c.resolved, c.created_at, c.updated_at`

	args = append(args, in.UserID, in.CommentID)

	var updated entity.NoteComment
	var lineStart sql.NullInt32
	var lineEnd sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&updated.ID,
		&updated.NoteID,
		&updated.AuthorID,
		&updated.Body,
		&lineStart,
		&lineEnd,
		&updated.Resolved,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if lineStart.Valid {
		v := int(lineStart.Int32)
		updated.LineStart = &v
	}
	if lineEnd.Valid {
		v := int(lineEnd.Int32)
		updated.LineEnd = &v
	}

	return &updated, nil
}

func (r *noteCommentPGRepository) Delete(ctx context.Context, userID, commentID string) error {
	const query = `
DELETE FROM note_comments AS c
USING notes n
WHERE c.note_id = n.id
  AND n.user_id = $1
  AND c.id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, commentID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

ALTER TABLE note_comments
  ADD COLUMN parent_comment_id UUID REFERENCES note_comments(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_note_comments_parent ON note_comments(parent_comment_id);

CREATE TABLE note_comments (
  id           UUID PRIMARY KEY,
  note_id      UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  author_id    UUID NOT NULL REFERENCES users(id),
  body         TEXT NOT NULL,
  line_start   INT,
  line_end     INT,
  resolved     BOOLEAN NOT NULL DEFAULT FALSE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_note_comments_note ON note_comments(note_id);
CREATE INDEX idx_note_comments_author ON note_comments(author_id);

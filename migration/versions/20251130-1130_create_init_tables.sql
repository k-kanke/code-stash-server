CREATE TABLE users (
  id         UUID PRIMARY KEY,
  name       TEXT NOT NULL,
  email      TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE collections (
  id          UUID PRIMARY KEY,
  user_id     UUID NOT NULL REFERENCES users(id),
  name        TEXT NOT NULL,
  description TEXT,
  note_count  INT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_collections_user ON collections(user_id);

CREATE TABLE folders (
  id              UUID PRIMARY KEY,
  collection_id   UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
  parent_folder_id UUID REFERENCES folders(id),
  name            TEXT NOT NULL,
  sort_order      INT NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_folders_collection ON folders(collection_id);
CREATE INDEX idx_folders_parent ON folders(parent_folder_id);

CREATE TABLE notes (
  id           UUID PRIMARY KEY,
  collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
  folder_id    UUID REFERENCES folders(id),
  user_id      UUID NOT NULL REFERENCES users(id),
  title        TEXT NOT NULL,
  code         TEXT NOT NULL,
  language     TEXT NOT NULL,
  note         TEXT,
  tags         TEXT[],
  is_public    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notes_collection ON notes(collection_id);
CREATE INDEX idx_notes_folder ON notes(folder_id);
CREATE INDEX idx_notes_user ON notes(user_id);
CREATE INDEX idx_notes_tags_gin ON notes USING GIN (tags);
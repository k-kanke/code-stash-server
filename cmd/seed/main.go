package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

const seedDir = "seed/data"

type userSeed struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type collectionSeed struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NoteCount   int    `json:"note_count"`
}

type folderSeed struct {
	ID             string `json:"id"`
	CollectionID   string `json:"collection_id"`
	ParentFolderID string `json:"parent_folder_id"`
	Name           string `json:"name"`
	SortOrder      int    `json:"sort_order"`
}

type noteSeed struct {
	ID           string   `json:"id"`
	CollectionID string   `json:"collection_id"`
	FolderID     string   `json:"folder_id"`
	UserID       string   `json:"user_id"`
	Title        string   `json:"title"`
	Code         string   `json:"code"`
	Language     string   `json:"language"`
	Note         string   `json:"note"`
	Tags         []string `json:"tags"`
}

func main() {
	db, err := newDB()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	if err := seedDatabase(db); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	log.Println("✅ Seed data loaded successfully")
}

func seedDatabase(db *sql.DB) error {
	ctx := context.Background()

	users, err := loadSeed[userSeed]("users.json")
	if err != nil {
		return err
	}
	collections, err := loadSeed[collectionSeed]("collections.json")
	if err != nil {
		return err
	}
	folders, err := loadSeed[folderSeed]("folders.json")
	if err != nil {
		return err
	}
	notes, err := loadSeed[noteSeed]("notes.json")
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	steps := []func(context.Context, *sql.Tx) error{
		func(ctx context.Context, tx *sql.Tx) error { return seedUsers(ctx, tx, users) },
		func(ctx context.Context, tx *sql.Tx) error { return seedCollections(ctx, tx, collections) },
		func(ctx context.Context, tx *sql.Tx) error { return seedFolders(ctx, tx, folders) },
		func(ctx context.Context, tx *sql.Tx) error { return seedNotes(ctx, tx, notes) },
		refreshNoteCounts,
	}

	for _, step := range steps {
		if err := step(ctx, tx); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func seedUsers(ctx context.Context, tx *sql.Tx, users []userSeed) error {
	const query = `INSERT INTO users (id, name, email)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email`

	for _, u := range users {
		if _, err := tx.ExecContext(ctx, query, u.ID, u.Name, u.Email); err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
	}
	return nil
}

func seedCollections(ctx context.Context, tx *sql.Tx, collections []collectionSeed) error {
	const query = `INSERT INTO collections (id, user_id, name, description, note_count)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description`

	for _, c := range collections {
		if _, err := tx.ExecContext(ctx, query, c.ID, c.UserID, c.Name, c.Description, c.NoteCount); err != nil {
			return fmt.Errorf("seed collections: %w", err)
		}
	}
	return nil
}

func seedFolders(ctx context.Context, tx *sql.Tx, folders []folderSeed) error {
	const query = `INSERT INTO folders (id, collection_id, parent_folder_id, name, sort_order)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  sort_order = EXCLUDED.sort_order`

	for _, f := range folders {
		var parent interface{}
		if f.ParentFolderID != "" {
			parent = f.ParentFolderID
		} else {
			parent = nil
		}

		if _, err := tx.ExecContext(ctx, query, f.ID, f.CollectionID, parent, f.Name, f.SortOrder); err != nil {
			return fmt.Errorf("seed folders: %w", err)
		}
	}
	return nil
}

func seedNotes(ctx context.Context, tx *sql.Tx, notes []noteSeed) error {
	const query = `INSERT INTO notes (id, collection_id, folder_id, user_id, title, code, language, note, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE SET
  title = EXCLUDED.title,
  code = EXCLUDED.code,
  language = EXCLUDED.language,
  note = EXCLUDED.note,
  tags = EXCLUDED.tags`

	for _, n := range notes {
		var folder interface{}
		if n.FolderID != "" {
			folder = n.FolderID
		} else {
			folder = nil
		}

		if _, err := tx.ExecContext(ctx, query, n.ID, n.CollectionID, folder, n.UserID, n.Title, n.Code, n.Language, n.Note, pq.Array(n.Tags)); err != nil {
			return fmt.Errorf("seed notes: %w", err)
		}
	}
	return nil
}

func refreshNoteCounts(ctx context.Context, tx *sql.Tx) error {
	const query = `UPDATE collections c
SET note_count = sub.count
FROM (
  SELECT collection_id, COUNT(*) AS count
  FROM notes
  GROUP BY collection_id
) sub
WHERE c.id = sub.collection_id`

	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("update note counts: %w", err)
	}
	return nil
}

func loadSeed[T any](filename string) ([]T, error) {
	path := filepath.Join(seedDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read seed file %s: %w", filename, err)
	}

	var result []T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", filename, err)
	}
	return result, nil
}

func newDB() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5437")
		user := getEnv("DB_USER", "root")
		password := getEnv("DB_PASSWORD", "password")
		dbName := getEnv("DB_NAME", "codestash")
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

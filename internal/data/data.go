package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/akikareha/himewiki/internal/config"
)

var db *pgxpool.Pool

const createTablesSql = `
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS revisions (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_revisions_name_created_at
	ON revisions (name, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_revisions_name_id
	ON revisions (name, id DESC);

CREATE TABLE IF NOT EXISTS pages (
	name TEXT PRIMARY KEY,
	content TEXT NOT NULL,
	revision_id INT NOT NULL REFERENCES revisions(id),
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pages_name_trgm
	ON pages USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_pages_content_trgm
	ON pages USING gin (content gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_pages_updated_at
	ON pages(updated_at DESC);

CREATE TABLE IF NOT EXISTS image_revisions (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	content BYTEA NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_image_revisions_name_created_at
	ON image_revisions (name, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_image_revisions_name_id
	ON image_revisions (name, id DESC);

CREATE TABLE IF NOT EXISTS images (
	name TEXT PRIMARY KEY,
	content BYTEA NOT NULL,
	revision_id INT NOT NULL REFERENCES image_revisions(id),
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_images_name_trgm
	ON images USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_images_updated_at
	ON images(updated_at DESC);
`

func Connect(cfg *config.Config) *pgxpool.Pool {
	var err error

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	_, err = db.Exec(context.Background(), createTablesSql)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	return db
}

func Load(name string) (int, string, error) {
	var id int
	var content string
	err := db.QueryRow(context.Background(),
		"SELECT revision_id, content FROM pages WHERE name=$1", name).
		Scan(&id, &content)
	if err != nil {
		return 0, "", err
	}

	return id, content, nil
}

func diff(oldText, newText string) string {
	diff := difflib.UnifiedDiff{
		A: difflib.SplitLines(oldText),
		B: difflib.SplitLines(newText),
		FromFile: "old",
		ToFile:  "new",
		Context: 3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
}

func Save(name, content string, baseRevID int) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentRevID int
	err = tx.QueryRow(ctx, "SELECT revision_id FROM pages WHERE name=$1", name).
		Scan(&currentRevID)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	if currentRevID != 0 && currentRevID != baseRevID {
		return fmt.Errorf("edit conflict")
	}

	var newRevID int
	err = tx.QueryRow(ctx,
		`INSERT INTO revisions (name, content, created_at)
		 VALUES ($1, $2, now())
		 RETURNING id`,
		name, content).Scan(&newRevID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO pages (name, content, revision_id, updated_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (name) DO UPDATE
		 SET content=EXCLUDED.content,
		     revision_id=EXCLUDED.revision_id,
		     updated_at=now()`,
		name, content, newRevID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func LoadAll() ([]string, error) {
	rows, err := db.Query(context.Background(),
		"SELECT name FROM pages ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		results = append(results, name)
	}
	return results, nil
}

type RecentRecord struct {
	Name string
	Diff string
}

func Recent() ([]RecentRecord, error) {
	rows, err := db.Query(context.Background(),
		`SELECT
			p.name,
			r1.content AS content,
			r2.content AS prev_content
		 FROM pages p
		 JOIN revisions r1 ON r1.id = p.revision_id
		 LEFT JOIN revisions r2
			ON r2.name = p.name
			AND r2.id = (
				SELECT id
				FROM revisions
				WHERE name = p.name
				AND id < p.revision_id
				ORDER BY id DESC
				LIMIT 1
			)
		 ORDER BY updated_at DESC, name ASC
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RecentRecord
	for rows.Next() {
		var name, content string
		var prevContent sql.NullString
		if err := rows.Scan(&name, &content, &prevContent); err != nil {
			return nil, err
		}
		var diffText string
		if prevContent.Valid {
			diffText = diff(prevContent.String, content)
		} else {
			diffText = diff("", content)
		}
		record := RecentRecord{ Name: name, Diff: diffText }
		results = append(results, record)
	}
	return results, nil
}

type Revision struct {
	ID int
	Name string
	Content string
	Diff string
	CreatedAt time.Time
}

func LoadRevisions(name string) ([]Revision, error) {
	rows, err := db.Query(context.Background(),
		`SELECT id, name, content, created_at
		 FROM revisions
		 WHERE name=$1
		 ORDER BY created_at DESC`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revs []Revision
	for rows.Next() {
		var r Revision
		if err := rows.Scan(&r.ID, &r.Name, &r.Content, &r.CreatedAt); err != nil {
			return nil, err
		}
		revs = append(revs, r)
	}

	for i := 0; i < len(revs) - 1; i++ {
		revs[i].Diff = diff(revs[i + 1].Content, revs[i].Content)
	}
	if len(revs) > 0 {
		revs[len(revs) - 1].Diff = diff("", revs[len(revs) - 1].Content)
	}
	return revs, nil
}

func Revert(name string, revID int) error {
	var content string
	err := db.QueryRow(context.Background(),
		"SELECT content FROM revisions WHERE id=$1 AND name=$2",
		revID, name).Scan(&content)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(),
		"UPDATE pages SET content=$1, revision_id=$2, updated_at=now() WHERE name=$3",
		content, revID, name)

	return err
}

func LoadRevision(name string, revID int) (string, error) {
	var content string
	var created time.Time
	err := db.QueryRow(context.Background(),
		"SELECT content, created_at FROM revisions WHERE id=$1 AND name=$2",
		revID, name).Scan(&content, &created)

	return content, err
}

func SearchNames(word string) ([]string, error) {
	rows, err := db.Query(context.Background(),
		"SELECT name FROM pages WHERE name ILIKE '%' || $1 || '%' ORDER BY name",
		word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		results = append(results, name)
	}
	return results, nil
}

func SearchContents(word string) ([]string, error) {
	rows, err := db.Query(context.Background(),
		"SELECT name FROM pages WHERE content ILIKE '%' || $1 || '%' ORDER BY name",
		word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		results = append(results, name)
	}
	return results, nil
}

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/util"
)

var db *pgxpool.Pool

const createTablesSql = `
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS pages (
	name TEXT PRIMARY KEY,
	content TEXT NOT NULL,
	revision_id INT NOT NULL,
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pages_name_trgm
	ON pages USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_pages_content_trgm
	ON pages USING gin (content gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_pages_updated_at
	ON pages(updated_at DESC);

ALTER TABLE pages SET (autovacuum_enabled = true);

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

ALTER TABLE revisions SET (autovacuum_enabled = false);

CREATE TABLE IF NOT EXISTS images (
	name TEXT PRIMARY KEY,
	content BYTEA NOT NULL,
	revision_id INT NOT NULL,
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_images_name_trgm
	ON images USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_images_updated_at
	ON images(updated_at DESC);

ALTER TABLE images SET (autovacuum_enabled = true);

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

ALTER TABLE image_revisions SET (autovacuum_enabled = false);

CREATE TABLE IF NOT EXISTS state (
	id INT PRIMARY KEY DEFAULT 1,
	boot_counter BIGINT NOT NULL DEFAULT 0,
	page_counter BIGINT NOT NULL DEFAULT 0,
	image_counter BIGINT NOT NULL DEFAULT 0
);

ALTER TABLE state SET (autovacuum_enabled = true);
`

var bootCount int64

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

	err = db.QueryRow(context.Background(), `
		INSERT INTO state (id, boot_counter, page_counter, image_counter)
		VALUES (1, 1, 0, 0)
		ON CONFLICT (id)
		DO UPDATE SET boot_counter = state.boot_counter + 1
		RETURNING boot_counter;
	`).Scan(&bootCount)
	if err != nil {
		log.Fatalf("failed to count up boot conter: %v", err)
	}

	return db
}

type Info struct {
	Size string
	PageCount int
	RevisionCount int
	ImageCount int
	ImageRevisionCount int
}

func Stat() Info {
	var size string
	err := db.QueryRow(context.Background(), "SELECT pg_size_pretty(pg_database_size('himewiki'))").Scan(&size)
	if err != nil {
		size = "unknown"
	}

	var pageCount, revisionCount, imageCount, imageRevisionCount int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM pages").Scan(&pageCount)
	if err != nil {
		pageCount = -1
	}
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM revisions").Scan(&revisionCount)
	if err != nil {
		revisionCount = -1
	}
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM images").Scan(&imageCount)
	if err != nil {
		imageCount = -1
	}
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM image_revisions").Scan(&imageRevisionCount)
	if err != nil {
		imageRevisionCount = -1
	}

	return Info {
		Size: size,
		PageCount: pageCount,
		RevisionCount: revisionCount,
		ImageCount: imageCount,
		ImageRevisionCount: imageRevisionCount,
	}
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

func LoadPrev(name string) (int, string, error) {
	rows, err := db.Query(context.Background(),
		`SELECT id, content
		 FROM revisions
		 WHERE name=$1
		 ORDER BY created_at DESC
		 LIMIT 2
		`, name)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	var ids []int
	var contents []string
	for rows.Next() {
		var id int
		var content string
		if err := rows.Scan(&id, &content); err != nil {
			return 0, "", err
		}
		ids = append(ids, id)
		contents = append(contents, content)
	}

	if len(ids) < 2 {
		return 0, "", errors.New("no previous revisions")
	} else {
		return ids[1], contents[1], nil
	}
}

func Save(cfg *config.Config, name, content string, baseRevID int) (int64, error) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var currentRevID int
	err = tx.QueryRow(ctx, "SELECT revision_id FROM pages WHERE name=$1", name).
		Scan(&currentRevID)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}

	if currentRevID != 0 && currentRevID != baseRevID {
		return 0, fmt.Errorf("edit conflict")
	}

	var newRevID int
	err = tx.QueryRow(ctx,
		`INSERT INTO revisions (name, content, created_at)
		 VALUES ($1, $2, now())
		 RETURNING id`,
		name, content).Scan(&newRevID)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	var pageCount int64
	err = tx.QueryRow(ctx, `
		UPDATE state
		SET page_counter = page_counter + 1
		WHERE id = 1
		RETURNING page_counter
	`).Scan(&pageCount)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	if pageCount % int64(cfg.Vacuum.CheckEvery) == 0 {
		var sizeBytes int64
		err = db.QueryRow(ctx, `
			SELECT pg_total_relation_size('pages') +
			       pg_total_relation_size('revisions')
		`).Scan(&sizeBytes)
		if err != nil {
			return 0, err
		}

		if sizeBytes >= cfg.Vacuum.Threshold {
			_, err = db.Exec(ctx, "VACUUM FULL pages")
			if err != nil {
				return 0, err
			}

			_, err = db.Exec(ctx, `
				WITH cutoff AS (
					SELECT COUNT(*) / 2 AS limit_count
					FROM revisions
					GROUP BY name
					ORDER BY COUNT(*) DESC
					LIMIT 1
				)
				DELETE FROM revisions r
				USING cutoff
				WHERE r.id IN (
					SELECT id
					FROM revisions r2, cutoff
					WHERE r2.name = r.name
					ORDER BY r2.created_at ASC
					OFFSET cutoff.limit_count
				)
			`)
			if err != nil {
				return 0, err
			}

			_, err = db.Exec(ctx, "VACUUM FULL revisions")
			if err != nil {
				return 0, err
			}
		}
	}

	return pageCount, nil
}

func LoadAll(page int, perPage int) ([]string, error) {
	if page < 1 {
		return nil, errors.New("invalid page")
	}
	if perPage < 1 {
		return nil, errors.New("invalid perPage")
	}
	offset := (page - 1) * perPage

	rows, err := db.Query(context.Background(),
		"SELECT name FROM pages ORDER BY name LIMIT $1 OFFSET $2",
		perPage, offset)
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

func Recent(page int, perPage int) ([]RecentRecord, error) {
	if page < 1 {
		return nil, errors.New("invalid page")
	}
	if perPage < 1 {
		return nil, errors.New("invalid perPage")
	}
	offset := (page - 1) * perPage

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
		 LIMIT $1 OFFSET $2
		`, perPage, offset)
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
			diffText = util.Diff(prevContent.String, content)
		} else {
			diffText = util.Diff("", content)
		}
		record := RecentRecord{ Name: name, Diff: diffText }
		results = append(results, record)
	}
	return results, nil
}

func RecentNames(limit int) ([]string, error) {
	if limit < 0 {
		return nil, errors.New("invalid limit")
	}

	rows, err := db.Query(context.Background(),
		`SELECT name FROM pages
		 ORDER BY updated_at DESC, name ASC
		 LIMIT $1
		`, limit)
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

type Revision struct {
	ID int
	Name string
	Content string
	Diff string
	CreatedAt time.Time
}

func LoadRevisions(name string, page int, perPage int) ([]Revision, error) {
	if page < 1 {
		return nil, errors.New("invalid page")
	}
	if perPage < 1 {
		return nil, errors.New("invalid perPage")
	}
	offset := (page - 1) * perPage

	rows, err := db.Query(context.Background(),
		`SELECT id, name, content, created_at
		 FROM revisions
		 WHERE name=$1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3
		`, name, perPage + 1, offset)
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
		revs[i].Diff = util.Diff(revs[i + 1].Content, revs[i].Content)
	}
	if len(revs) > 0 && len(revs) < perPage + 1 {
		revs[len(revs) - 1].Diff = util.Diff("", revs[len(revs) - 1].Content)
	}
	if len(revs) > perPage {
		revs = revs[:perPage]
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

func SearchNames(word string, page int, perPage int) ([]string, error) {
	if page < 1 {
		return nil, errors.New("invalid page")
	}
	if perPage < 1 {
		return nil, errors.New("invalid perPage")
	}
	offset := (page - 1) * perPage

	rows, err := db.Query(context.Background(),
		`SELECT name FROM pages WHERE name ILIKE '%' || $1 || '%'
		 ORDER BY name
		 LIMIT $2 OFFSET $3
		`, word, perPage, offset)
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

func SearchContents(word string, page int, perPage int) ([]string, error) {
	if page < 1 {
		return nil, errors.New("invalid page")
	}
	if perPage < 1 {
		return nil, errors.New("invalid perPage")
	}
	offset := (page - 1) * perPage

	rows, err := db.Query(context.Background(),
		`SELECT name FROM pages WHERE content ILIKE '%' || $1 || '%'
		 ORDER BY name
		 LIMIT $2 OFFSET $3
		`, word, perPage, offset)
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

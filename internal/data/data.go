package data

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/akikareha/himewiki/internal/config"
)

var db *pgxpool.Pool

const createTablesSql = `
CREATE TABLE IF NOT EXISTS pages (
	name TEXT PRIMARY KEY,
	content TEXT NOT NULL,
	revision_id INT NOT NULL REFERENCES revisions(id),
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS revisions (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	content TEXT NOT NULL,
	diff TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_revisions_name_created_at
	ON revisions (name, created_at DESC);
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
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	var result []string

	max := len(oldLines)
	if len(newLines) > max {
		max = len(newLines)
	}

	for i := 0; i < max; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		switch {
		case oldLine == newLine:
			result = append(result, " " + oldLine)
		case oldLine == "":
			result = append(result, "+" + newLine)
		case newLine == "":
			result = append(result, "-" + oldLine)
		default:
			result = append(result, "-" + oldLine)
			result = append(result, "+" + newLine)
		}
	}

	return strings.Join(result, "\n")
}

func Save(name, newContent string, baseRevID int) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentRevID int
	var oldContent string
	err = tx.QueryRow(ctx, "SELECT revision_id, content FROM pages WHERE name=$1", name).
		Scan(&currentRevID, &oldContent)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	if currentRevID != 0 && currentRevID != baseRevID {
		return fmt.Errorf("edit conflict")
	}

	diffText := diff(oldContent, newContent)

	var newRevID int
	err = tx.QueryRow(ctx,
		`INSERT INTO revisions (name, content, diff, created_at)
		 VALUES ($1, $2, $3, now())
		 RETURNING id`,
		name, newContent, diffText).Scan(&newRevID)
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
		name, newContent, newRevID)
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

type Revision struct {
	ID int
	Name string
	Diff string
	CreatedAt time.Time
}

func LoadRevisions(name string) ([]Revision, error) {
	rows, err := db.Query(context.Background(),
		`SELECT id, name, diff, created_at
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
		if err := rows.Scan(&r.ID, &r.Name, &r.Diff, &r.CreatedAt); err != nil {
			return nil, err
		}
		revs = append(revs, r)
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

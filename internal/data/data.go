package data

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/format"
)

var db *pgxpool.Pool

const createTablesSql = `
CREATE TABLE IF NOT EXISTS pages (
	name TEXT PRIMARY KEY,
	content TEXT NOT NULL,
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

func Load(name string) (string, error) {
	var content string
	err := db.QueryRow(context.Background(),
		"SELECT content FROM pages WHERE name=$1", name).
		Scan(&content)
	if err != nil {
		return "", err
	}

	return content, nil
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

func Save(name, newContent string) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var oldContent string
	_ = tx.QueryRow(ctx, "SELECT content FROM pages WHERE name=$1", name).Scan(&oldContent)

	diffText := diff(oldContent, newContent)

	_, err = tx.Exec(ctx,
		`INSERT INTO pages (name, content, updated_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (name) DO UPDATE
		 SET content=EXCLUDED.content, updated_at=now()`,
		name, newContent)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO revisions (name, content, diff, created_at)
		 VALUES ($1, $2, $3, now())`,
		name, newContent, diffText)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type Name struct {
	Raw string
	Escaped string
}

func LoadAll() ([]Name, error) {
	rows, err := db.Query(context.Background(),
		"SELECT name FROM pages ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Name
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		escaped := url.PathEscape(name)
		result := Name{ Raw: name, Escaped: escaped }
		results = append(results, result)
	}
	return results, nil
}

type Revision struct {
	ID int
	Name string
	Diff string
	CreatedAt time.Time
	Escaped string
	DiffHTML template.HTML
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
		r.Escaped = url.PathEscape(r.Name)
		r.DiffHTML = template.HTML(format.Diff(r.Diff))
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
		"UPDATE pages SET content=$1, updated_at=now() WHERE name=$2",
		content, name)

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

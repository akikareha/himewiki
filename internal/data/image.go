package data

import (
	"context"
	"errors"

	"github.com/akikareha/himewiki/internal/config"
)

func LoadImage(name string) (int, []byte, error) {
	var id int
	var content []byte
	err := db.QueryRow(context.Background(),
		"SELECT revision_id, content FROM images WHERE name=$1", name).
		Scan(&id, &content)
	if err != nil {
		return 0, nil, err
	}

	return id, content, nil
}

func SaveImage(cfg *config.Config, name string, content []byte) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var newRevID int
	err = tx.QueryRow(ctx,
		`INSERT INTO image_revisions (name, content, created_at)
		 VALUES ($1, $2, now())
		 RETURNING id`,
		name, content).Scan(&newRevID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO images (name, content, revision_id, updated_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (name) DO UPDATE
		 SET content=EXCLUDED.content,
		     revision_id=EXCLUDED.revision_id,
		     updated_at=now()`,
		name, content, newRevID)
	if err != nil {
		return err
	}

	var imageCount int64
	err = tx.QueryRow(ctx, `
		UPDATE state
		SET image_counter = image_counter + 1
		WHERE id = 1
		RETURNING image_counter
	`).Scan(&imageCount)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	if imageCount % int64(cfg.Vacuum.CheckEvery) == 0 {
		var sizeBytes int64
		err = db.QueryRow(ctx, `
			SELECT pg_total_relation_size('images') +
			       pg_total_relation_size('image_revisions')
		`).Scan(&sizeBytes)
		if err != nil {
			return err
		}

		if sizeBytes >= cfg.Vacuum.ImageThreshold {
			_, err = db.Exec(ctx, "VACUUM FULL images")
			if err != nil {
				return err
			}

			_, err = db.Exec(ctx, `
				WITH cutoff AS (
					SELECT COUNT(*) / 2 AS limit_count
					FROM image_revisions
					GROUP BY name
					ORDER BY COUNT(*) DESC
					LIMIT 1
				)
				DELETE FROM image_revisions r
				USING cutoff
				WHERE r.id IN (
					SELECT id
					FROM image_revisions r2, cutoff
					WHERE r2.name = r.name
					ORDER BY r2.created_at ASC
					OFFSET cutoff.limit_count
				)
			`)
			if err != nil {
				return err
			}

			_, err = db.Exec(ctx, "VACUUM FULL image_revisions")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func LoadAllImages(page int, perPage int) ([]string, error) {
	if page < 1 {
		return nil, errors.New("invalid page")
	}
	if perPage < 1 {
		return nil, errors.New("invalid perPage")
	}
	offset := (page - 1) * perPage

	rows, err := db.Query(context.Background(),
		"SELECT name FROM images ORDER BY name LIMIT $1 OFFSET $2",
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

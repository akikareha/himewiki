package data

import (
	"context"
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

func SaveImage(name string, content []byte) error {
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

	_, err = tx.Exec(ctx,
		`UPDATE state
		 SET image_counter = image_counter + 1
		 WHERE id = 1`,
		)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

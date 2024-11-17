package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/jcserv/portfolio-api/internal/utils"
	"github.com/pkg/errors"

	_ "modernc.org/sqlite"
)

type LibSQL struct {
	db *sql.DB
}

func NewLibSQL(ctx context.Context, dbPath string) (*LibSQL, error) {
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create database directory")
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "failed to ping database")
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	l := &LibSQL{db: db}

	if err := l.CreateEmbeddingsTable(ctx, db); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *LibSQL) CreateEmbeddingsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS embeddings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			embedding_blob BLOB NOT NULL,
			category TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (l *LibSQL) CreateCosineSimilarityFunc(ctx context.Context) error {
	// Create helper function to compute dot product of JSON arrays
	_, err := l.db.ExecContext(ctx, `
		CREATE FUNCTION IF NOT EXISTS json_dot_product(vec1 TEXT, vec2 TEXT) 
		RETURNS REAL AS $$
			WITH 
			a AS (SELECT json_each.value AS x FROM json_each(vec1)),
			b AS (SELECT json_each.value AS y FROM json_each(vec2))
			SELECT COALESCE(SUM(x * y), 0.0)
			FROM a JOIN b ON rowid = rowid;
		$$ LANGUAGE SQL;
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create dot product function")
	}

	// Create helper function to compute vector magnitude of JSON array
	_, err = l.db.ExecContext(ctx, `
		CREATE FUNCTION IF NOT EXISTS json_magnitude(vec TEXT) 
		RETURNS REAL AS $$
			WITH a AS (
				SELECT json_each.value * json_each.value AS squared 
				FROM json_each(vec)
			)
			SELECT SQRT(COALESCE(SUM(squared), 0.0))
			FROM a;
		$$ LANGUAGE SQL;
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create magnitude function")
	}

	// Create the main cosine similarity function
	_, err = l.db.ExecContext(ctx, `
		CREATE FUNCTION IF NOT EXISTS cosine_similarity(vec1 TEXT, vec2 TEXT) 
		RETURNS REAL AS $$
			WITH similarity AS (
				SELECT 
					json_dot_product(vec1, vec2) as dot_prod,
					json_magnitude(vec1) as mag1,
					json_magnitude(vec2) as mag2
			)
			SELECT 
				CASE 
					WHEN mag1 = 0 OR mag2 = 0 THEN 0
					ELSE dot_prod / (mag1 * mag2)
				END
			FROM similarity;
		$$ LANGUAGE SQL;
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create cosine similarity function")
	}

	// Create an optional helper function to convert arrays to JSON if needed
	_, err = l.db.ExecContext(ctx, `
		CREATE FUNCTION IF NOT EXISTS array_to_json(arr TEXT) 
		RETURNS TEXT AS $$
			SELECT CASE 
				WHEN arr LIKE '[%]' THEN arr  -- Already JSON
				ELSE '[' || arr || ']'        -- Convert comma-separated to JSON
			END;
		$$ LANGUAGE SQL;
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create array conversion function")
	}

	return nil
}

func (l *LibSQL) Close() error {
	return l.db.Close()
}

func (l *LibSQL) StoreEmbedding(ctx context.Context, text string, embedding []float32, category string) error {
	byteSlice := utils.Float32SliceToBytes(embedding)

	_, err := l.db.ExecContext(ctx,
		"INSERT INTO embeddings (text, embedding_blob, category) VALUES (?, ?, ?)",
		text, byteSlice, category,
	)
	return errors.Wrap(err, "failed to store embedding")
}

func (l *LibSQL) FindSimilar(ctx context.Context, queryEmbedding []float32, limit int) ([]string, error) {
	queryBytes := utils.Float32SliceToBytes(queryEmbedding)

	rows, err := l.db.QueryContext(ctx, `
		SELECT text, category 
		FROM embeddings 
		ORDER BY cosine_similarity_blob(embedding_blob, ?) DESC 
		LIMIT ?
	`, queryBytes, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query similar embeddings")
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var text, category string
		if err := rows.Scan(&text, &category); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		results = append(results, text)
	}
	return results, nil
}

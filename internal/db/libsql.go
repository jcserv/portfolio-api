package db

import (
	"context"
	"database/sql"
	"math"
	"os"
	"path/filepath"
	"sort"

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
	rows, err := l.db.QueryContext(ctx, `
        SELECT text, category, embedding_blob 
        FROM embeddings
    `)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query embeddings")
	}
	defer rows.Close()

	// Store results with their similarities for sorting
	type Result struct {
		Text       string
		Similarity float64
	}
	var results []Result

	// Calculate similarity for each embedding
	for rows.Next() {
		var text, category string
		var embeddingBlob []byte
		if err := rows.Scan(&text, &category, &embeddingBlob); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		embedding := utils.BytesToFloat32Slice(embeddingBlob)
		similarity := calculateCosineSimilarity(queryEmbedding, embedding)

		results = append(results, Result{
			Text:       text,
			Similarity: similarity,
		})
	}

	// Sort results by similarity (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Take top N results
	var texts []string
	for i := 0; i < limit && i < len(results); i++ {
		texts = append(texts, results[i].Text)
	}

	return texts, nil
}

func calculateCosineSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, magnitude1, magnitude2 float64
	for i := 0; i < len(vec1); i++ {
		v1 := float64(vec1[i])
		v2 := float64(vec2[i])
		dotProduct += v1 * v2
		magnitude1 += v1 * v1
		magnitude2 += v2 * v2
	}

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}

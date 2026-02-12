package main

import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	_ "github.com/marcboeker/go-duckdb/v2"

	"github.com/hybridgroup/yzma/pkg/llama"
)

// Document represents a stored document with its content and embedding vector.
type Document struct {
	ID        string
	Content   string
	Embedding []float32
}

// RAGSystem provides retrieval-augmented generation backed by a llama embedding model and DuckDB.
type RAGSystem struct {
	db           *sql.DB
	model        llama.Model
	vocab        llama.Vocab
	ctx          llama.Context
	embeddingDim int32
}

// NewRAGSystem creates a new RAGSystem by loading the llama model and opening the DuckDB database.
func NewRAGSystem(modelPath, libPath, dbPath string) (*RAGSystem, error) {
	if err := llama.Load(libPath); err != nil {
		return nil, fmt.Errorf("unable to load llama library: %w", err)
	}

	if cfg != nil && !cfg.Verbose {
		llama.LogSet(llama.LogSilent())
	}
	llama.Init()

	model, err := llama.ModelLoadFromFile(modelPath, llama.ModelDefaultParams())
	if err != nil {
		return nil, fmt.Errorf("unable to load model from %s: %w", modelPath, err)
	}
	if model == 0 {
		return nil, fmt.Errorf("failed to load model from %s", modelPath)
	}

	ctxParams := llama.ContextDefaultParams()
	if cfg != nil {
		ctxParams.NCtx = uint32(cfg.ContextSize)
		ctxParams.NBatch = uint32(cfg.BatchSize)
	}
	ctxParams.PoolingType = llama.PoolingTypeMean
	ctxParams.Embeddings = 1

	lctx, err := llama.InitFromModel(model, ctxParams)
	if err != nil {
		llama.ModelFree(model)
		return nil, fmt.Errorf("unable to initialize context: %w", err)
	}

	vocab := llama.ModelGetVocab(model)
	embeddingDim := llama.ModelNEmbd(model)

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		llama.Free(lctx)
		llama.ModelFree(model)
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	rag := &RAGSystem{
		db:           db,
		model:        model,
		vocab:        vocab,
		ctx:          lctx,
		embeddingDim: embeddingDim,
	}

	if err := rag.initDB(); err != nil {
		rag.Close()
		return nil, err
	}

	return rag, nil
}

// initDB creates the documents table in DuckDB if it does not already exist.
func (r *RAGSystem) initDB() error {
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS documents (
			id VARCHAR PRIMARY KEY,
			content VARCHAR,
			embedding FLOAT[%d]
		)
	`, r.embeddingDim)

	_, err := r.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create documents table: %w", err)
	}
	return nil
}

// Close releases all resources held by the RAGSystem, including the database, context, and model.
func (r *RAGSystem) Close() {
	if r.db != nil {
		r.db.Close()
	}
	if r.ctx != 0 {
		llama.Free(r.ctx)
	}
	if r.model != 0 {
		llama.ModelFree(r.model)
	}
	llama.Close()
}

// GenerateEmbedding returns a normalized embedding vector for the given text using the loaded model.
func (r *RAGSystem) GenerateEmbedding(text string) ([]float32, error) {
	tokens := llama.Tokenize(r.vocab, text, true, true)

	batch := llama.BatchGetOne(tokens)
	llama.Decode(r.ctx, batch)

	vec, err := llama.GetEmbeddingsSeq(r.ctx, 0, r.embeddingDim)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	return normalizeVector(vec), nil
}

// normalizeVector returns a unit-length copy of vec using L2 normalization.
func normalizeVector(vec []float32) []float32 {
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	if sum == 0 {
		return vec
	}

	norm := float32(1.0 / math.Sqrt(sum))
	result := make([]float32, len(vec))
	for i, v := range vec {
		result[i] = v * norm
	}
	return result
}

// AddDocument generates an embedding for content and stores the document in the database with the given id.
func (r *RAGSystem) AddDocument(id, content string) error {
	embedding, err := r.GenerateEmbedding(content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	embeddingStr := floatArrayToSQL(embedding)

	_, err = r.db.Exec(`
		INSERT OR REPLACE INTO documents (id, content, embedding)
		VALUES (?, ?, ?::FLOAT[])
	`, id, content, embeddingStr)

	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	return nil
}

// SearchResult holds a document returned by a similarity query along with its cosine similarity score.
type SearchResult struct {
	ID      string
	Content string
	Score   float64
}

// Query returns the topK documents most similar to queryText, ordered by descending cosine similarity.
func (r *RAGSystem) Query(queryText string, topK int) ([]SearchResult, error) {
	queryEmbedding, err := r.GenerateEmbedding(queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	embeddingStr := floatArrayToSQL(queryEmbedding)

	query := fmt.Sprintf(`
		SELECT 
			id, 
			content, 
			array_cosine_similarity(embedding, %s::FLOAT[%d]) AS score
		FROM documents
		WHERE embedding IS NOT NULL
		ORDER BY score DESC
		LIMIT ?
	`, embeddingStr, r.embeddingDim)

	rows, err := r.db.Query(query, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		if err := rows.Scan(&result.ID, &result.Content, &result.Score); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// ListDocuments returns all documents in the database ordered by id.
func (r *RAGSystem) ListDocuments() ([]Document, error) {
	rows, err := r.db.Query(`SELECT id, content FROM documents ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.ID, &doc.Content); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// DeleteDocument removes the document with the given id from the database.
func (r *RAGSystem) DeleteDocument(id string) error {
	result, err := r.db.Exec(`DELETE FROM documents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("document '%s' not found", id)
	}
	return nil
}

// floatArrayToSQL formats a float32 slice as a DuckDB array literal (e.g. "[1.0, 2.0]").
func floatArrayToSQL(arr []float32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range arr {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%f", v))
	}
	sb.WriteString("]")
	return sb.String()
}

// truncate returns s shortened to at most maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

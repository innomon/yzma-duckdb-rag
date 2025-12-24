package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	_ "github.com/marcboeker/go-duckdb/v2"

	"github.com/hybridgroup/yzma/pkg/llama"
)

type Document struct {
	ID        string
	Content   string
	Embedding []float32
}

type RAGSystem struct {
	db         *sql.DB
	model      llama.Model
	vocab      llama.Vocab
	ctx        llama.Context
	embeddingDim int32
}

var (
	modelFile   = flag.String("model", "", "path to embedding model file (GGUF format)")
	libPath     = flag.String("lib", os.Getenv("YZMA_LIB"), "path to llama.cpp library")
	dbPath      = flag.String("db", "rag.db", "path to DuckDB database file (use :memory: for in-memory)")
	contextSize = flag.Int("context", 512, "context size for embeddings")
	batchSize   = flag.Int("batch", 512, "batch size for processing")
	verbose     = flag.Bool("verbose", false, "enable verbose logging")
)

func main() {
	flag.Parse()

	if *modelFile == "" {
		fmt.Println("Usage: ydrag [options] <command> [args]")
		fmt.Println("\nCommands:")
		fmt.Println("  add <id> <content>     Add a document to the knowledge base")
		fmt.Println("  query <text> [top_k]   Query the knowledge base (default top_k=5)")
		fmt.Println("  list                   List all documents")
		fmt.Println("  delete <id>            Delete a document by ID")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf add doc1 \"The capital of France is Paris\"")
		fmt.Println("  ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf query \"What is the capital of France?\"")
		os.Exit(1)
	}

	rag, err := NewRAGSystem(*modelFile, *libPath, *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing RAG system: %v\n", err)
		os.Exit(1)
	}
	defer rag.Close()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Please specify a command: add, query, list, or delete")
		os.Exit(1)
	}

	switch args[0] {
	case "add":
		if len(args) < 3 {
			fmt.Println("Usage: add <id> <content>")
			os.Exit(1)
		}
		content := strings.Join(args[2:], " ")
		if err := rag.AddDocument(args[1], content); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding document: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Document '%s' added successfully\n", args[1])

	case "query":
		if len(args) < 2 {
			fmt.Println("Usage: query <text> [top_k]")
			os.Exit(1)
		}
		topK := 5
		if len(args) >= 3 {
			fmt.Sscanf(args[2], "%d", &topK)
		}
		query := args[1]
		results, err := rag.Query(query, topK)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nTop %d results for: %q\n\n", topK, query)
		for i, r := range results {
			fmt.Printf("%d. [%.4f] %s: %s\n", i+1, r.Score, r.ID, truncate(r.Content, 100))
		}

	case "list":
		docs, err := rag.ListDocuments()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing documents: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Documents in knowledge base (%d total):\n\n", len(docs))
		for _, doc := range docs {
			fmt.Printf("  %s: %s\n", doc.ID, truncate(doc.Content, 80))
		}

	case "delete":
		if len(args) < 2 {
			fmt.Println("Usage: delete <id>")
			os.Exit(1)
		}
		if err := rag.DeleteDocument(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting document: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Document '%s' deleted successfully\n", args[1])

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		os.Exit(1)
	}
}

func NewRAGSystem(modelPath, libPath, dbPath string) (*RAGSystem, error) {
	if err := llama.Load(libPath); err != nil {
		return nil, fmt.Errorf("unable to load llama library: %w", err)
	}

	if !*verbose {
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
	ctxParams.NCtx = uint32(*contextSize)
	ctxParams.NBatch = uint32(*batchSize)
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

type SearchResult struct {
	ID      string
	Content string
	Score   float64
}

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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

package main

import (
	"fmt"
	"strconv"
)

func init() {
	RegisterCommand(&QueryCommand{})
}

// QueryCommand implements the "query" CLI command for searching the knowledge base by similarity.
type QueryCommand struct{}

// Name returns the command name "query".
func (c *QueryCommand) Name() string {
	return "query"
}

// Description returns a short summary of what the query command does.
func (c *QueryCommand) Description() string {
	return "Query the knowledge base for similar documents"
}

// Usage returns the usage string showing expected arguments for the query command.
func (c *QueryCommand) Usage() string {
	return "query <text> [top_k]"
}

// Run executes the query command, searching the RAG system for documents similar
// to the provided text and displaying the top-k results ranked by score.
func (c *QueryCommand) Run(rag *RAGSystem, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.Usage())
	}

	query := args[0]
	topK := 5

	if len(args) >= 2 {
		if k, err := strconv.Atoi(args[1]); err == nil {
			topK = k
		}
	}

	results, err := rag.Query(query, topK)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	fmt.Printf("\nTop %d results for: %q\n\n", topK, query)
	for i, r := range results {
		fmt.Printf("%d. [%.4f] %s: %s\n", i+1, r.Score, r.ID, truncate(r.Content, 100))
	}

	return nil
}

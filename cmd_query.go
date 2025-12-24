package main

import (
	"fmt"
	"strconv"
)

func init() {
	RegisterCommand(&QueryCommand{})
}

type QueryCommand struct{}

func (c *QueryCommand) Name() string {
	return "query"
}

func (c *QueryCommand) Description() string {
	return "Query the knowledge base for similar documents"
}

func (c *QueryCommand) Usage() string {
	return "query <text> [top_k]"
}

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

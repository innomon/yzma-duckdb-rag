package main

import "fmt"

func init() {
	RegisterCommand(&ListCommand{})
}

type ListCommand struct{}

func (c *ListCommand) Name() string {
	return "list"
}

func (c *ListCommand) Description() string {
	return "List all documents in the knowledge base"
}

func (c *ListCommand) Usage() string {
	return "list"
}

func (c *ListCommand) Run(rag *RAGSystem, args []string) error {
	docs, err := rag.ListDocuments()
	if err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}

	fmt.Printf("Documents in knowledge base (%d total):\n\n", len(docs))
	for _, doc := range docs {
		fmt.Printf("  %s: %s\n", doc.ID, truncate(doc.Content, 80))
	}

	return nil
}

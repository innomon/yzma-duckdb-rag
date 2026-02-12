package main

import "fmt"

func init() {
	RegisterCommand(&ListCommand{})
}

// ListCommand implements the "list" CLI command for displaying all documents in the knowledge base.
type ListCommand struct{}

// Name returns the command name "list".
func (c *ListCommand) Name() string {
	return "list"
}

// Description returns a short summary of what the list command does.
func (c *ListCommand) Description() string {
	return "List all documents in the knowledge base"
}

// Usage returns the usage string for the list command.
func (c *ListCommand) Usage() string {
	return "list"
}

// Run executes the list command, printing all documents in the RAG system's
// knowledge base with their IDs and truncated content.
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

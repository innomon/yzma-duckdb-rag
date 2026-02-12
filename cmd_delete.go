package main

import "fmt"

func init() {
	RegisterCommand(&DeleteCommand{})
}

// DeleteCommand implements the "delete" CLI command for removing documents from the knowledge base.
type DeleteCommand struct{}

// Name returns the command name "delete".
func (c *DeleteCommand) Name() string {
	return "delete"
}

// Description returns a short summary of what the delete command does.
func (c *DeleteCommand) Description() string {
	return "Delete a document from the knowledge base"
}

// Usage returns the usage string showing expected arguments for the delete command.
func (c *DeleteCommand) Usage() string {
	return "delete <id>"
}

// Run executes the delete command, removing the document identified by the first
// argument from the RAG system's knowledge base.
func (c *DeleteCommand) Run(rag *RAGSystem, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.Usage())
	}

	id := args[0]

	if err := rag.DeleteDocument(id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	fmt.Printf("Document '%s' deleted successfully\n", id)
	return nil
}

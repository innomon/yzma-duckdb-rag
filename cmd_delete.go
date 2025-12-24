package main

import "fmt"

func init() {
	RegisterCommand(&DeleteCommand{})
}

type DeleteCommand struct{}

func (c *DeleteCommand) Name() string {
	return "delete"
}

func (c *DeleteCommand) Description() string {
	return "Delete a document from the knowledge base"
}

func (c *DeleteCommand) Usage() string {
	return "delete <id>"
}

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

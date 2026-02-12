package main

import (
	"fmt"
	"strings"
)

func init() {
	RegisterCommand(&AddCommand{})
}

// AddCommand implements the "add" CLI command for inserting documents into the knowledge base.
type AddCommand struct{}

// Name returns the command name "add".
func (c *AddCommand) Name() string {
	return "add"
}

// Description returns a short summary of what the add command does.
func (c *AddCommand) Description() string {
	return "Add a document to the knowledge base"
}

// Usage returns the usage string showing expected arguments for the add command.
func (c *AddCommand) Usage() string {
	return "add <id> <content>"
}

// Run executes the add command, parsing the document ID and content from args
// and storing them in the RAG system's knowledge base.
func (c *AddCommand) Run(rag *RAGSystem, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.Usage())
	}

	id := args[0]
	content := strings.Join(args[1:], " ")

	if err := rag.AddDocument(id, content); err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	fmt.Printf("Document '%s' added successfully\n", id)
	return nil
}

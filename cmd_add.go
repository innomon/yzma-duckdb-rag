package main

import (
	"fmt"
	"strings"
)

func init() {
	RegisterCommand(&AddCommand{})
}

type AddCommand struct{}

func (c *AddCommand) Name() string {
	return "add"
}

func (c *AddCommand) Description() string {
	return "Add a document to the knowledge base"
}

func (c *AddCommand) Usage() string {
	return "add <id> <content>"
}

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

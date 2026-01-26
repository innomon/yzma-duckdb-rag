package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	RegisterCommand(&ServeCommand{})
}

type ServeCommand struct{}

func (c *ServeCommand) Name() string {
	return "serve"
}

func (c *ServeCommand) Description() string {
	return "Start the MCP server (stdio transport)"
}

func (c *ServeCommand) Usage() string {
	return "serve"
}

func (c *ServeCommand) Run(rag *RAGSystem, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	server := NewMCPServer(rag)

	if *verbose {
		fmt.Fprintln(os.Stderr, "Starting MCP server on stdio...")
	}

	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}

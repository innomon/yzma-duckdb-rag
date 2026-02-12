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
	return "Start the MCP server (supports stdio, sse, streamable-http transports)"
}

func (c *ServeCommand) Usage() string {
	return "serve [--transport stdio|sse|streamable-http] [--port PORT]"
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

	transport := cfg.Server.Transport
	addr := ":" + cfg.Server.Port

	if *verbose {
		fmt.Fprintf(os.Stderr, "Starting MCP server (transport=%s)...\n", transport)
	}

	if err := server.Run(ctx, transport, addr); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}
